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
	"encoding/json"
	"net/http"
	"sync"
	"time"

	api "sim/api/v1"
	"sim/pkg/errors"
	"sim/pkg/logging"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

const (
	waitTime = 1 << 7
)

type Protocol int

const (
	ProtocolJson     Protocol = iota +1
	ProtocolProtobuf
)

type MessageType int

const (
	MessageTypeText   MessageType = iota + 1
	MessageTypeBinary
)

// 当前这个连接基于 github.com/gorilla/websocket
type gorilla struct {
	once   sync.Once
	con    *websocket.Conn
	token  *string
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
	closeChan     chan<- string
	protocol      Protocol    // json /protobuf
	messageType   MessageType // text /binary
}

func NewGorilla(token *string, closeChan chan<- string, option *Option, w http.ResponseWriter, r *http.Request,handlerReceive Receive) (Connect, error) {
	result := &gorilla{
		once:          sync.Once{},
		con:           nil,
		token:         token,
		buffer:        make(chan []byte, option.ClientBufferSize),
		heartBeatTime: time.Now().Unix(),
		closeChan:     closeChan,
		protocol:      option.ClientProtocol,
		messageType:   option.ClientMessageType,
	}
	err := result.upgrade(w, r, option.ClientReaderBufferSize, option.ClientWriteBufferSize)
	if err != nil {
		return nil, err
	}
	go result.monitorSend()
	go result.monitorReceive(handlerReceive)
	return result, nil
}

func (c *gorilla) Token() string {
	return *c.token
}

func (c *gorilla) Send(data []byte) error {
	if len(c.buffer)*10 > cap(c.buffer)*7 {
		// todo 这里处理方式展示不够优雅，用户可以根据自身业务情况处理
		// 丢弃用户消息，// 此时表明用户网络处于非常差的情况，后续将兼容
		// 延迟重发，不过需要引入新的中间件进行消息存储，对于不是强一致性
		// 的场景建议不用存储，在我实际公司业务，如果到这一步了就会让用户
		// 断开连接，等用户网络好后先通过api同步数据
		return errors.ErrUserBufferIsFull
	}
	c.handlerProtocol(data)
	return nil
}

func (c *gorilla) Close(retry bool) error {
	c.close(retry)
	return nil
}

func (c *gorilla) SetMessageType(messageType MessageType) {
	c.messageType = messageType
}

func (c *gorilla) SetProtocol(protocol Protocol) {
	c.protocol = protocol
}

func (c *gorilla) ReFlushHeartBeatTime() {
	c.heartBeatTime = time.Now().Unix()
}

func (c *gorilla) GetLastHeartBeatTime() int64 {
	return c.heartBeatTime
}

func (c *gorilla) handlerProtocol(data []byte) error {
	var (
		sid int64
		d   []byte
		err error
	)
	basic := &api.PushToClient{
		Sid: sid,
		Msg: data,
	}

	switch c.protocol {
	case ProtocolJson:
		d, err = json.Marshal(basic)
	case ProtocolProtobuf:
		d, err = proto.Marshal(basic)
	}
	if err != nil {
		return err
	}
	c.buffer <- d
	return nil
}

func (c *gorilla) monitorSend() {
	defer func() {
		if err := recover(); err != nil {
			logging.Errorf("sim : monitorSend 发生panic %v", err)
		}
	}()
	for {
		data := <-c.buffer
		startTime := time.Now()
		err := c.con.WriteMessage(int(c.messageType), data)
		spendTime := time.Since(startTime)
		if spendTime > time.Duration(2)*time.Second {
			logging.Warnf("sim : token '%v'网络状态不好，消息写入通道时间过长 :'%v'", c.token, spendTime)
		}
		if err != nil {
			goto loop
		}

	}
loop:
	c.close(false)
}

func (c *gorilla) monitorReceive(handleReceive Receive) {
	defer func() {
		if err := recover(); err != nil {
			logging.Errorf("sim : monitorReceive 发生panic %v", err)
		}
	}()
	for {
		_, data, err := c.con.ReadMessage()
		if err != nil {
			goto loop
		}
		handleReceive.Handle(c,data)
	}
loop:
	c.close(false)
}

func (c *gorilla) close(forRetry bool) {
	c.once.Do(func() {
		if !forRetry {
			c.closeChan <- *c.token
		}
		c.con.Close()
		logging.Infof("sim : token %v 正常下线", c.token)
	})
}

func (c *gorilla) upgrade(w http.ResponseWriter, r *http.Request, readerSize, writeSize int) error {
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
