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
	"net/http"
)

const (
	// Buffer : buffer the data that need to send, it's not the connection write buffer
	// if you start this option ,there have risk to lose the data
	DefaultConnectionBuffer = 1 << 3

	// ConnectionWriteBuffer : connection write buffer
	DefaultConnectionWriteBuffer = 1 << 10

	// ConnectionReadBuffer : connection read buffer
	DefaultConnectionReadBuffer = 1 << 10
)

const (
	MessageTypeText = iota + 1
	MessageTypeBinary
)

type Option struct {
	ConnectionBuffer      int16 // Buffer the data that need to send
	MessageType           uint8 // Message type
	ConnectionWriteBuffer int16 // connection write buffer
	ConnectionReadBuffer  int16 // connection read buffer

	// CheckOrigin : check the origin, if return false , the connection will be closed
	CheckOrigin func(r *http.Request) bool

	// Error : if occur the error , you can handle it with this function
	Error func(w http.ResponseWriter, r *http.Request, status int, reason error)
}

func validate(option *Option) error {
	if option.ConnectionBuffer < 1 {
		return errors.New("buffer param must be greater than 0")
	} else if option.ConnectionReadBuffer < 1 {
		return errors.New("connection read buffer param must be greater than 0")
	} else if option.ConnectionWriteBuffer < 1 {
		return errors.New("connection write buffer param must be greater than 0")
	} else if option.MessageType != MessageTypeText && option.MessageType != MessageTypeBinary {
		return errors.New("message type param must be 1 or 2")
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
		origin:  message,
		counter: &atomic.Int64{},
	}
}
