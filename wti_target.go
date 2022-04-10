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
	li     *list.List
	num    *int  // online user
	limit  *int  // max online user for group
	offset uint8 // the next user group offset
}

func newTarget(targetName string) *target {
	res := &target{rw: sync.RWMutex{}}
	res.Init(targetName)
	return res
}

// Init 实例化list
func (t *target) Init(name string) *target {
	t.li = list.New()
	t.name = name
	t.li.PushFront(NewGroup())
	return t
}

// Add 添加用户到target中
func (t *target) Add(cli Client) {
	offset := t.offset
	var node = t.li.Front()
	for offset != 0 {
		if node.Next() != nil {
			node = node.Next()
		} else {
			continue
		}
		offset -= 1
	}
	g := node.Value.(*Group)
	if err := g.add(cli); err != nil {
		t.extension(cli)
	}
}

// Del 将用户从target剔除
func (t *target) Del(token string) {
	// node.del(token)

}

func (t *target) Num() int {
	return t.num
}

func (t *target) CreateTime() int64 {
	return t.num
}

// BroadCast 给同组内的用户进行广播
func (t *target) BroadCast(data []byte) {}

// BroadCastAndWithTag 给同组内伴随另外的组的用户进行广播，这个就是求交集，
// 用户有tag1 标签，并且有tag2 标签，进行求交集广播
func (t *target) BroadCastWithInnerJoinTag(data []byte, otherTag []string) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		node.Value.(*Group).broadCast(data)
		if node.Next() != nil {
			node = node.Next()
			continue
		}
		break
	}

}

// extension 创建一个新的group
func (t *target) extension(cli Client) {
	newG := NewGroup()
	_ = newG.add(cli)

	go func() {
		var node := t.li.Front()
		for node != nil {

		}
	}()
}

// free 释放掉当前tag
func (t *target) free() error {
	newG := NewGroup()
	_ = newG.add(cli)

	go func() {
		var node := t.li.Front()
		for node != nil {

		}
	}()
}
