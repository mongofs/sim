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
	"sync"
	"time"
)

type Group struct {
	rw         *sync.RWMutex
	set        map[string]Client
	createTime int64
}

func NewGroup() *Group {
	return &Group{
		rw:         &sync.RWMutex{},
		set:        map[string]Client{},
		createTime: time.Now().Unix(),
	}
}


// 给所有用户广播
func (g *Group) broadCast(content []byte) {
	g.rw.RLock()
	defer g.rw.RUnlock()
	for _, v := range g.set {
		v.Send(content)
	}
}

//add 添加cli
func (g *Group) add(clis ...Client) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, v := range clis {
		g.set[v.Token()] = v
	}
}

//del 删除cli
func (g *Group) del(tokens ...string) {
	g.rw.Lock()
	defer g.rw.Unlock()
	for _, token := range tokens {
		delete(g.set, token)
	}
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
func (g *Group) counter() int64 {
	g.rw.RLock()
	defer g.rw.RUnlock()
	return int64(len(g.set))
}

func (g *Group) Update(tokens ...string) {
	g.del(tokens...)
}
