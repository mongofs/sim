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

	//============================================ Tag ===============================

	// HaveTag 判断用户是否存在某个tag
	HaveTag(tags []string) bool

	// SetTag 为用户添加tag
	SetTag(tag string) error

	// DelTag 删除用户的tag
	DelTag(tag string)

	// RangeTag 遍历所有tag
	RangeTag() (res []string)
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
	c.rw.RLock()
	defer c.rw.RUnlock()
	for _, tag := range tags {
		if _, ok := c.tags[tag]; !ok {
			return false
		}
	}
	return true
}

func (c *Cli) SetTag(tag string) error {
	c.rw.Lock()
	defer c.rw.Unlock()
	tgAd, err := WTIAdd(tag, c)
	if err != nil {
		return err
	}
	c.tags[tag] = tgAd
	return nil
}

func (c *Cli) DelTag(tag string) {
	c.rw.Lock()
	defer c.rw.Unlock()
	if tar, ok := c.tags[tag]; ok {
		delete(c.tags, tag)
		tar.Del([]string{c.Token()})
	}
}

func (c *Cli) RangeTag() (res []string) {
	c.rw.RLock()
	defer c.rw.RUnlock()
	for k, _ := range c.tags {
		res = append(res, k)
	}
	return res
}
