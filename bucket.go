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
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"time"

	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
	"go.uber.org/atomic"
)

type bucketInterface interface {
	// you can register the user to the bucket set
	Register(client conn.Connect) (string,int64,error)

	// you can offline the user in anytime
	Offline(identification string)

	// send message to users , if empty of users set ,will send message to all users
	SendMessage(message []byte, users ...string /* if no param , it will use broadcast */)

	// return the signal channel , you can use the channel to notify the bucket
	// uses is offline , and delete the users' identification
	SignalChannel() chan<- string

	Count() int
}

type bucket struct {

	id string
	rw sync.RWMutex
	// Element Number
	np atomic.Int64
	// users set
	users map[string]conn.Connect
	// Here is  different point you need pay attention
	// the close signal received by component that is connection
	// so we need use channel to inform bucket that user is out of line
	closeSig chan string
	ctx context.Context
	callback func()
	opts *Options
}

func NewBucket(option *Options, i int ) *bucket {
	res := &bucket{
		id : "bucket_" + strconv.Itoa(i),
		rw:       sync.RWMutex{},
		np:       atomic.Int64{},
		closeSig: make(chan string),
		opts:     option,
	}
	res.users = make(map[string]conn.Connect, res.opts.BucketSize)
	go res.monitorDelChannel()
	go res.keepAlive()
	return res
}

func (h *bucket) Offline(identification string) {
	h.rw.RLock()
	cli, ok := h.users[identification]
	h.rw.RUnlock()
	if ok {
		cli.Close()
	}
}

func (h *bucket) Register(cli conn.Connect) (string,int64,error) {
	if cli == nil {
		return "",0,errors.New("sim : the obj of cli is nil ")
	}
	h.rw.Lock()
	defer h.rw.Unlock()
	old, ok := h.users[cli.Identification()]
	if ok {
		old.Close()
	}
	h.users[cli.Identification()] = cli
	h.np.Add(1)
	return h.id,h.np.Load(),nil
}

func (h *bucket) SendMessage(message []byte, users ...string /* if no param , it will use broadcast */) {
	if len(users)-1 >= 0 {
		for _, user := range users {
			h.send(message, user, false)
		}
		return
	}
	h.broadCast(message, false)
}

func (h *bucket) SignalChannel() chan<- string {
	return h.closeSig
}

func (h *bucket) Count() int {
	h.rw.RLock()
	defer h.rw.RUnlock()
	return len(h.users)
}

// this function need a lot of  logs
func (h *bucket) send(data []byte, token string, Ack bool) {
	h.rw.RLock()
	cli, ok := h.users[token]
	h.rw.RUnlock()
	if !ok { // user is not online
		return
	} else {
		err := cli.Send(data)
		// todo
		logging.Error(err)
	}
	return
}

func (h *bucket) broadCast(data []byte, Ack bool) {
	h.rw.RLock()
	for _, cli := range h.users {
		err := cli.Send(data)
		if err != nil {
			// todo
			logging.Error(err)
			continue
		}
	}
	h.rw.RUnlock()
}

func (h *bucket) delUser(identification string) {
	h.rw.Lock()
	defer h.rw.Unlock()
	_, ok := h.users[identification]
	if !ok {
		return
	}
	delete(h.users, identification)
	//更新在线用户数量
	h.np.Add(-1)

	if h.callback != nil {
		h.callback()
	}
}

// To monitor the whole bucket
// run in a goroutine
func (h *bucket) monitorDelChannel() {
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

// To keepAlive the whole bucket
// run in a goroutine
func (h *bucket) keepAlive() {
	if h.opts.ClientHeartBeatInterval == 0 {
		return
	}
	for {
		var cancelCli []conn.Connect
		now := time.Now().Unix()
		h.rw.Lock()
		for _, cli := range h.users {
			inter := now - cli.GetLastHeartBeatTime()
			if inter < 2*int64(h.opts.ClientHeartBeatInterval) {
				continue
			}
			cancelCli = append(cancelCli, cli)
		}
		h.rw.Unlock()
		for _, cancel := range cancelCli {
			cancel.Close()
		}
		time.Sleep(10 * time.Second)
	}
}


