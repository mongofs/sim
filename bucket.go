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
	"errors"
	"net/http"
)

type Bucket interface {

	// Send can send
	// send data to someone
	// ACK indicates whether the message is a receipt message
	Send(data []byte, token string ,Ack bool)error

	// BroadCast Send messages to all online users
	BroadCast(data []byte ,Ack bool)error

	// OffLine Kick users offline
	OffLine(token string)

	// Register user to basket
	Register(cli Client,token string)error

	//IsOnline Judge whether the user is online
	IsOnline(token string)bool


	Online()int64


	Flush()


	NotifyBucketConnectionIsClosed()chan <- string


	CreateConn(w http.ResponseWriter,r * http.Request,token string)(Client,error)
}


var (
	ErrUserExist =errors.New("im/bucket : Cannot login repeatedly")
	ErrCliISNil  =errors.New("im/bucket : cli is nil")
)