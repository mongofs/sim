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

import "net/http"

type Client interface {

	// Send
	Send([]byte, ...int64) error

	// Offline
	Offline()

	// ResetHeartBeatTime 重置用户的心跳
	ResetHeartBeatTime()

	// LastHeartBeat 获取用户的最后一次心跳
	LastHeartBeat() int64

	// Token 获取用户的token
	Token() string

	// Request 获取到用户的请求的链接
	Request() *http.Request

	// SetMessageType 设置用户接收消息的格式
	SetMessageType(int)

	// SetProtocol 设置用户接收消息的协议：
	SetProtocol(int)
}

const (
	waitTime = 1 << 7

	ProtocolJson     = 1
	ProtocolProtobuf = 2

	MessageTypeText   = 1
	MessageTypeBinary = 2
)
