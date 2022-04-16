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
	"sim/pkg/logging"
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
	TargetFLAGShouldReBalance            // reBalance
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

func (t *target) Add(cli Client) {
	if cli == nil {
		return
	}
	t.add(cli)
}

func (t *target) Del(token []string) {
	if token == nil {
		return
	}

}

func (t *target) BroadCast(data []byte) {
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*Group)
		g.BroadCast(data)
		node = node.Next()
	}
}

func (t *target) BroadCastWithInnerJoinTag(data []byte, otherTag []string) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		node.Value.(*Group).BroadCastWithOtherTag(data, otherTag)
		node = node.Next()
	}
}

func (t *target) Expansion() {
	t.expansion()
}

func (t *target) Status() targetFlag {
	return t.flag
}

func (t *target) ReBalance() {
	t.reBalance()
}

// ======================================== helper =====================================

func (t *target) add(cli Client) {
	t.rw.Lock()
	defer t.rw.Unlock()
	g := t.offset.Value.(*Group)
	if same := g.Add(cli); same {
		return
	}
	t.num++
	t.moveOffset()
	if t.num > t.limit*t.numG && t.flag == TargetFlagNORMAL {
		t.flag = TargetFLAGShouldEXTENSION
	}
	return
}

func (t *target) del(token []string) {
	t.rw.Lock()
	defer t.rw.Unlock()
	var res []string
	node := t.li.Front()
	for node != nil {
		gp := node.Value.(*Group)
		stop, result := gp.Del(token)
		res = append(res, result...)
		t.num -= len(result)
		if stop {
			break
		}
		node = node.Next()
	}
}

func (t *target) moveOffset() {
	if t.offset.Next() != nil {
		t.offset = t.offset.Next()
	} else {
		t.offset = t.li.Front()
	}
}

func (t *target) Init(name string) *target {
	t.li = list.New()
	g := NewGroup(t.limit)
	elm := t.li.PushFront(g)
	t.offset = elm
	t.name = name
	t.numG++
	return t
}

func (t *target) setFlag(flag targetFlag) {
	logging.Infof("sim/wti : change target level %s ,target name is %s", flag, t.name)
	t.flag = flag
}

func (t *target) expansion() {
	t.rw.Lock()
	defer t.rw.Unlock()
	t.setFlag(TargetFLAGEXTENSION)
	newG := NewGroup(t.limit)
	t.li.PushBack(newG)
	t.numG += 1
	t.setFlag(TargetFlagNORMAL)
}

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
			clis, _ := nG.Free()
			temCli = append(temCli, clis...)
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
			ng.BatchAdd(temCli[:])
		} else {
			ng.BatchAdd(temCli[:avg])
			temCli = temCli[avg:]
		}
	}
}

func (t *target) reBalance() {
	// 根据当前节点进行平均每个节点的人数
	avg := t.num / t.numG

	// 负载小于等于 10 的节点都属于 紧急节点
	// 负载小于 0 的节点属于立刻节点
	// 获取到所有节点的负载情况 ： 负载小于 0 的优先移除
	var lowLoadG []*Group
	var steals []Client
	t.rw.Lock()
	defer t.rw.Unlock()
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*Group)
		if g.Num() > avg {
			// num >avg ,说明超载
			steal := g.Num() - avg
			st, _ := g.Move(steal)
			steals = append(steals, st...)
		} else {
			//  进入这里说明当前节点load 偏低
			diff := avg - g.Num()
			if diff > 0 {
				if len(steals) > diff {
					g.BatchAdd(steals[:diff])
					steals = steals[diff:]
					node =node.Next()
					continue
				} else {
					g.BatchAdd(steals)
					steals = []Client{}
					node = node.Next()
					continue
				}
			}
			// 到这里说明情况当前这个
			lowLoadG = append(lowLoadG, g)
		}
		node = node.Next()
	}

	for _, g := range lowLoadG {
		w := avg - g.Num()
		if w >= len(steals) {
			g.BatchAdd(steals)
			break
		}
		g.BatchAdd(steals[:w])
		steals = steals[w:]
	}

}
