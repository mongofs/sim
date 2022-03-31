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
	"fmt"
	"net/http"
	"sync"

	"sim/pkg/errors"

	"go.uber.org/atomic"
)

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

	opts *Option
}

func newBucket(option *Option) Bucket {
	res := &bucket{
		rw:       sync.RWMutex{},
		np:       &atomic.Int64{},
		closeSig: make(chan string, 0),
		opts:     option,
	}
	res.clis = make(map[string]Client, res.opts.BucketSize)
	res.start()
	return res
}

func (h *bucket) Flush() {
	h.rw.RLock()
	defer h.rw.RUnlock()
	h.np.Store(int64(len(h.clis)))
}

func (h *bucket) CreateConn(w http.ResponseWriter, r *http.Request, token string) (Client, error) {
	return NewClient(w, r, h.closeSig, &token, h.ctx, h.opts)
}

func (h *bucket) Online() int64 {
	return h.np.Load()
}

func (h *bucket) Send(data []byte, token string, Ack bool) error {
	h.rw.RLock()
	cli, ok := h.clis[token]
	h.rw.RUnlock()
	if !ok { //用户不在线
		return errors.ErrCliISNil
	} else {
		return h.send(cli, token, data, Ack)
	}
}

func (h *bucket) BroadCast(data []byte, Ack bool) error {
	counter := 0
	success := 0

	failedTokens := ""
	h.rw.RLock()
	for token, cli := range h.clis {
		err := h.send(cli, token, data, Ack)
		if err != nil {
			counter++
			failedTokens = failedTokens + "." + token
		} else {
			success++
		}
	}
	h.rw.RUnlock()
	if counter != 0 {
		return fmt.Errorf("im/bucket :  bucket 广播成功数量 %v ，广播失败数量 is %v,具体tokens :%s", success, counter, failedTokens)
	}
	return nil
}

func (h *bucket) OffLine(token string) {
	h.rw.RLock()
	cli, ok := h.clis[token]
	h.rw.RUnlock()
	if ok {
		cli.Offline()
	}
}

// 将用户注册到bucket中
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

func (h *bucket) IsOnline(token string) bool {
	h.rw.RLock()
	defer h.rw.RUnlock()
	if _, ok := h.clis[token]; ok {
		return true
	}
	return false
}

func (h *bucket) NotifyBucketConnectionIsClosed() chan<- string {
	return h.closeSig
}

func (h *bucket) send(cli Client, token string, data []byte, ack bool) error {
	if ack {
		//todo 这里如果被设置，那么就应该给用户
		return cli.Send(data)
	} else {
		return cli.Send(data)
	}
	return nil
}
