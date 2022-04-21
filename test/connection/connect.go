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

package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
)

func main() {
	config := InitConfig()
	RunServer(config)
	time.Sleep(time.Duration(config.keepTime) * time.Second)
}

// RandString 生成随机字符串做Token, 注意的是这里生成token的规则，
// 需要你能够在validate的接口实现中自己能解出来
func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

// Tokens 获取tokens
func Tokens(numbers int) []string {
	var tokens []string
	for i := 0; i < numbers; i++ {
		tokens = append(tokens, RandString(20))
	}
	return tokens
}

// RunServer 启动服务
func RunServer(cof *config) {
	tokens := Tokens(cof.keepTime)
	for k, v := range tokens {
		if k%cof.concurrency == 0 {
			time.Sleep(1 * time.Second)
		}
		go CreateMockClient(cof.host, v,cof.tagNum ,k )
	}
}



// CreateMockClient 图形界面化也可以使用这个网站进行查看 http://www.baidu.com/conn?token=1080&version=v.10
// 模拟连接，在此包内可
func CreateMockClient(Host, token string, tagN ,id int ) error {
	dialer := websocket.Dialer{}
	url := fmt.Sprintf(Host+"?token=%s&tag=%s", token,createTags(tagN))
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		fmt.Printf("error occurs during runtime id : %v, url : %s ,err :%s\r\n",id ,url,err.Error())
		return nil
	}
	defer conn.Close()
	for {
		messageType, messageData, err := conn.ReadMessage()
		if nil != err {
			return err
		}
		switch messageType {
		case websocket.TextMessage:
			fmt.Printf("content : %v \r\n", string(messageData))
		case websocket.BinaryMessage:
		default:
		}
	}
}
