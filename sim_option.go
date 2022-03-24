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

import "log"

const (
	// 对客户端进行默认参数设置
	DefaultSIMClientHeartBeatInterval = 120
	DefaultSIMClientReaderBufferSize  = 1024
	DefaultSIMClientWriteBufferSize   = 1024
	DefaultSIMClientBufferSize        = 8
	DefaultSIMClientMessageType       = 1
	DefaultSIMClientProtocol          = 1

	// 对分片进行基础设置
	DefaultSIMBucketSize = 1 << 8 // 256

	// 默认基础的server配置
	DefaultSIMServerBucketNumber = 1 << 6 // 64
	DefaultSIMServerRpcPort      = ":8081"
	DefaultSIMServerHttpPort     = ":8080"

	// 设置对广播能力的参数支持
	DefaultSIMBroadCastHandler = 10
	DefaultSIMBroadCastBuffer  = 200

	// plugins 的参数支持
	PluginWTISupport = false // 是否支持WTI 进行扩展
)

var DefaultValidate Validater = &DefaultValidate{}
var DefaultReceive client.Receiver = &client.Example{}
var DefaultLogger log.Logger = &log.DefaultLog{}

type SIMOption struct {
	// client
	ClientSIMHeartBeatInterval int // 用户心跳间隔
	ClientSIMReaderBufferSize  int // 用户连接读取buffer
	ClientSIMWriteBufferSize   int // 用户连接写入buffer
	ClientSIMBufferSize        int // 用户应用层buffer
	ClientSIMMessageType       int // 用户发送的数据类型
	ClientSIMProtocol          int // 压缩协议

	// bucket
	BucketSize         int // bucket用户

	// server
	ServerBucketNumber int // 所有
	ServerRpcPort      string
	ServerHttpPort     string
	ServerValidate     Validater
	ServerReceive      Receive
	ServerLogger       log.Logger

	//broadcast
	BroadCastBuffer  int
	BroadCastHandler int

	//plugins
	SupportPluginWTI bool // 是否支持wti插件
}

func DefaultOption() *SIMOption {
	return &SIMOption{
		ClientHeartBeatInterval: DefaultSIMClientHeartBeatInterval,
		ClientReaderBufferSize:  DefaultSIMClientReaderBufferSize,
		ClientWriteBufferSize:   DefaultSIMClientWriteBufferSize,
		ClientBufferSize:        DefaultSIMClientBufferSize,
		ClientMessageType:       DefaultSIMClientMessageType,
		ClientProtocol:          DefaultSIMClientProtocol,
		BucketSize:              DefaultSIMBucketSize,

		ServerBucketNumber: DefaultSIMServerBucketNumber, // 所有
		ServerRpcPort:      DefaultSIMServerRpcPort,
		ServerHttpPort:     DefaultSIMServerHttpPort,
		ServerValidate:     DefaultValidate,
		ServerReceive:      DefaultReceive,
		ServerLogger:       DefaultLogger,

		BroadCastBuffer:  DefaultSIMBroadCastBuffer,
		BroadCastHandler: DefaultSIMBroadCastHandler,

		// 插件支持
		SupportPluginWTI: PluginWTISupport,
	}
}

func NewOption(Opt ...OptionFunc) *Option {
	opt := DefaultOption()
	for _, o := range Opt {
		o(opt)
	}
	return opt
}

type OptionFunc func(b *Option)

func WithSIMServerHttpPort(ServerHttpPort string) OptionFunc {
	return func(b *Option) {
		b.ServerHttpPort = ServerHttpPort
	}
}

func WithSIMServerRpcPort(ServerRpcPort string) OptionFunc {
	return func(b *Option) {
		b.ServerRpcPort = ServerRpcPort
	}
}

func WithSIMServerValidate(ServerValidate Validater) OptionFunc {
	return func(b *Option) {
		b.ServerValidate = ServerValidate
	}
}

func WithSIMServerLogger(ServerLogger log.Logger) OptionFunc {
	return func(b *Option) {
		b.ServerLogger = ServerLogger
	}
}

func WithSIMServerBucketNumber(ServerBucketNumber int) OptionFunc {
	return func(b *Option) {
		b.ServerBucketNumber = ServerBucketNumber
	}
}

func WithSIMServerReceive(ServerReceive client.Receiver) OptionFunc {
	return func(b *Option) {
		b.ServerReceive = ServerReceive
	}
}

func WithSIMClientHeartBeatInterval(ClientHeartBeatInterval int) OptionFunc {
	return func(b *Option) {
		b.ClientHeartBeatInterval = ClientHeartBeatInterval
	}
}

func WithSIMClientReaderBufferSize(ClientReaderBufferSize int) OptionFunc {
	return func(b *Option) {
		b.ClientReaderBufferSize = ClientReaderBufferSize
	}
}

func WithSIMClientWriteBufferSize(ClientWriteBufferSize int) OptionFunc {
	return func(b *Option) {
		b.ClientWriteBufferSize = ClientWriteBufferSize
	}
}

func WithSIMClientBufferSize(ClientBufferSize int) OptionFunc {
	return func(b *Option) {
		b.ClientBufferSize = ClientBufferSize
	}
}

func WithSIMClientMessageType(ClientMessageType int) OptionFunc {
	return func(b *Option) {
		b.ClientMessageType = ClientMessageType
	}
}

func WithSIMClientProtocol(ClientProtocol int) OptionFunc {
	return func(b *Option) {
		b.ClientProtocol = ClientProtocol
	}
}

func WithSIMBucketSize(BucketSize int) OptionFunc {
	return func(b *Option) {
		b.BucketSize = BucketSize
	}
}

func WithSIMBroadCastBuffer(BroadCastBuffer int) OptionFunc {
	return func(b *Option) {
		b.BroadCastBuffer = BroadCastBuffer
	}
}

func WithSIMBroadCastHandler(BroadCastHandler int) OptionFunc {
	return func(b *Option) {
		b.BroadCastHandler = BroadCastHandler
	}
}

//设置plugin内容
func WithSIMPluginsWTI(SupportPluginWTI bool) OptionFunc {
	return func(b *Option) {
		if SupportPluginWTI {
			wti.SetSupport()
		}
		b.SupportPluginWTI = SupportPluginWTI
	}
}
