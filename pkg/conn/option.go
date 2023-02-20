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

type MessageType uint

const (
	Buffer                = 1 << 3
	ConnectionWriteBuffer = 1 << 10
	ConnectionReadBuffer  = 1 << 10

	MessageTypeText MessageType = iota + 1
	MessageTypeBinary
)

type Option struct {
	Buffer                int         // Buffer the data that need to send
	MessageType           MessageType // Message type
	ConnectionWriteBuffer int         // connection write buffer
	ConnectionReadBuffer  int         // connection read buffer
}

func DefaultOption ()*Option{
	return &Option{
		Buffer:                Buffer,
		MessageType:           MessageTypeBinary,
		ConnectionWriteBuffer: ConnectionWriteBuffer,
		ConnectionReadBuffer:  ConnectionReadBuffer,
	}
}

var userOption *Option = DefaultOption()

func SetOption (option *Option) {
	userOption =option
}