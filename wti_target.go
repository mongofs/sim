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
	"container/list"
	"errors"
	"sync"
	"time"
)

type targetFlag int

const (
	TargetFlagNORMAL          = iota + 1 // normal
	TargetFLAGShouldEXTENSION            // start extension
	TargetFLAGEXTENSION                  // extension
	TargetFLAGShouldSHRINKS              // start shrinks
	TargetFLAGSHRINKS                    // shrinks
)

// target 存放的内容是 target -> groupA->GroupB
type target struct {
	rw     sync.RWMutex
	name   string
	num    int           // online user
	numG   int           // online Group
	offset *list.Element // the next user group offset

	flag       targetFlag //
	li         *list.List
	limit      int   // max online user for group
	createTime int64 // create time
}

func NewTarget(targetName string, limit int) (*target, error) {
	if targetName == "" || limit == 0 {
		return nil, errors.New("bad param of target")
	}
	res := &target{
		name:       targetName,
		rw:         sync.RWMutex{},
		flag:       TargetFlagNORMAL,
		limit:      limit,
		createTime: time.Now().Unix(),
	}
	res.Init(targetName)
	return res, nil
}

// ============================================= API =================================

// Add 添加用户到target中， 每次进行用户添加的时候需要判断用户是否已经在线，如果用户在线的话需要
// 将用户的连接覆盖，所有需要遍历所有group进行，将旧链接删除，放入新连接。但是在这里不考虑这种处理
// 方式，如果出现重复连接，将会出现标识同时出现，前端需要避免重复连接的情况
func (t *target) Add(cli Client) {
	t.rw.Lock()
	defer t.rw.Unlock()

	g := t.offset.Value.(*Group)
	if overCap := g.add(cli); overCap {
		t.setFlag(TargetFLAGShouldEXTENSION)
		return
	}

	t.num++
	return
}

// Del 将用户从target剔除
func (t *target) Del(token string) {
	node := t.li.Front()
	for node != nil {
		gp := node.Value.(*Group)
		if gp.del(token) {
			t.num--
			break
		}
		node = node.Next()
	}
}

// BroadCast 给同组内的用户进行广播
func (t *target) BroadCast(data []byte) {
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*Group)
		g.broadCast(data)
		node = node.Next()
	}
}

// BroadCastWithInnerJoinTag  进行交集广播
func (t *target) BroadCastWithInnerJoinTag(data []byte, otherTag []string) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		node.Value.(*Group).broadCastWithOtherTag(data, otherTag)
		node = node.Next()
	}
}

func (t *target) Status() targetFlag {
	return t.flag
}

// ======================================== helper =====================================

// Init 实例化list
func (t *target) Init(name string) *target {
	t.li = list.New()
	elem := &list.Element{Value: NewGroup(t.limit)}
	t.li.PushFront(elem)
	t.numG++
	return t
}

// setFlag 所有的setFlag 都应该在加锁情况操作，
func (t *target) setFlag(flag targetFlag) {
	t.flag = flag
}

// expansion 促发扩容
func (t *target) expansion() {
	t.rw.Lock()
	defer t.rw.RUnlock()
	t.setFlag(TargetFLAGEXTENSION)
	newG := NewGroup(t.limit)
	t.li.PushBack(list.Element{Value: newG})
	t.reBalance()
}

//shrinks 缩容
func (t *target) shrinks() {
	// 缩容的标准是 ： 缩容标志
	// 缩容的G的个数 ： num / limit +2
	// 缩容的步骤： 从尾部释放，释放到具体个数
	t.setFlag(TargetFLAGSHRINKS)
	targetG := t.num/t.limit + 2
	numG := 0
	var temCli []Client
	for node := t.li.Front(); node != nil; node = node.Next() {
		numG++
		if numG > targetG {
			// 释放group 用户
			nG := node.Value.(*Group)
			temCli = append(temCli, nG.free()...)
			// 将前置node置为空
			prevNode := node.Prev()
			prevNode.Next().Value = nil
		}
	}

	dist := len(temCli)
	avg := dist / targetG

	for node := t.li.Front(); node != nil; node = node.Next() {
		ng := node.Value.(*Group)
		if node.Next() == nil {
			ng.batchAdd(temCli[:])
		} else {
			ng.batchAdd(temCli[:avg])
			temCli = temCli[avg:]
		}
	}
}

//reBalance 将所有的用户进行重新分配一下
func (t *target) reBalance() {
	avg := t.num / t.numG

	node := t.li.Front()
	var temCli []Client
	for node != nil {
		g := node.Value.(*Group)
		diff := g.counter() - avg
		if diff > 0 {
			temCli = append(temCli, g.remove(diff)...)
		}
		node = node.Next()
	}

	for node := t.li.Front(); node != nil; node = node.Next() {
		g := node.Value.(*Group)
		diff := g.counter() - avg
		if diff < 0 {
			diff = -diff
			distr := temCli[0:diff]
			temCli = temCli[diff:]
			g.batchAdd(distr)
		}
	}
}
