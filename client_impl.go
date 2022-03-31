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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sim/pkg/logging"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Cli struct {
	lastHeartBeatT int64
	conn           *websocket.Conn
	reader         *http.Request
	token          *string
	closeFunc      sync.Once
	done           chan struct{}
	ctx            context.Context
	buf            chan []byte
	closeSig       chan<- string
	handleReceive  Receive

	protocol    int // json /protobuf
	messageType int // text /binary
}

func NewClient(w http.ResponseWriter, r *http.Request, closeSig chan<- string, token *string, ctx context.Context,
	option *Option) (Client, error) {
	res := &Cli{
		lastHeartBeatT: time.Now().Unix(),
		done:           make(chan struct{}),
		reader:         r,
		closeFunc:      sync.Once{},
		buf:            make(chan []byte, option.ClientBufferSize),
		token:          token,
		ctx:            ctx,
		closeSig:       closeSig,
		protocol:       option.ClientProtocol,
		messageType:    option.ClientMessageType,
		handleReceive:  option.ServerReceive,
	}
	if err := res.upgrade(w, r, option.ClientReaderBufferSize, option.ClientWriteBufferSize); err != nil {
		return nil, err
	}
	if err := res.start(); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Cli) Token() string {
	return *c.token
}

func (c *Cli) SetMessageType(messageType int) {
	c.messageType = messageType
}

func (c *Cli) SetProtocol(protocal int) {
	c.protocol = protocal
}

func (c *Cli) Send(data []byte, i ...int64) error {
	var (
		sid int64
		d   []byte
		err error
	)
	if len(i) > 0 {
		sid = i[0]
	}
	basic := &temData{
		Sid: sid,
		Msg: data,
	}
	if c.protocol == ProtocolJson {
		d, err = json.Marshal(basic)
	} else {
		//d, err = proto.Marshal(basic)
	}
	if err != nil {
		return err
	}
	if err := c.send(d); err != nil {
		return err
	}
	return nil
}

func (c *Cli) LastHeartBeat() int64 {
	return c.lastHeartBeatT
}

func (c *Cli) Offline() {
	c.close(false)
}

func (c *Cli) ResetHeartBeatTime() {
	c.lastHeartBeatT = time.Now().Unix()
}

func (c *Cli) Request() *http.Request {

	return c.reader
}

type temData struct {
	Sid int64
	Msg []byte
}

func (c *Cli) upgrade(w http.ResponseWriter, r *http.Request, readerSize, writeSize int) error {
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  readerSize,
		WriteBufferSize: writeSize,
	}).Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Cli) send(data []byte) error {
	if len(c.buf)*10 > cap(c.buf)*7 {
		// 记录当前用户被丢弃的信息
		//c.log.Infof(fmt.Sprintf("im/client: 用户消息通道堵塞 , token is %s ,len %v but user cap is %v",c.token,len(c.buf),cap(c.buf)))

		return errors.New(fmt.Sprintf("im/client: too much data , user len %v but user cap is %s", len(c.buf), cap(c.buf)))
	}

	c.buf <- data
	return nil
}

func (c *Cli) OfflineForRetry(retry ...bool) {
	c.close(retry...)
}

func (c *Cli) start() error {
	go c.sendProc()
	go c.recvProc()
	return nil
}

func (c *Cli) sendProc() {
	defer func() {
		if err := recover(); err != nil {
			logging.Errorf("sim : sendProc 发生panic %v",err)
		}
	}()
	for {
		select {
		case data := <-c.buf:
			starttime := time.Now()
			err := c.conn.WriteMessage(c.messageType, data)
			spendTime := time.Since(starttime)
			if spendTime > time.Duration(2)*time.Second {
				logging.Warnf("sim : token '%v'网络状态不好，消息写入通道时间过长 :'%v'", c.token,spendTime)
			}
			if err != nil {
				goto loop
			}
		case <-c.done:
			goto loop
		}
	}
loop:
	c.close()
}

func (c *Cli) close(forRetry ...bool) {
	flag := false
	if len(forRetry) > 0 {
		flag = forRetry[0]
	}

	c.closeFunc.Do(func() {
		close(c.done)
		c.conn.Close()
		if ! flag {
			c.closeSig <- *c.token
		}
		logging.Infof("sim : token %v 正常下线",c.token)
	})
}

func (c *Cli) recvProc() {
	defer func() {
		if err := recover(); err != nil {
			logging.Errorf("sim : recvProc 发生panic %v",err)
		}
	}()
	for {
		select {
		case <-c.done:
			goto loop
		default:
			_, data, err := c.conn.ReadMessage()
			if err != nil {
				goto loop
			}
			c.handleReceive.Handle(c, data)
		}
	}
loop:
	c.close()
}
