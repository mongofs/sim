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
	"net/http"
	"testing"
)

type MockClient struct {
	token string
}

func (m MockClient) Send(bytes []byte) error {
	panic("implement me")
}

func (m MockClient) Close(forRetry bool) error {
	panic("implement me")
}

func (m MockClient) ReFlushHeartBeatTime() {
	panic("implement me")
}

func (m MockClient) GetLastHeartBeatTime() int64 {
	panic("implement me")
}

func (m MockClient) Token() string {
	panic("implement me")
}

func (m MockClient) Request() *http.Request {
	panic("implement me")
}

func (m MockClient) SetMessageType(messageType MessageType) {
	panic("implement me")
}

func (m MockClient) SetProtocol(protocol Protocol) {
	panic("implement me")
}

func NewTestClient(token string ) Client {
	return &MockClient{token: token}
}



func TestTg_DelTAG(t *testing.T) {
	SetSupport()
	tests := []struct{
		token string
		tag []string
	}{
		{
			token: "1234",
			tag: []string{"v1","v2","v3"},
		},
		{
			token: "12345",
			tag: []string{"v1","v2"},
		},
	}
	// 先设置全局变量
	for _,v := range tests{
		SetTAG(NewTestClient(v.token),v.tag...) // 将每个client 进行设置tag
	}
	dis,err := Distribute()
	if err != nil {
		t.Fatal(err)
	}
	for k,v:=range  dis{
		fmt.Println(k,v)
	}

	// 删除对应的tag
	for _,v := range tests{
		DelTAG(NewTestClient(v.token),v.tag[0])
	}
	dis,err = Distribute()
	if err != nil {
		t.Fatal(err)
	}
	for k,v:=range  dis{
		fmt.Println("after delete tag zero",k,v)
	}
}





func TestTg_SetTAG(t *testing.T) {
	tests := []struct{
		tag string
		number int
	}{
		{
			tag: "v1",
			number: 100,
		},
		{
			tag: "v2",
			number: 200,
		},
	}

	for _,v := range tests{
		for i :=0 ;i< v.number;i++ {
			SetTAG(NewTestClient("1234"),v.tag) //,todo 待优化
		}
	}
}

// 广播，针对tag进行广播，这也是wti的核心接口，分类广播也是基于这个接口
func TestTg_BroadCast(t *testing.T) {
	tests := []struct {
		target []string
		content []byte
	}{
		{
			target: []string{"v1"} ,
			content: []byte("hello content"),
		},
		{
			target: []string{"v1","v2"} ,
			content: []byte("hello content"),
		},
		{
			target: []string{"v1","v2","v3"} ,
			content: []byte("hello content"),
		},
	}
	// 测试 v1、v2 、v1 三个版本的不同，如果要模式真实连接，则需要执行
	// 1. 创建im 连接服务器，并开启wti 配置参数
	// 2. 要在handler方法中调用factory 进行调用SetTAG
	// 3. 需要建立连接
	for _,v := range tests {
		err := BroadCast(v.content,v.target...)
		if err !=nil {
			t.Fatal(err)
		}
	}
	// v1  websocket output : hello content ,and v2,v3 no content output
	// v1,v2 websocket output : hello content,and v3 no content output
	// v1,v2,v3 websocket output : hello content

}



// 主要应对数据发送的时候版本的问题，比如某一条数据由于协议更改需要向上兼容老的版本,因为这是应用层的内容
// 所以使用wti 接口来进行兼容处理。避免进行内容的感染
func TestTg_BroadCastByTarget(t *testing.T) {
	tests := []struct {
		give map[string][]byte
	}{
		{
			give: map[string][]byte{
				"v1": []byte("first v1 "),
				"v2": []byte("second v2 "),
				"v3": []byte("third v3 "),
			},

		},
		{
			give: map[string][]byte{
				"v1": []byte("hello v1 "),
				"v2": []byte("hello v2 "),
				"v3": []byte("hello v3 "),
			},
		},

	}
	// 测试 v1、v2 、v1 三个版本的不同，如果要模式真实连接，则需要执行
	// 1. 创建im 连接服务器，并开启wti 配置参数
	// 2. 要在handler方法中调用factory 进行调用SetTAG
	// 3. 需要建立连接
	for _,v := range tests {
		err:= BroadCastByTarget(v.give)
		if err !=nil {
			t.Fatal(err)
		}
	}
	// v1,v2,v3  websocket output :first v1 | second v2 | third v3
	// v1,v2,v3  websocket output :hello v1 | hello v2 | hello v3
}




// 主要应对数据发送的时候版本的问题，比如某一条数据由于协议更改需要向上兼容老的版本,因为这是应用层的内容
// 所以使用wti 接口来进行兼容处理。避免进行内容的感染
func TestTg_UpdateAndF(t *testing.T) {
	tests := []struct {
		give string
	}{
		{
			give: "1234",
		},
	}
	// 这个测试条件相对比较苛刻，update接口主要作用是接收globalclosed 的信号，如果某个用户关闭连接
	// im线程就会释放连接之前就会告诉当前的update方法，所以只需要判断当前用户是否被删除就好了
	// 1. 创建im 连接服务器，并开启wti 配置参数
	// 2. 要在handler方法中调用factory 进行调用SetTAG
	// 3. 需要建立连接
	for _,v := range tests {
		err := Update(v.give) //
		if err !=nil {
			t.Fatal(err)
		}
		res ,_:= GetClientTAGs(v.give)
		fmt.Println(res)
	}
	// output : []
}