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
package conn

import (
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"

	"github.com/mongofs/sim/pkg/logging"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	StatusConnectionClosed = iota + 1
	StatusConnectionRunning
)

var (
	ErrConnectionIsClosed = errors.New("connection is closed")
	ErrConnectionIsWeak   = errors.New("connection is in weak status")
)

// This connection is upgrade of  github.com/gorilla/websocket
type conn struct {
	once           sync.Once
	con            *websocket.Conn
	identification string
	// buffer 这里是用户进行设置缓冲区的，这里和websocket的缓冲区不同的是，这里的内容是单独
	// 按照消息个数来缓冲的，而websocket是基于tcp的缓冲区进行字节数组缓冲，本质是不同
	// 的概念，值得注意的是，slice是指针类型，意味着传输的内容是可以很大的，在chan层
	// 表示仅仅是8字节的指针，建议单个传输内容不要太大，否则在用户下发的过程中如果用户网络
	// 不是很好，TCP连接写入能力较差，内容都会堆积在内存中导致内存上涨，这个参数也建议不要
	// 设置太大，建议在8个
	buffer chan []byte

	// heartBeatTime 这里是唯一一个伴随业务性质的1结构，值得注意的是，在我们实际应用场景中
	// 这里会容易出错，如果我将连接本身close掉，然后将连接标示放入closeChan，此时
	// 如果通道阻塞，本次连接的用户拿着同样的token进行连接，那么就会出现新的
	// 连接在bucket不存在的情况，建议做法是：最后在客户端能保证，每次发起连接
	// 都是一个全新的token，这样就能完全隔离掉这种情况
	// 由于本身业务复杂性，客户端某些功能不能实现，那么就只能采取：建立连接在
	// 之前先查后写，目前默认采取这种方案，但是又会伴随另外一个问题： 如果旧链接
	// 依旧在线，那么就得发送信号释放old conn ，整体性能就会降低
	// 针对第二种，我们踩过坑： 前台调用接口进入具体聊天室，聊天室内用户一直停留
	// 用户连接死掉或者被客观下线，前台发起重连，然后旧的连接下线新的链接收不到消息
	heartBeatTime int64

	// closeChan 是一个上层传入的一个chan，当用户连接关闭，可以将本身token传入closeChan
	// 通知到bucket层以及其他层进行处理，但是bucket作为connect管理单元，在做上层channel监听
	// 的时候尽量不要读取closeChan

	notify chan<- string

	status int

	// closeChan
	closeChan   chan struct{}
	messageType MessageType // text /binary
}

type Receive func(conn Connect, data []byte)

func NewConn(Id string, sig chan<- string, w http.ResponseWriter, r *http.Request, Receive Receive) (Connect, error) {
	result := &conn{
		once:           sync.Once{},
		identification: Id,
		buffer:         make(chan []byte, userOption.Buffer),
		heartBeatTime:  time.Now().Unix(),
		notify:         sig,
		closeChan:      make(chan struct{}),
		messageType:    userOption.MessageType,
	}
	err := result.upgrade(w, r, userOption.ConnectionReadBuffer, userOption.ConnectionWriteBuffer)
	if err != nil {
		return nil, err
	}
	result.status = StatusConnectionRunning
	go result.monitorSend()
	go result.monitorReceive(Receive)
	return result, nil
}

func (c *conn) Identification() string {
	return c.identification
}

func (c *conn) Send(data []byte) error {
	if c.status != StatusConnectionRunning {
		// judge the status of connection
		return ErrConnectionIsClosed
	}
	if len(c.buffer)*10 > cap(c.buffer)*7 {
		sendLoseContent.Inc()
		// judge the Send channel first
		return ErrConnectionIsWeak
	}
	c.buffer <- data
	return nil
}

func (c *conn) Close(reason string) {
	c.close(reason)
}

func (c *conn) ReFlushHeartBeatTime() {
	c.heartBeatTime = time.Now().Unix()
}

func (c *conn) GetLastHeartBeatTime() int64 {
	return c.heartBeatTime
}

var (
	sendContent       *atomic.Int64 = &atomic.Int64{}
	sendContentLength *atomic.Int64 = &atomic.Int64{}
	sendLoseContent   *atomic.Int64 = &atomic.Int64{}
)

func SwapSendData() (content, loseContent, contentLength int64) {
	content = sendContent.Swap(0)
	loseContent = sendLoseContent.Swap(0)
	contentLength = sendContentLength.Swap(0)
	return
}

func (c *conn) monitorSend() {
	defer func() {
		if err := recover(); err != nil {
			logging.Log.Error("monitorSend", zap.Any("PANIC", err))
		}
	}()
	for {
		select {
		case <-c.closeChan:
			goto loop
		case data := <-c.buffer:
			startTime := time.Now()
			err := c.con.WriteMessage(int(c.messageType), data)
			if err != nil {
				logging.Log.Warn("monitorSend", zap.Error(err))
				goto loop
			}
			spendTime := time.Since(startTime)
			if spendTime > time.Duration(2)*time.Second {
				logging.Log.Warn("monitorSend weak net ", zap.String("ID", c.identification), zap.Any("WEAK_NET", spendTime))
			}
			sendContent.Inc()
			sendContentLength.Add(int64(len(data)))
		}
	}
loop:
	c.close("monitorSend")
}

func (c *conn) monitorReceive(handleReceive Receive) {
	defer func() {
		if err := recover(); err != nil {
			logging.Log.Error("monitorReceive ", zap.Any("panic", err))
		}
	}()
	var temErr error
	for {
		_, data, err := c.con.ReadMessage()
		if err != nil {
			temErr = err
			goto loop
		}
		handleReceive(c, data)
	}
loop:
	c.close("monitorReceive", temErr)
}

func (c *conn) close(cause string, err ...error) {
	c.once.Do(func() {
		c.status = StatusConnectionClosed
		c.notify <- c.Identification()
		if len(err) > 0 {
			if err[0] != nil {
				// todo
				//logging.Log.Error("close ", zap.String("ID",c.identification),zap.Error(err[0]))
			}
		}
		close(c.closeChan)
		if err := c.con.Close(); err != nil {
			logging.Log.Error("close ", zap.String("ID", c.identification), zap.Error(err))
		}
		logging.Log.Info("close", zap.String("ID", c.identification), zap.String("OFFLINE_CAUSE", cause))
	})
}

func (c *conn) upgrade(w http.ResponseWriter, r *http.Request, readerSize, writeSize int) error {
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
	c.con = conn
	return nil
}
