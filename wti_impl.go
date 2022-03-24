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
	rw *sync.RWMutex
	set  map[string]client.Clienter
	createTime int64
}

func NewGroup()*Group {
	return &Group{
		rw:  &sync.RWMutex{},
		set: map[string]client.Clienter{},
		createTime: time.Now().Unix(),
	}
}




func (g *Group) CreateTime ()int64{
	// should add mutex ,but maybe not
	return g.createTime
}

// 给所有用户广播
func (g *Group) broadCast(content []byte){
	g.rw.RLock()
	defer g.rw.RUnlock()
	for _,v := range g.set {
		v.Send(content)
	}
}

// 添加cli
func (g *Group) addCli(clis ...client.Clienter){
	g.rw.Lock()
	defer g.rw.Unlock()
	for _,v := range clis{
		g.set[v.Token()]=v
	}
}

// 删除cli
func (g *Group) rmCli(tokens ... string){
	g.rw.Lock()
	defer g.rw.Unlock()
	for _,token := range tokens {
		delete(g.set,token)
	}
}

// 是否存在cli
func (g *Group) isExsit(token string)bool{
	g.rw.RLock()
	defer g.rw.RUnlock()
	if _,ok:=g.set[token];ok{
		return true
	}
	return false
}

// 是否存在cli
func (g *Group) Counter()int64{
	g.rw.RLock()
	defer g.rw.RUnlock()
	return int64(len(g.set))
}


// 就是使用这个方法将g 注册到Observer 上面去。
func (g *Group) Update(tokens ... string){
	g.rmCli(tokens...)
}