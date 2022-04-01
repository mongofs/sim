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
	"net/http"
)

type Bucket interface {

	// Send 发送消息给某个token，并且标识是否需要ACK回执
	Send(data []byte, token string, Ack bool) error

	// BroadCast 发送消息给这个bucket的全部用户，并且标识是否需要ACK回执
	BroadCast(data []byte, Ack bool) error

	// OffLine 下线某个用户
	OffLine(token string)

	// Register user to basket
	Register(cli Client, token string) error

	// Online 查看在线用户有多少个
	Online() int64


	CreateConn(w http.ResponseWriter, r *http.Request, token string) (Client, error)
}
