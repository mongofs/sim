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
	"context"
	"net/http"
	"sim/pkg/logging"
	"sync"
	"time"

	"sim/pkg/errors"

	"go.uber.org/atomic"
)

type Monitor func()

type bucket struct {
	rw sync.RWMutex

	// Number of people
	np *atomic.Int64

	// users set
	clis map[string]Client

	// User offline notification
	closeSig chan string

	ctx context.Context

	callback func()

	// monitorHeartbeat  监控心跳的方法，需要外部传入，不传入默认是不存在的，但是为了保障连接的高可用性
	// 在服务启动的时候是一个必传参，monitorHeartbeat是一个阻塞的方法，在bucket启动的时候进行赋值
	// 建议这种写法
	//	for {
	//		cancelCli := []Client{}
	//		now := time.Now().Unix()
	//		b.rw.Lock()
	//		for _, cli := range b.clis {
	//
	//			interval := now - cli.LastHeartBeat()
	//
	//			if interval < 2*int64(b.opts.ClientHeartBeatInterval) {
	//				continue
	//			}
	//			cancelClis = append(cancelClis, cli)
	//		}
	//		b.rw.Unlock()
	//		for _, cancel := range cancelClis {
	//			cancel.Offline()
	//		}
	//
	//		time.Sleep(10 * time.Second)
	//	}
	monitorHeartBeat Monitor

	opts *Option
}

func newBucket(option *Option) Bucket {
	res := &bucket{
		rw:       sync.RWMutex{},
		np:       &atomic.Int64{},
		closeSig: make(chan string),
		opts:     option,
	}
	res.clis = make(map[string]Client, res.opts.BucketSize)
	res.start()
	return res
}

func (h *bucket) CreateConn(w http.ResponseWriter, r *http.Request, token string) (Client, error) {
	return NewClient(w, r, h.closeSig, &token, h.ctx, h.opts)
}

func (h *bucket) Online() int64 {
	return h.np.Load()
}

func (h *bucket) Send(data []byte, token string, Ack bool) error {
	return h.send(data, token, Ack)
}

func (h *bucket) BroadCast(data []byte, Ack bool) error {
	return h.broadCast(data, Ack)
}

func (h *bucket) OffLine(token string) {
	h.rw.RLock()
	cli, ok := h.clis[token]
	h.rw.RUnlock()
	if ok {
		cli.Offline()
	}
}

func (h *bucket) Register(cli Client, token string) error {
	if cli == nil {
		return errors.ErrCliISNil
	}
	h.rw.Lock()
	defer h.rw.Unlock()
	old, ok := h.clis[token]
	if ok {
		clienter, _ := old.(*Cli)
		clienter.OfflineForRetry(true)
	}
	h.clis[token] = cli
	h.np.Add(1)
	return nil
}

func (h *bucket) send(data []byte, token string, Ack bool) error {
	h.rw.RLock()
	cli, ok := h.clis[token]
	h.rw.RUnlock()
	if !ok { //用户不在线
		return errors.ErrCliISNil
	} else {
		return cli.Send(data)
	}
}

func (h *bucket) broadCast(data []byte, Ack bool) error {
	h.rw.RLock()
	for _, cli := range h.clis {
		err := cli.Send(data)
		if err != nil {
			logging.Error(err)
			continue
		}
	}
	h.rw.RUnlock()
	return nil
}

func (h *bucket) start() {
	go h.monitor()
	go h.keepAlive()
}

func (h *bucket) delUser(token string) {
	h.rw.Lock()
	delete(h.clis, token)
	h.rw.Unlock()

	//更新在线用户数量
	h.np.Add(-1)

	// 这里去通知wti的内容
	Update(token)
	if h.callback != nil {
		h.callback()
	}
}

func (h *bucket) monitor() {
	if h.ctx != nil {
		for {
			select {
			case token := <-h.closeSig:
				h.delUser(token)
			case <-h.ctx.Done():
				return
			}
		}
	}
	for {
		select {
		case token := <-h.closeSig:
			h.delUser(token)
		}
	}
}

func (b *bucket) keepAlive() {

	for {
		cancelClis := []Client{}
		now := time.Now().Unix()
		b.rw.Lock()
		for _, cli := range b.clis {

			interval := now - cli.LastHeartBeat()

			if interval < 2*int64(b.opts.ClientHeartBeatInterval) {
				continue
			}
			cancelClis = append(cancelClis, cli)
		}
		b.rw.Unlock()
		for _, cancel := range cancelClis {
			cancel.Offline()
		}

		time.Sleep(10 * time.Second)
	}
}
