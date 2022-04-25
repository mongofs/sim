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
	"strconv"
	"sync"
	"time"
)

const DefaultCapacity = 128

type gStatus int

const (
	gStatusNormal gStatus = iota + 1
	gStatusClosed
)

type group struct {
	rw             *sync.RWMutex
	flag           gStatus
	cap, num, load int
	set            map[string]Client
	createTime     int64
}

var groupPool = sync.Pool{
	New: func() interface{} {
		return &group{
			rw:  &sync.RWMutex{},
			set: make(map[string]Client, DefaultCapacity),
		}
	},
}

func getG(cap int) *group {
	if cap == 0 {
		cap = DefaultCapacity
	}
	grp := groupPool.Get().(*group)
	grp.createTime = time.Now().Unix()
	grp.cap = cap
	return grp
}

func (g *group) info() *map[string]string {
	g.rw.RLock()
	defer g.rw.RUnlock()
	res := &map[string]string{
		"status":      strconv.Itoa(int(g.flag)),
		"online":      strconv.Itoa(g.num),
		"load":        strconv.Itoa(g.load),
		"create_time": strconv.Itoa(int(g.createTime)),
	}
	return res
}


func (g *group) free() ([]Client, error) {
	return g.move(g.num), nil
}

func (g *group) add(cli Client) bool {
	g.rw.Lock()
	defer g.rw.Unlock()
	if _, ok := g.set[cli.Identification()]; ok {
		g.set[cli.Identification()] = cli
		return true
	} else {
		g.set[cli.Identification()] = cli
		g.num++
		g.calculateLoad()
	}

	if g.num > g.cap {
		return false
	}
	return false
}

func (g *group) addMany(cliS []Client) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, cli := range cliS {
		if _, ok := g.set[cli.Identification()]; ok {
			g.set[cli.Identification()] = cli
		} else {
			g.set[cli.Identification()] = cli
			g.num++
		}
	}
	g.calculateLoad()
}

func (g *group) del(tokens []string) (clear bool, success []string, current int) {
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

func (g *group) move(num int) []Client {
	var (
		counter  = 0
		res     []Client
	)
	g.rw.Lock()
	defer g.rw.Unlock()
	if num > g.num {num= g.num}
	if num == g.num {
		g.flag = gStatusClosed
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

func (g *group) broadcast(content []byte) []string {
	g.rw.RLock()
	defer g.rw.RUnlock()
	var res []string
	for _, v := range g.set {
		err := v.Send(content)
		if err != nil {
			res = append(res, v.Identification())
		}
	}
	return res
}

func (g *group) broadcastWithTag(content []byte, tags []string) []string {
	var res []string
	g.rw.RLock()
	defer g.rw.RUnlock()
	for _, v := range g.set {
		if v.HaveTags(tags) {
			err := v.Send(content)
			if err != nil {
				res = append(res, v.Identification())
			}
		}
	}
	return res
}

func (g *group) calculateLoad() {
	g.load = g.cap - g.num // cap - len
}

func (g *group) destroy() error {
	g.cap, g.num, g.cap, g.load, g.createTime = 0, 0, 0, 0, 0
	g.flag = gStatusNormal
	groupPool.Put(g)
	return nil
}
