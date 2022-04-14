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
	"sync"
	"time"
)

type Group struct {
	tag        *target
	rw         *sync.RWMutex
	cap, num   int
	set        map[string]Client
	createTime int64
}

func NewGroup(cap int) *Group {
	return &Group{
		rw:         &sync.RWMutex{},
		set:        map[string]Client{},
		cap:        cap,
		createTime: time.Now().Unix(),
	}
}

// ================================ action =============================

//add 添加cli
func (g *Group) add(cli Client) (overCap bool) {
	g.rw.Lock()
	defer g.rw.Unlock()
	g.set[cli.Token()] = cli
	g.num ++
	if g.num > g.cap {
		return true
	}
	return false
}

//del 先删除group内的内容，然后删除用户内的内容
func (g *Group) del(tokens ...string) bool {
	g.rw.Lock()
	defer g.rw.Unlock()
	var flag = true
	for _, token := range tokens {
		if _, ok := g.set[token]; ok {
			delete(g.set, token)
		} else {
			flag = false
		}
	}
	return flag
}

//exist 是否存在cli
func (g *Group) exist(token string) bool {
	g.rw.RLock()
	defer g.rw.RUnlock()
	if _, ok := g.set[token]; ok {
		return true
	}
	return false
}

//counter 是否存在cli
func (g *Group) counter() int {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return len(g.set)
}

// remove 移除用户的token
func (g *Group) remove(num int) []Client {
	g.rw.Lock()
	defer g.rw.Unlock()
	var (
		counter int = 0
		res     []Client
	)

	for k, v := range g.set {
		if counter == num {
			break
		}
		delete(g.set, k)
		res = append(res, v)
		counter++
	}
	return res
}

// batchAdd 批量新增用户
func (g *Group) batchAdd(cliS []Client) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, cli := range cliS {
		g.set[cli.Token()] = cli
	}
}

// free 释放group的用户
func (g *Group) free() []Client {
	g.rw.Lock()
	defer g.rw.RLock()
	var res []Client
	for _, v := range g.set {
		res = append(res, v)
	}
	return res
}

//broadCast  给所有用户广播
func (g *Group) broadCast(content []byte) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	for _, v := range g.set {
		v.Send(content)
	}
}

//broadCastWithOtherTag 如果用户存在对应的标签可以将内容发送给对应的用户
func (g *Group) broadCastWithOtherTag(content []byte, otherTags []string) error {
	if len(otherTags) == 0 {
		return errors.ErrCliISNil
	}
	g.rw.RLock()
	defer g.rw.RUnlock()
	for _, v := range g.set {
		if v.HaveTag(otherTags) {
			v.Send(content)
		}
	}
	return nil
}

func (g *Group) Update(tokens ...string) {
	g.del(tokens...)
}
