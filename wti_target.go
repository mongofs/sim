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

// target 存放的内容是 target -> groupA->GroupB
type target struct {
	rw     sync.RWMutex
	name   string
	num    int           // online user
	numG   int           // online Group
	offset *list.Element // the next user group offset

	flag          targetFlag //
	capChangeTime time.Duration
	li            *list.List
	change        int   // 进行扩容缩容操作次数
	limit         int   // max online user for group
	createTime    int64 // create time
}

type targetInfo struct {
	name       string
	online     int
	limit      int
	createTime int64
	status     int
	numG       int
	change     int //状态变更次数
	GInfo      []*map[string]string
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
	tg.Init(targetName)
	return tg, nil
}

func (t *target) Init(name string) *target {
	g := NewGroup(t.limit)
	elm := t.li.PushFront(g)
	t.offset = elm
	t.name = name
	t.numG ++
	return t
}

// ============================================= API =================================

func (t *target) Info() *targetInfo {
	return t.info()
}

func (t *target) Add(cli Client) {
	if cli == nil {
		return
	}
	t.add(cli)
}

func (t *target) Num() int {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.num
}

func (t *target) Del(token []string) ([]string, int) {
	if token == nil {
		return nil, 0
	}
	return t.del(token)
}

func (t *target) BroadCast(data []byte) []string {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	var res []string
	for node != nil {
		g := node.Value.(*Group)
		res = append(res, g.BroadCast(data)...)
		node = node.Next()
	}
	return res
}

func (t *target) BroadCastWithInnerJoinTag(data []byte, otherTag []string) (res []string) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		res = append(res, node.Value.(*Group).BroadCastWithOtherTag(data, otherTag)...)
		node = node.Next()
	}
	return
}

func (t *target) Expansion() {
	since := time.Now()
	t.expansion()
	escape := time.Since(since)
	logging.Infof("sim :  target Expansion , spend time %v ,", escape)
}

func (t *target) Shrinks() {
	since := time.Now()
	t.shrinks()
	escape := time.Since(since)
	logging.Infof("sim :  target Shrinks , spend time %v ,", escape)
}

func (t *target) Status() targetFlag {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return t.flag
}

func (t *target) Balance() {
	since := time.Now()
	t.reBalance()
	escape := time.Since(since)
	logging.Infof("sim :  target Banlance , spend time %v ,", escape)

}

func (t *target) Distribute() (res []int) {
	return t.distribute()
}

func (t *target) Destroy() {
	if t.num != 0 {
		return
	}
	t.destroy()
}

// ======================================== helper =====================================

func (t *target) info() *targetInfo {
	var res = &targetInfo{}
	t.rw.RLock()
	defer t.rw.RUnlock()
	res.name = t.name
	res.limit = t.limit
	res.online = t.num
	res.numG = t.numG
	res.change =t.change
	res.createTime = t.createTime
	res.status = int(t.flag)
	var numG []*map[string]string
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*Group)
		numG = append(numG, g.Info())
		node = node.Next()
	}
	res.GInfo=numG
	return res
}

func (t *target) add(cli Client) {
	t.rw.Lock()
	defer t.rw.Unlock()
	g := t.offset.Value.(*Group)
	if same := g.Add(cli); same {
		return
	}
	t.num++
	t.moveOffset()
	t.judgeExpansion()
	return
}

func (t *target) del(token []string) (res []string, current int) {
	t.rw.Lock()
	defer t.rw.Unlock()
	node := t.li.Front()
	for node != nil {
		gp := node.Value.(*Group)
		_, result, cur := gp.Del(token)
		current += cur
		res = append(res, result...)
		node = node.Next()
	}
	t.num = current
	t.judgeShrinks()
	return
}

func (t *target) moveOffset() {
	if t.offset.Next() != nil {
		t.offset = t.offset.Next()
	} else {
		t.offset = t.li.Front()
	}
}

func (t *target) setFlag(flag targetFlag) {
	flagName := ""
	switch flag {
	case TargetFlagNORMAL:
		flagName = "TargetFlagNORMAL"
	case TargetFLAGShouldEXTENSION:
		flagName = "TargetFLAGShouldEXTENSION"
	case TargetFLAGEXTENSION:
		flagName = "TargetFLAGEXTENSION"
	case TargetFLAGShouldSHRINKS:
		flagName = "TargetFLAGShouldSHRINKS"
	case TargetFLAGSHRINKS:
		flagName = "TargetFLAGSHRINKS"
	case TargetFLAGShouldReBalance:
		flagName = "TargetFLAGShouldReBalance"
	case TargetFLAGREBALANCE:
		flagName = "TargetFLAGREBALANCE"

	}
	_ = flagName
	//logging.Infof("sim/wti : change target level %v ,target name is %v", flagName, t.name)
	t.flag = flag
}

func (t *target) expansion() {
	t.rw.Lock()
	defer t.rw.Unlock()
	targetG := t.num/t.limit + 1
	if t.numG >= targetG {
		t.setFlag(TargetFlagNORMAL)
		return
	}
	diff := targetG - t.numG
	t.setFlag(TargetFLAGEXTENSION)
	for i := 0; i < diff; i++ {
		newG := NewGroup(t.limit)
		t.li.PushBack(newG)
		t.numG += 1
	}
	t.setFlag(TargetFLAGShouldReBalance)
}

func (t *target) judgeExpansion() {
	if t.num > t.limit*t.numG && t.flag == TargetFlagNORMAL {
		t.flag = TargetFLAGShouldEXTENSION
	}
}

func (t *target) shrinks() {
	// 缩容的几个重要问题
	// 1.  什么时候判断是否应该缩容 ：
	// 2.  缩容应该由谁来判定  :  每次删除用户就需要进行判断，
	// 3.  缩容的标准是什么 ： 总在线人数和总的G 所分摊的人数来判断需不需要缩容，但是需要额外容量规划，如果使用量低于30% 可以开启缩容
	// 3.1 缩容的目标是： 将利用率保证在60% 左右
	// 4.  谁来执行shrink : 应该由target 聚合层进行统一状态管理
	t.rw.Lock()
	t.rw.Unlock()
	t.setFlag(TargetFLAGSHRINKS)

	targetG := (t.num*10)/(t.limit*6) + 1 // 100 / 25 = 4 , 1000 /125 =6
	if t.numG <= targetG {
		t.setFlag(TargetFlagNORMAL)
		return
	}

	diff := t.numG - targetG

	var free []Client
	var freeNode []*list.Element
	node := t.li.Front()
	for i := 0; i < diff; i++ {
		ng := node.Value.(*Group)
		res, err := ng.Free()
		if err != nil {
			logging.Error(err)
			break
		}
		free = append(free, res...)
		t.numG--
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
	ng := node1.Value.(*Group)
	ng.BatchAdd(free)
	t.setFlag(TargetFLAGShouldReBalance)
}

func (t *target) judgeShrinks() {
	if (t.num/t.numG)*10 < t.limit*3 && t.flag == TargetFlagNORMAL {
		t.flag = TargetFLAGShouldSHRINKS
	}
	return
}

func (t *target) reBalance() {
	// 根据当前节点进行平均每个节点的人数
	avg := t.num/t.numG + 1

	// 负载小于等于 10 的节点都属于 紧急节点
	// 负载小于 0 的节点属于立刻节点
	// 获取到所有节点的负载情况 ： 负载小于 0 的优先移除
	var lowLoadG []*Group
	var steals []Client
	t.rw.Lock()
	defer t.rw.Unlock()
	t.change++
	node := t.li.Front()
	for node != nil {
		g := node.Value.(*Group)
		if g.Num() > avg {
			// num >avg ,说明超载
			steal := g.Num() - avg
			st, err := g.Move(steal)
			if err != nil {
				return
			}
			steals = append(steals, st...)
		} else {
			//  进入这里说明当前节点load 偏低
			diff := avg - g.Num()
			if diff > 0 {
				if len(steals) > diff {
					g.BatchAdd(steals[:diff])
					steals = steals[diff:]
					node = node.Next()
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

func (t *target) distribute() (res []int) {
	t.rw.RLock()
	defer t.rw.RUnlock()
	node := t.li.Front()
	for node != nil {
		res = append(res, node.Value.(*Group).num)
		node = node.Next()
	}
	return
}

func (t *target) destroy() {
	t.createTime, t.num, t.limit, t.numG = 0, 0, 0, 0
	t.flag = 0
	t.li = nil
	t.offset = nil
	targetPool.Put(t)
}
