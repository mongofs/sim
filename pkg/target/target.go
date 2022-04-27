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

package target

import (
	"container/list"
	"errors"
	"sim/pkg/logging"
	"sync"
	"time"
)

// target 是相同的标签的管理单元，相同的target都会放置到相同的
type target struct {
	rw             sync.RWMutex
	name           string
	num            int           // online user
	numG           int           // online Group
	targetG        int           // should be numG
	maxGOnlineDiff int           // 可容忍的最大差值
	offset         *list.Element // the next user group offset

	flag          TargetStatus //
	capChangeTime time.Duration
	li            *list.List
	change        int   // 进行扩容缩容操作次数
	limit         int   // max online user for group
	createTime    int64 // create time
}

var targetPool = sync.Pool{New: func() interface{} {
	return &target{
		rw: sync.RWMutex{},
		li: list.New(),
	}
}}

func NewTarget(targetName string, limit int) (*target, error) {
	if targetName == "" || limit == 0 {
		return nil, errors.New("bad param of target")
	}
	tg := targetPool.Get().(*target)
	tg.name = targetName
	tg.limit = limit
	tg.createTime = time.Now().Unix()
	g := GetG(tg.limit)
	elm := tg.li.PushFront(g)
	tg.offset = elm
	tg.numG++
	return tg, nil
}

// ============================================= API =================================

func (t *target) Info() *TargetInfo {
	return t.info()
}

func (t *target) Add(cli Client)ClientManager {
	if cli == nil {
		return nil
	}
	t.add(cli)
	return t
}

func (t *target) Del(token []string) ([]string, int) {
	if token == nil {
		return nil, 0
	}
	return t.del(token)
}

func (t *target) Count() int {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.num
}

func (t *target) BroadCast(data []byte, tags ...string) []string {
	if len(data) == 0 {
		return nil
	}
	return t.broadcast(data, tags...)
}

func (t *target) Expansion() {
	since := time.Now()
	t.rw.Lock()
	defer t.rw.Unlock()
	t.expansion(t.targetG - t.numG)
	escape := time.Since(since)
	logging.Infof("sim :  target Expansion , spend time %v ,", escape)
}

func (t *target) Shrinks() {
	t.rw.Lock()
	t.rw.Unlock()
	shrinksNum  := t.num -t.targetG
	if shrinksNum >= 0 {return}
	since := time.Now()
	t.shrinks(shrinksNum)
	escape := time.Since(since)
	logging.Infof("sim :  target Shrinks , spend time %v ,", escape)
}

func (t *target) Balance() {
	since := time.Now()
	t.balance()
	escape := time.Since(since)
	logging.Infof("sim :  target %v Balance ,online user  %v ,countG  %v , spend time %v ,", t.name, t.num, t.numG, escape)
}

func (t *target) Status() TargetStatus {
	return t.fixStatus()
}

func (t *target) Destroy() {
	if t.num != 0 {
		return
	}
	t.destroy()
}

func (t *target) broadcast(data []byte, tags ...string) []string {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	var res []string

	if len(tags) == 0 {
		for node != nil {
			g := node.Value.(*group)
			res = append(res, g.broadcast(data)...)
			node = node.Next()
		}
	} else {
		for node != nil {
			res = append(res, node.Value.(*group).broadcastWithTag(data, tags)...)
			node = node.Next()
		}
	}
	return res
}

func (t *target) info() *TargetInfo {
	var res = &TargetInfo{}
	t.rw.RLock()
	defer t.rw.RUnlock()
	res.Name = t.name
	res.Limit = t.limit
	res.Online = t.num
	res.NumG = t.numG
	res.Change = t.change
	res.CreateTime = t.createTime
	res.Status = int(t.flag)
	var numG []*map[string]string
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*group)
		numG = append(numG, g.info())
		node = node.Next()
	}
	res.GInfo = numG
	return res
}

func (t *target) add(cli Client) {
	t.rw.Lock()
	defer t.rw.Unlock()
	g := t.offset.Value.(*group)
	if same := g.add(cli); same {
		return
	}
	t.num++
	t.moveOffset()
	return
}

func (t *target) del(token []string) (res []string, current int) {
	t.rw.Lock()
	defer t.rw.Unlock()
	node := t.li.Front()
	for node != nil {
		gp := node.Value.(*group)
		_, result, cur := gp.del(token)
		current += cur
		res = append(res, result...)
		node = node.Next()
	}
	t.num = current
	return
}

func (t *target) moveOffset() {
	if t.offset.Next() != nil {
		t.offset = t.offset.Next()
	} else {
		t.offset = t.li.Front()
	}
}

func (t *target) fixStatus() TargetStatus {
	t.rw.Lock()
	defer t.rw.Unlock()

	// 修正targetG
	tg := t.num/t.limit + 1 // 2 /2 =1 , 3/2 = 1
	if tg == t.targetG {
		return t.flag
	}
	t.targetG = tg

	if t.numG == t.targetG {
		return t.flag
	}

	if t.targetG > t.numG {
		t.flag = TargetStatusShouldEXTENSION
		return t.flag
	}
	if t.targetG < t.numG {
		t.flag = TargetStatusShouldSHRINKS
		return t.flag
	}
	if t.num == 0 && time.Now().Unix()-t.createTime > 120 {
		t.flag = TargetStatusShouldDestroy
		return t.flag
	}
	node := t.li.Front()
	var n []int
	for node != nil {
		n = append(n, node.Value.(*group).num)
		node = node.Next()
	}

	var min, max = 10000, 0
	for _, v := range n {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max-min >= t.maxGOnlineDiff {
		t.flag = TargetStatusShouldReBalance
	}
	return t.flag
}

func (t *target) expansion(num int) {
	for i := 0; i < num; i++ {
		newG := GetG(t.limit)
		t.li.PushBack(newG)
		t.numG += 1
	}
}

func (t *target) shrinks(num int) {
	// 缩容的几个重要问题
	// 1.  什么时候判断是否应该缩容 ：
	// 2.  缩容应该由谁来判定  :  每次删除用户就需要进行判断，
	// 3.  缩容的标准是什么 ： 总在线人数和总的G 所分摊的人数来判断需不需要缩容，但是需要额外容量规划，如果使用量低于30% 可以开启缩容
	// 3.1 缩容的目标是： 将利用率保证在60% 左右
	// 4.  谁来执行shrink : 应该由target 聚合层进行统一状态管理


	var free []Client
	var freeNode []*list.Element
	node := t.li.Front()
	for i := 0; i < num; i++ {
		ng := node.Value.(*group)
		res, err := ng.free()
		if err != nil {
			logging.Error(err)
			break
		}
		free = append(free, res...)
		t.numG--
		ng.Destroy()
		if node.Next() != nil {
			freeNode = append(freeNode, node)
			node = node.Next()
			continue
		}
		break
	}

	for _, v := range freeNode {
		t.li.Remove(v)
	}

	node1 := t.li.Front()
	ng := node1.Value.(*group)
	ng.addMany(free)
}

// @ forTesting
func (t *target) distribute() (res []int) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		res = append(res, node.Value.(*group).num)
		node = node.Next()
	}
	return
}

func (t *target) balance() {
	// 根据当前节点进行平均每个节点的人数
	avg := t.num/t.numG + 1

	// 负载小于等于 10 的节点都属于 紧急节点
	// 负载小于 0 的节点属于立刻节点
	// 获取到所有节点的负载情况 ： 负载小于 0 的优先移除
	var lowLoadG []*group
	var steals []Client
	t.rw.Lock()
	defer t.rw.Unlock()
	t.change++
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*group)
		gnum := g.num
		if gnum > avg {
			// num >avg ,说明超载
			steal := gnum - avg
			st := g.move(steal)
			steals = append(steals, st...)
		} else {
			//  进入这里说明当前节点load 偏低
			diff := avg - gnum
			if diff > 0 {
				if len(steals) > diff {
					g.addMany(steals[:diff])
					steals = steals[diff:]
					node = node.Next()
					continue
				} else {
					g.addMany(steals)
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
		w := avg - g.num
		if w >= len(steals) {
			g.addMany(steals)
			break
		}
		g.addMany(steals[:w])
		steals = steals[w:]
	}
}

func (t *target) destroy() {
	t.createTime, t.num, t.limit, t.numG = 0, 0, 0, 0
	t.flag = 0
	t.li = list.New()
	t.offset = nil
	targetPool.Put(t)
}