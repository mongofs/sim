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

type Client interface {

	//============================================ Connection ===============================

	// Send 发送消息下去
	Send([]byte) error

	// Close 关闭连接
	Close(forRetry bool) error

	// ReFlushHeartBeatTime 重置用户的心跳
	ReFlushHeartBeatTime()

	// GetLastHeartBeatTime 获取用户的最后一次心跳
	GetLastHeartBeatTime() int64

	// Identification 获取用户的token
	Identification() string

	// Request 获取到用户的请求的链接
	Request() *http.Request

	// SetMessageType 设置用户接收消息的格式
	SetMessageType(messageType MessageType)


	// -----------------------tag --------------------

	HaveTags(tags []string) bool

	SetTag(tag string) error

	DelTag(tag string)

	RangeTag() (res []string)

}



type Cli struct {
	Connect
}

func NewClient(w http.ResponseWriter, r *http.Request, closeSig chan<- string, token *string, option *Options) (Client, error) {
	res := &Cli{}
	conn, err := NewGorilla(token, closeSig, option, w, r, option.ServerReceive)
	if err != nil {
		return nil, err
	}
	res.Connect = conn
	return res, nil
}






