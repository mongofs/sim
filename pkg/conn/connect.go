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
	"errors"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io"
	"net/http"
	"sim/pkg/logging"
	"sync"
	"time"
)

type Connection interface {

	// Send : send data to the connection
	Send(data []byte) error

	// GetRemoteAddr : get the remote address
	GetRemoteAddr() string

	// GetLocalAddr : get the local address
	GetLocalAddr() string
}

type Identification interface {

	// GetIdentification : get the identification
	GetIdentification() interface{}

	// SetIdentification : set the identification
	SetIdentification(identification interface{})
}

type HeartBeater interface {

	// FlushHeartBeatTime : flush the heart beat time ,call this  function to reset the time
	// ticker , if no call this function , when the ticker is timeout
	// the connection will be closed
	FlushHeartBeatTime()

	// GetLastHeartBeatTime :  get the last heart beat time
	GetLastHeartBeatTime() int64
}

type ConnectionWithCloser interface {

	// io.Closer : close the connection
	io.Closer

	// Connection : use connection to send data
	Connection
}

type ConnectionWithHeartBeat interface {

	// HeartBeater : flush the heart beat time
	HeartBeater

	// Connection : use connection to send data
	Connection
}

type ConnectionWithIdentification interface {

	// Identification : get the identification
	Identification

	// Connection : use connection to send data
	Connection
}

type ConnectionWithAnyButClose interface {
	Connection

	Identification

	HeartBeater
}

type ConnectionWithAny interface {

	// io.Closer : close the connection
	io.Closer

	Connection

	Identification

	HeartBeater
}

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

	// once : to avoid  multiple close
	once sync.Once

	// con : the websocket connection
	con *websocket.Conn

	// identification : the identification of the connection, it also can be used to
	// find the connection
	identification interface{}

	// buffer : is set for user , this buffer is not the same as websocket buffer
	// this buffer is a slice buffer , and the slice element type is a pointer type ,
	// it means if you set a large content , it will spend a lot of memory,and if your
	// project under heavy load , the memory will increase quickly , so you should set
	// a small buffer or decrease the content size.
	buffer chan []byte

	// heartBeatTime : the last heart beat time
	heartBeatTime int64

	// notify ： to notify container where is the place to save the connection and the
	// identification. the place I implemented  is  bucket :github://mongofs/sim/bucket.go ,
	// because connection status's change can't be captured by the connection itself or
	// read/write operation .
	//
	// connection status's change is very complex , it can't be captured by the connection itself，but
	// we can use TCP read/write operation to capture the connection  change ,  the problem is
	// it is not accurate . for example , when the connection is disconnected very quickly : (if you
	// want implement this  ,you can close the net router)
	//
	// TCP Read : we can use for loop to monitor the read operation , when the read operation return EOF
	// or error , we can assume the connection is disconnected
	// TCP Write : when we write data to the connection , if the connection is disconnected , the write
	// operation will return error , so we can assume the connection is disconnected
	//
	// To solve the problem , we can use heart beat to monitor the connection status's change , the intrinsically
	// is use TCP write operation , and get the call back , if callback is error , we can assume the connection
	// is disconnected.

	notify chan<- interface{}

	// status : the connection status
	status int

	// closeChan : close channel , when the connection is closed , the channel will be closed , the read/write
	// goroutine will be notified
	closeChan chan struct{}

	// messageType : the message type , text or binary
	messageType uint8 // text /binary
}

// ReceiveHandler : this is the callback function when the connection receive data, you must
// implement this function , and pass it to the NewConn function
type ReceiveHandler func(conn ConnectionWithAnyButClose, data []byte)

type CheckOrigin func(r *http.Request) bool

// NewConn : create a new connection
func NewGorillaConn(option *Option, sig chan<- interface{}, w http.ResponseWriter, r *http.Request, Receive ReceiveHandler) (Connection, error) {
	result := &conn{
		once:          sync.Once{},
		buffer:        make(chan []byte, option.BufferSize),
		heartBeatTime: time.Now().Unix(),
		notify:        sig,
		closeChan:     make(chan struct{}),
		messageType:   option.MessageType,
	}
	err := result.upgrade(w, r, option.ConnectionReadBuffer, userOption.ConnectionWriteBuffer)
	if err != nil {
		return nil, err
	}
	result.status = StatusConnectionRunning
	go result.monitorSend()
	go result.monitorReceive(Receive)
	return result, nil
}

func (c *conn) GetRemoteAddr() string {
	return c.con.RemoteAddr().String()
}

func (c *conn) GetLocalAddr() string {
	return c.con.LocalAddr().String()
}

func (c *conn) SetIdentification(identification interface{}) {
	c.identification = identification
}

func (c *conn) GetIdentification() interface{} {
	return c.identification
}

func (c *conn) Send(data []byte) error {
	if c.status != StatusConnectionRunning {
		// judge the status of connection
		return ErrConnectionIsClosed
	}
	if len(c.buffer)*10 > cap(c.buffer)*7 {
		// judge the Send channel first
		return ErrConnectionIsWeak
	}
	c.buffer <- data
	return nil
}

func (c *conn) Close(reason string) {
	c.close(reason)
}

func (c *conn) FlushHeartBeatTime() {
	c.heartBeatTime = time.Now().Unix()
}

func (c *conn) GetLastHeartBeatTime() int64 {
	return c.heartBeatTime
}

func (c *conn) monitorSend() {
	defer func() {
		if err := recover(); err != nil {
			logging.Log.Error("monitorSend", zap.Any("PANIC", err))
		}
	}()
	var err error
	for {
		select {
		case <-c.closeChan:
			goto loop
		case data := <-c.buffer:
			startTime := time.Now()
			err = c.con.WriteMessage(int(c.messageType), data)
			if err != nil {
				logging.Log.Warn("monitorSend", zap.Error(err))
				goto loop
			}
			spendTime := time.Since(startTime)
			if spendTime > time.Duration(2)*time.Second {
				logging.Log.Warn("monitorSend ", zap.Any("Identification", c.identification), zap.Any("week net", spendTime))
			}

		}
	}
loop:
	c.close("monitorSend is closed")
}

func (c *conn) monitorReceive(handleReceive ReceiveHandler) {
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
		c.notify <- c.GetIdentification()
		if len(err) > 0 {
			if err[0] != nil {
				// todo
				logging.Log.Error("close ", zap.Any("ID", c.GetIdentification()), zap.Error(err[0]))
			}
		}
		close(c.closeChan)
		if err := c.con.Close(); err != nil {
			logging.Log.Error("close ", zap.Any("ID", c.GetIdentification()), zap.Error(err))
		}
		logging.Log.Info("close", zap.Any("ID", c.GetIdentification()), zap.String("OFFLINE_CAUSE", cause))
	})
}

func (c *conn) upgrade(w http.ResponseWriter, r *http.Request, readerSize, writeSize int, checkOrigin CheckOrigin) error {
	conn, err := (&websocket.Upgrader{
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {

		},
		CheckOrigin:      checkOrigin,
		ReadBufferSize:   readerSize,
		WriteBufferSize:  writeSize,
		HandshakeTimeout: time.Duration(5) * time.Second,
	}).Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	c.con = conn
	return nil
}
