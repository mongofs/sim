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
	"github.com/gorilla/websocket"
	"net/http"
	"sim/pkg/errors"
	"sim/pkg/logging"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	api "sim/api/v1"
)

const (
	waitTime = 1 << 7
)

type Protocol int

const (
	ProtocolJson     Protocol = 1
	ProtocolProtobuf Protocol = 2
)

type MessageType int

const (
	MessageTypeText   MessageType = 1
	MessageTypeBinary MessageType = 2
)

// 当前这个连接基于 github.com/gorilla/websocket
type gorilla struct {
	once   sync.Once
	con    *websocket.Conn
	token  *string
	buffer chan []byte
	output chan []byte

	// 这里是唯一一个伴随业务性质的结构，值得注意的是，在我们实际应用场景中
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
	closeChan     chan<- string
	protocol      Protocol    // json /protobuf
	messageType   MessageType // text /binary
}

func NewGorilla(token *string, closeChan chan<- string, option *Option, w http.ResponseWriter, r *http.Request) (Connect, error) {
	result := &gorilla{
		once:          sync.Once{},
		con:           nil,
		token:         token,
		buffer:        make(chan []byte, 10),
		output:        make(chan []byte, 0),
		heartBeatTime: time.Now().Unix(),
		closeChan:     closeChan,
		protocol:      option.ClientProtocol,
		messageType:   option.ClientMessageType,
	}
	err := result.upgrade(w, r, option.ClientReaderBufferSize, option.ClientWriteBufferSize)
	if err != nil {
		return nil, err
	}
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

func (c *gorilla) Read() (dataCh <-chan []byte) {
	return c.output
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
			logging.Errorf("sim : sendProc 发生panic %v", err)
		}
	}()
	for {
		data := <-c.buffer
		starttime := time.Now()
		err := c.con.WriteMessage(int(c.messageType), data)
		spendTime := time.Since(starttime)
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

func (c *gorilla) monitorReceive() {
	defer func() {
		if err := recover(); err != nil {
			logging.Errorf("sim : recvProc 发生panic %v", err)
		}
	}()
	for {
		_, data, err := c.con.ReadMessage()
		if err != nil {
			goto loop
		}
		// 这里拿到的都是整条消息，
		c.output <- data
	}
loop:
	c.close(false)
}

func (c *gorilla) close(forRetry bool) {
	c.once.Do(func() {
		close(c.output)
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
