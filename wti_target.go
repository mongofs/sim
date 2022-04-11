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
	"sync"
)

// target 存放的内容是 target -> groupA->GroupB
type target struct {
	rw     sync.RWMutex

	name   string
	num    int // online user
	numG   int
	offset int // the next user group offset

	flag  int // 0 不动，1是扩容，2是缩容
	li    *list.List
	limit int // max online user for group
	createTime int64 // create time
}

func newTarget(targetName string) *target {
	res := &target{rw: sync.RWMutex{}}
	res.Init(targetName)
	return res
}

// ============================================= API =================================

// Init 实例化list
func (t *target) Init(name string) *target {
	t.li = list.New()
	t.name = name
	elem := &list.Element{Value: NewGroup(t.limit)}
	t.li.PushFront(elem)
	return t
}

// Add 添加用户到target中
func (t *target) add(cli Client) {
	offset := t.offset
	node := t.li.Front()
	for offset != 0 {
		if node.Next() != nil {
			node = node.Next()
		} else {
			g := node.Value.(*Group)
			if err := g.add(cli); err != nil {
				t.expansion(cli)
			}
		}
		offset -= 1
	}
	t.num++
}

// Del 将用户从target剔除
func (t *target) del(token string) {
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
func (t *target) broadCast(data []byte) {
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*Group)
		g.broadCast(data)
		node = node.Next()
	}
}

// broadCastWithInnerJoinTag  进行交集广播
func (t *target) broadCastWithInnerJoinTag(data []byte, otherTag []string) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		node.Value.(*Group).broadCastWithOtherTag(data, otherTag)
		node = node.Next()
	}
}

// expansion 促发扩容
func (t *target) expansion(cli Client) {
	newG := NewGroup(t.limit)
	_ = newG.add(cli)
	go t.handleExpansion(newG)
}

//shrinks 缩容
func (t *target) shrinks() {
	// 缩容的标准是 ： 缩容标志
	// 缩容的G的个数 ： num / limit +2
	// 缩容的步骤： 从尾部释放，释放到具体个数

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

// handleExpansion 具体的扩容逻辑
func (t *target) handleExpansion(newg *Group) {
	defer func() {
		if err := recover(); err != nil {
			// logging.Error(err)
		}
	}()
	t.li.PushBack(list.Element{Value: newg})
	t.reBalance()
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

// free 释放掉当前target
func (t *target) free() error {
	newG := NewGroup()
	_ = newG.add(cli)

	go func() {
		var node := t.li.Front()
		for node != nil {

		}
	}()
}
