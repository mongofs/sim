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

// 这里是track ,出现一个这么样的问题，前端在写入数据的时候经常提到消息未搜到，
// 整体结构： track
// 1. content : map[sequenceID] []*string , 如果说某场比赛发送消息我会记录当前发送成功的用户，
// 2. 由于会出现 用户在消息发送的过程中退出，可能此时消息还用户buffer中还存在很多内容发送，最多为buffer值

type traceSet struct {
	rw sync.RWMutex
	set map[string] *trace
}



//AddMessage  1. Add 添加新消息
func (t * traceSet) AddMessage(sequnceID string ,UserSet map[string]struct{}){
	t.rw.Lock()
	t.set[sequnceID]= newTrace(UserSet)
	t.rw.Unlock()
}

// DelMessage 2. 删除Message
func (t * traceSet)DelMessage(sequnceID string ,userToken string){
	t.rw.RLock()
	con ,ok:= t.set[sequnceID]
	t.rw.RUnlock()
	if ok {
		needDel := con.DelMessage(userToken)
		if needDel {
			t.rw.Lock()
			delete(t.set,sequnceID)
			t.rw.Unlock()

			// 删除单条信息标注
			logging.Infof("信息ID : %v 成功推送给 %v 人，无一人遗漏",)
		}
	}
}

func (t *traceSet)monitor (){
	defer func() {
		if err := recover() ;err !=nil {
			logging.Errorf("sim : 发送panic 错误 ：%v",err)
		}
	}()
	for {
		t.rw.Lock()
		for k,v:= range t.set{
			if time.Now().Unix() - v.Uptime() > 2{
				// 注意查看这个用户的关闭连接日志
				logging.Errorf("sim : 信息ID : %v ,发送失败人员 ：%v ,超时两秒",k,t.set[k])
				delete(t.set,k)
			}
		}
		t.rw.Unlock()
		time.Sleep(800 * time.Millisecond)
	}
}


type trace struct {
	rw sync.RWMutex
	timeNow int64
	set map[string] struct{}
}

func newTrace (data map[string]struct{})*trace{
	return &trace{
		rw: sync.RWMutex{},
		timeNow: time.Now().Unix(),
		set: data,
	}
}

// DelMessage 2. 删除Message
// 返回值表示是否可删除
func (t * trace)DelMessage(userToken string)bool{
	t.rw.Lock()
	delete(t.set,userToken)
	t.rw.Unlock()

	if len(t.set) ==0 {
		return true
	}
	return false
}

// Uptime 1. 如果超过数量len（map）==0 ，就会标注完全下发成功
// 2。 如果超过200ms 还没有处理完成就会被删除
func (t *trace)Uptime()int64{
	return t.timeNow
}