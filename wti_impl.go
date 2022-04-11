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
	"sim/pkg/logging"
	"sync"
	"time"
)

type set struct {
	// mp tagName =>
	mp    map[string]*target // wti => []string
	rw    *sync.RWMutex
	async chan BCData
}

func newSet() *set {
	return &set{
		mp:    map[string]*target{},
		rw:    &sync.RWMutex{},
		async: make(chan BCData, 100),
	}
}

// ============================ for client - side =====================

//Find 查找用户所需的target
func (t *set) Find(tags []string) ([]*target, error) {
	if len(tags) == 0 {
		return nil, ERRNotSupportWTI
	}
	t.rw.Lock()
	defer t.rw.Unlock()
	var result []*target
	for _, tag := range tags {
		if target, ok := t.mp[tag]; !ok { // target not exist
			t.mp[tag] = newTarget(tag)
			result = append(result, target)
		} else { // wti exist
			result = append(result, target)
		}
	}
	return result, nil
}

//Del 删除target
func (t *set) Del(tags []string) {
	if len(tags) == 0 {
		return
	}
	t.rw.Lock()
	defer t.rw.Unlock()
	for _, tag := range tags {
		if traget, ok := t.mp[tag]; ok {
			traget.free()
			delete(t.mp, tag)
		}
	}
}

// ============================ for RPC server -side ====================

type PushSetting uint8

const (
	PushSettingUnion = iota + 1 // 并集
	PushSettingInner            // 交集
)

type BCData struct {
	data        *[]byte
	target      []string
	pushSetting PushSetting
}

//BroadCast 给某一个标签的群体进行广播
func (t *set) BroadCast(data *BCData) {
	if data == nil {
		return
	}
	t.rw.RLock()
	defer t.rw.RUnlock()
	switch data.pushSetting {
	case PushSettingUnion:
		for _, tag := range data.target {
			if target, ok := t.mp[tag]; ok {
				target.BroadCast(*data.data)
				continue
			}
			logging.Warnf("target  %v is not exist", tag)
		}
	case PushSettingInner:
		var (
			min      = 10 >> 2
			minIndex *target
		)
		for _, tag := range data.target {
			if target, ok := t.mp[tag]; ok {
				if min > target.Num() {
					min = target.Num()
					minIndex = target
				}
			}
			minIndex.BroadCastWithInnerJoinTag(*data.data, data.target)
		}
	}
}

//BroadCastGroupByTarget 调用这个就可以分类广播，可能出现不同的target 需要不同的内容,这种
//用法和循环调用broadCast 效果一样
func (t *set) BroadCastGroupByTarget(targetAndContent map[string][]byte) {
	if len(targetAndContent) == 0 {
		return
	}
	for k, v := range targetAndContent {
		data := BCData{
			data:        &v,
			target:      []string{k},
			pushSetting: PushSettingInner,
		}
		t.async <- data
	}
}

//BroadCastToInnerJoinTarget 调用 交集调用
func (t *set) BroadCastToInnerJoinTarget(content []byte, tag []string) {
	if len(tag) == 0 {
		return
	}
	data := BCData{
		data:        &content,
		target:      tag,
		pushSetting: PushSettingInner,
	}
	t.async <- data
}

//BroadCastToUnionJoinTarget 调用 交集调用
func (t *set) BroadCastToUnionJoinTarget(content []byte, tag []string) {
	if len(tag) == 0 {
		return
	}
	data := BCData{
		data:        &content,
		target:      tag,
		pushSetting: PushSettingUnion,
	}
	t.async <- data
}

type DistributeParam struct {
	TagName string
	Onlines int64
	CreateTime int64
}

//Distribute  获取到tagOnline 在线用户人数,以及对应的群组的分布情况
func (t *set) Distribute(tags ...string) map[string]*DistributeParam {
	var res = map[string]*DistributeParam{}
	if len(tags) == 0 {
		t.rw.RLock()
		for k, v := range t.mp {
			tem := &DistributeParam{
				TagName:    k,
				Onlines:    int64(v.Num()),
				CreateTime: v.CreateTime(),
			}
			res[k] = tem
		}
		t.rw.RUnlock()
		return res
	}
	t.rw.RLock()
	for _, tag := range tags {
		// get the tag
		if v, ok := t.mp[tag]; ok {
			tem := &DistributeParam{
				TagName:    tag,
				Onlines:    int64(v.Num()),
				CreateTime: v.CreateTime(),
			}
			res[tag] = tem
		}
	}
	t.rw.RUnlock()
	return res
}

//monitor is a goroutine to monitor the wti run state
func (t *set) monitor() {
	for {
		for _, v := range t.mp {
			if v.Num() == 0 && time.Now().Unix()-v.CreateTime() > 60 {
				err := v.free()
				if err != nil {
					logging.Error(err)
					continue
				}
				delete(t.mp, v.name)
			}
		}

		time.Sleep(20 * time.Second)
	}
}
