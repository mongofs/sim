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

//  Connect 目前暂时支持 "github.com/gorilla/websocket" ，后续笔者打算编写一个基于epoll+EL的一个
//  网络模型，目前初步名称称为snetpoll，将websocket支持该模型，不过当下来说gorilla的包还是一个非常不错
//  的选择，所以目前所有调用都抽象出来，后续可能增加底层扩展支持，目前只用到gorilla的基础方法,后续增加或者
//  切换底层支持的话会重新发版，也考虑通过参数控制让用户自行选择实现

type Connect interface {

	// the connection
	Identification() string

	Send(data []byte) error

	Close(reason string)

	ReFlushHeartBeatTime()

	GetLastHeartBeatTime() int64

}
