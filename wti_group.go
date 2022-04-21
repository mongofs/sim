/*
 * Copyright 2022 steven
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *    http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sim

import (
	"sim/pkg/errors"
	"strconv"
	"sync"
	"time"
)

const DefaultCapacity = 128

type GroupStatus int

const (
	GroupStatusNormal GroupStatus = iota + 1
	GroupStatusClosed
)

type Group struct {
	rw             *sync.RWMutex
	flag           GroupStatus
	cap, num, load int
	set            map[string]Client
	createTime     int64
}

func NewGroup(cap int) *Group {
	if cap == 0 {
		cap = DefaultCapacity
	}
	grp := groupPool.Get().(*Group)
	grp.createTime = time.Now().Unix()
	grp.cap = cap
	return grp
}

// ================================ action =============================

func (g *Group) Info() *map[string]string {
	g.rw.RLock()
	defer g.rw.RUnlock()
	res := &map[string]string{
		"online":      strconv.Itoa(g.num),
		"load":        strconv.Itoa(g.load),
		"create_time": strconv.Itoa(int(g.createTime)),
	}
	return res
}

func (g *Group) Add(cli Client) (same bool) {
	if cli == nil {
		return true
	}
	return g.add(cli)
}

func (g *Group) Destroy() error {
	if g.num != 0 {
		return errors.ERRWTIGroupNotClear
	}
	return g.destroy()
}

func (g *Group) Num() int {
	return g.num
}

func (g *Group) Del(tokens []string) (stop bool, success []string, current int) {
	if len(tokens) == 0 {
		return true, nil, 0
	}
	return g.del(tokens)
}

func (g *Group) Move(num int) ([]Client, error) {
	if num <= 0 || num >= g.num {
		return nil, errors.ErrGroupBadParam
	}
	return g.move(num), nil
}

func (g *Group) Free() ([]Client, error) {
	return g.move(g.num), nil
}

func (g *Group) BatchAdd(cliS []Client) {
	if len(cliS) == 0 {
		return
	}
	g.addMany(cliS)
}

func (g *Group) BroadCast(content []byte) []string {
	if content == nil {
		return nil
	}
	return g.broadcast(content)
}

func (g *Group) BroadCastWithOtherTag(content []byte, otherTags []string) ([]string, error) {
	if len(otherTags) == 0 {
		return nil, errors.ErrCliISNil
	}
	return g.broadcastWithTag(content, otherTags), nil
}

// ------------------------------------------ private ------------------------------------------

var groupPool = sync.Pool{
	New: func() interface{} {
		return &Group{
			rw:  &sync.RWMutex{},
			set: make(map[string]Client, DefaultCapacity),
		}
	},
}

func (g *Group) add(cli Client) bool {
	g.rw.Lock()
	defer g.rw.Unlock()
	if _, ok := g.set[cli.Token()]; ok {
		g.set[cli.Token()] = cli
		return true
	} else {
		g.set[cli.Token()] = cli
		g.num++
		g.calculateLoad()
	}

	if g.num > g.cap {
		return false
	}
	return false
}

func (g *Group) addMany(cliS []Client) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, cli := range cliS {
		if _, ok := g.set[cli.Token()]; ok {
			g.set[cli.Token()] = cli
		} else {
			g.set[cli.Token()] = cli
			g.num++
		}
	}
	g.calculateLoad()
}

func (g *Group) del(tokens []string) (clear bool, success []string, current int) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, token := range tokens {
		if _, ok := g.set[token]; ok {
			delete(g.set, token)
			g.num--
			g.calculateLoad()
			success = append(success, token)
		} else {
			clear = false
		}
	}
	return clear, success, g.num
}

func (g *Group) move(num int) []Client {
	var (
		counter int = 0
		res     []Client
	)
	g.rw.Lock()
	defer g.rw.Unlock()
	if num == g.num {
		g.flag = GroupStatusClosed
	}
	for k, v := range g.set {
		if counter == num {
			break
		}
		delete(g.set, k)
		res = append(res, v)
		g.num--
		counter++
	}
	g.calculateLoad()
	return res
}

func (g *Group) broadcast(content []byte) []string {
	g.rw.RLock()
	defer g.rw.RUnlock()
	var res []string
	for _, v := range g.set {
		err := v.Send(content)
		if err != nil {
			res = append(res, v.Token())
		}
	}
	return res
}

func (g *Group) broadcastWithTag(content []byte, tags []string) []string {
	var res []string
	g.rw.RLock()
	defer g.rw.RUnlock()
	for _, v := range g.set {
		if v.HaveTag(tags) {
			err := v.Send(content)
			if err != nil {
				res = append(res, v.Token())
			}
		}
	}
	return res
}

func (g *Group) calculateLoad() {
	g.load = g.cap - g.num // cap - len
}

func (g *Group) destroy() error {
	g.cap, g.num, g.cap, g.load, g.createTime = 0, 0, 0, 0, 0
	g.flag = GroupStatusNormal
	groupPool.Put(g)
	return nil
}
