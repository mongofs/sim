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

type tg struct {
	mp map[string] *Group // wti => []string
	rw *sync.RWMutex
}


func newwti() WTI {
	return &tg{
		mp: map[string]*Group{},
		rw: &sync.RWMutex{},
	}
}


// 给用户设置标签
func (t *tg)  SetTAG(cli Client, tags ...string) {
	if len(tags)== 0 {
		return
	}
	t.rw.Lock()
	defer t.rw.Unlock()
	for _,tag := range tags{
		if group,ok:= t.mp[tag];!ok { // wti not exist
			t.mp[tag]= NewGroup()
			t.mp[tag].addCli(cli)
		}else { // wti exist
			group.addCli(cli)
		}
	}
}


// 删除用户的标签
func (t *tg) DelTAG(cli Client, tags ...string){
	if len(tags) == 0 {return }
	t.rw.Lock()
	defer t.rw.Unlock()
	for _,tag := range tags {
		if group,ok := t.mp[tag];!ok{
			continue
		}else {
			group.rmCli(cli.Token())
		}
	}
}


// 给某一个标签的群体进行广播
func (t *tg) BroadCast(content []byte,tags ...string) {
	if len(tags)== 0 {
		return
	}
	t.rw.RLock()
	defer t.rw.RUnlock()

	for _,tag := range tags{
		if group,ok := t.mp[tag];ok{
			group.broadCast(content)
		}
	}
}


// 通知所有组进行自查某个用户，并删除
func (t *tg)Update(token ...string){
	t.rw.RLock()
	defer t.rw.RUnlock()
	for _,v := range t.mp {
		v.Update(token... )
	}
}


// 调用这个就可以分类广播，可能出现不同的targ 需要不同的内容
func(t *tg)BroadCastByTarget(targetAndContent map[string][]byte){
	if len(targetAndContent) == 0{ return }
	for target ,content := range targetAndContent {
		go t.BroadCast(content,target)
	}
}


// 获取到用户token的所有TAG，时间复杂度是O(m) ,m是所有的房间
func (t *tg)GetClientTAGs(token string)[]string{
	var res []string
	t.rw.RLock()
	defer t.rw.RUnlock()
	for k,v:= range t.mp{
		exsit:= v.isExsit(token)
		if exsit {
			res = append(res,k )
		}
	}
	return res
}

// 如果创建时间为0 ，表示没有这个房间
func (t *tg) GetTAGCreateTime(tag string) int64{
	t.rw.RLock()
	defer t.rw.RUnlock()
	if v,ok:=t.mp[tag];ok{
		return v.createTime
	}
	return 0
}


// 删除tag ,这个调用一个大锁将全局锁住清空过去的内容
func (t *tg) FlushWTI() {
	t.rw.Lock()
	defer t.rw.Unlock()
	for k,v := range t.mp{
		if v.Counter() ==0 && time.Now().Unix() -v.CreateTime() >60 {
			delete(t.mp,k)
		}
	}
}


// 获取到tagOnlines 在线用户人数
func (t *tg) Distribute(tags...  string) map[string]*DistributeParam {
	var res = map[string]*DistributeParam{}
	if len(tags) == 0 {
		t.rw.RLock()
		for k,v:= range t.mp {
			tem := &DistributeParam{
				TagName:    k,
				Onlines:    v.Counter(),
				CreateTime: v.createTime,
			}
			res[k]=tem
		}
		t.rw.RUnlock()
		return res
	}
	t.rw.RLock()
	for _,tag := range tags {
		// get the tag
		if v,ok := t.mp[tag];ok {
			tem := &DistributeParam{
				TagName:    tag,
				Onlines:    v.Counter(),
				CreateTime: v.createTime,
			}
			res[tag]= tem
		}
	}
	t.rw.RUnlock()
	return res
}