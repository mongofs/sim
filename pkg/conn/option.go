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
	"go.uber.org/atomic"
)

type MessageType uint

const (
	Buffer                = 1 << 3
	ConnectionWriteBuffer = 1 << 10
	ConnectionReadBuffer  = 1 << 10
)

const (
	MessageTypeText MessageType = iota + 1
	MessageTypeBinary
)

// Various errors contained in OpError.
var (
	// For connection write buffer param
	ErrConnWriteBufferParam = errors.New("conn write buffer param is wrong err , the value must bigger the 1")

	// For connection read buffer param
	ErrConnReadBufferParam = errors.New("conn read buffer param is wrong err , the value must bigger the 1")

	// For connection  buffer param
	ErrBufferParam = errors.New("conn  buffer param is wrong err , the value must bigger the 1")
	// For connection  Message Type param
	ErrMessageTypeParam = errors.New("conn  MessageType param is wrong err , the value must be 1 or 2")
)

type Option struct {
	Buffer                int         // Buffer the data that need to send
	MessageType           MessageType // Message type
	ConnectionWriteBuffer int         // connection write buffer
	ConnectionReadBuffer  int         // connection read buffer
}

func DefaultOption() *Option {
	return &Option{
		Buffer:                Buffer,
		MessageType:           MessageTypeText,
		ConnectionWriteBuffer: ConnectionWriteBuffer,
		ConnectionReadBuffer:  ConnectionReadBuffer,
	}
}

var userOption *Option = DefaultOption()

func SetOption(option *Option) error {
	if err := validate(option); err != nil {
		return err
	}
	userOption = option
	return nil
}

func validate(option *Option) error {
	if option.Buffer < 1 {
		return ErrBufferParam
	} else if option.ConnectionReadBuffer < 1 {
		return ErrConnReadBufferParam
	} else if option.ConnectionWriteBuffer < 1 {
		return ErrConnWriteBufferParam
	} else if option.MessageType != MessageTypeText && option.MessageType != MessageTypeBinary {
		return ErrMessageTypeParam
	} else {
		return nil
	}
}

// counter message wrapper add a counter for message , the counter is for record
// that times of message send by net card
type CounterMessageWrapper struct {
	origin  *[]byte
	counter *atomic.Int64
}

// wrap message
func WrapSendMessage(message *[]byte) CounterMessageWrapper {
	return CounterMessageWrapper{
		origin: message,
		counter: &atomic.Int64{},
	}
}
