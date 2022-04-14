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
	"sync"
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

	// Token 获取用户的token
	Token() string

	// Request 获取到用户的请求的链接
	Request() *http.Request

	// SetMessageType 设置用户接收消息的格式
	SetMessageType(messageType MessageType)

	// SetProtocol 设置用户接收消息的协议：
	SetProtocol(protocol Protocol)

	//============================================ Tag ===============================

	// HaveTag 判断用户是否存在某个tag
	HaveTag(tags [] string) bool

	// SetTag 为用户添加tag
	SetTag(tags []string)error

	// DelTag 删除用户的tag
	DelTag(tags [] string)
}

type Cli struct {
	// protect Cli tag
	rw sync.RWMutex
	Connect
	tags   map[string]*target
	reader *http.Request
}



func NewClient(w http.ResponseWriter, r *http.Request, closeSig chan<- string, token *string, option *Options) (Client, error) {
	res := &Cli{
		reader: r,
	}

	conn, err := NewGorilla(token, closeSig, option, w, r, option.ServerReceive)
	if err != nil {
		return nil, err
	}
	res.Connect = conn
	return res, nil
}

func (c *Cli) Request() *http.Request {
	return c.reader
}

func (c *Cli) HaveTag(tags []string) bool {
	c.rw.Lock()
	defer c.rw.RUnlock()
	for _,tag:= range tags{
		if _, ok := c.tags[tag]; !ok {
			return false
		}
	}
	return true
}

func (c *Cli) SetTag(tags []string) error{
	if len(tags) == 0 {return errors.New("")}
	c.rw.Lock()
	defer c.rw.RUnlock()

	// 查找对应的tags，如果不存在就创建
/*	targets := factoryWTI.Find(tags)
	for _,target := range targets {
		target.Add(c)
		c.tags[target.name] =target
	}*/
	return nil
}

func (c *Cli) DelTag(tags [] string) {
	c.rw.Lock()
	defer c.rw.RUnlock()
	for _,tag := range tags{
		if _,ok := c.tags[tag];ok{
			//target.Del(c.Token())
			delete(c.tags, tag)
		}
	}
}
