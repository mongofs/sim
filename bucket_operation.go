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

import "time"

func (h *bucket) start (){
	go h.monitor()
	go h.keepAlive()
}

var temcounter  atomic.Int64


// 删除用户
func (h *bucket)monitor (){
	if h.opts.ctx !=nil {
		for  {
			select {
			case token :=<- h.closeSig	:
				h.delUser(token)
			case <- h.opts.ctx.Done():
				return
			}
		}
	}
	for  {
		select {
		case token :=<- h.closeSig	:
			h.delUser(token)
		}
	}
}


// 在线心跳
func (b *bucket)keepAlive (){
	if b.opts.ctx !=nil {
		for {
			select {
			case <-b.opts.ctx.Done():
				return
			default:
				cancelClis := []Client{}
				now := time.Now().Unix()
				b.rw.Lock()
				for _, cli := range b.clis {
					// 如果心跳间隔 时间超过两个心跳包的时间，那么默认用户连接不可用
					if now-cli.LastHeartBeat() < 2*b.opts.HeartBeatInterval {
						continue
					}
					cancelClis = append(cancelClis,cli)
				}
				b.rw.Unlock()
				for _,cancel := range cancelClis{
					cancel.Offline()
				}
			}
			time.Sleep(10 * time.Second)
		}
	}

	for {
		cancelClis := []Client{}
		now := time.Now().Unix()
		b.rw.Lock()
		for _, cli := range b.clis {
			// 如果心跳间隔 时间超过两个心跳包的时间，那么默认用户连接不可用
			interval := now-cli.LastHeartBeat()

			if  interval< 2*b.opts.HeartBeatInterval {
				continue
			}
			cancelClis = append(cancelClis,cli)
		}
		b.rw.Unlock()
		for _,cancel := range cancelClis{
			cancel.Offline()
		}

		time.Sleep(10 * time.Second)
	}
}





// 删除用户
func (h *bucket)delUser(token string) {
	h.rw.Lock()
	delete(h.clis, token)
	h.rw.Unlock()
	h.np.Add(-1)
	// todo 这里需要用个观察者模式
	wti.Update(token)
	if h.opts.callback != nil {
		h.opts.callback()
	}
}
