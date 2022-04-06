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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
	Address = ""
)

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


func TestBucket_CreateConn(t *testing.T) {
	tests := []struct{
		tag string
		number int
	}{
		{
			tag: "v1",
			number: 50,
		},
		{
			tag: "v2",
			number: 60,
		},
	}
	for _,v := range tests{
		for i :=0 ;i< v.number;i++ {
			go CreateMockClient(v.tag,"")
		}
	}
	time.Sleep(1000 * time.Second)
}


// CreateClient 图形界面化也可以使用这个网站进行查看 http://www.baidu.com/conn?token=1080&version=v.10
// 模拟连接
func CreateMockClient(version ,token string) error{
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(fmt.Sprintf(Address+"?token=%s&version=%s", token,version), nil)
	if nil != err {
		return err
	}
	defer conn.Close()
	for {
		messageType, messageData, err := conn.ReadMessage()
		if nil != err {
			return err
		}
		switch messageType {
		case websocket.TextMessage:
			fmt.Printf( "content : %v \r\n",string(messageData))
		case websocket.BinaryMessage:
		default:
		}
	}
}