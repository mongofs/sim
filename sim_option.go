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

const (
	// DefaultClientHeartBeatInterval 对客户端进行默认参数设置
	DefaultClientHeartBeatInterval = 120
	// DefaultClientReaderBufferSize tcp客户端的写入buffer设置
	DefaultClientReaderBufferSize = 1024
	// DefaultClientWriteBufferSize  tcp客户端的buffer 设置
	DefaultClientWriteBufferSize = 1024
	// DefaultClientBufferSize 每个用户缓存队列，每次消息写入是写入每个用户的待发送队列，
	// 队列对象是字节数组
	DefaultClientBufferSize = 8
	// DefaultClientMessageType 默认客户收到的消息类型,单个连接可以单独设置，创建连接的时候
	// 有相应的方法
	DefaultClientMessageType = 1
	// DefaultClientProtocol 默认的客户端的交互协议，目前支持protoc，和json
	DefaultClientProtocol = 1

	// DefaultBucketSize 对分片进行基础设置，默认给多少个分片，如果所有用户都在一个分片上
	// 那么容易导致分片上的锁竞争严重
	DefaultBucketSize = 1 << 8 // 256

	// DefaultServerBucketNumber  每个bucket 可以存放的用户数量，当用户数量上来了，会被扩容，
	// 这个参数只是初始值，可以结合公司体量，如果维护在线用户数，没有大量下推压力，可以将bucket设置大
	DefaultServerBucketNumber = 1 << 6 // 64

	// DefaultServerRpcPort 默认的RPC监听端口
	DefaultServerRpcPort = ":8081"
	// DefaultServerHttpPort 默认HTTP监听端口
	DefaultServerHttpPort = ":8080"

	// DefaultBroadCastHandler 设置对广播的单独设置handler ，对广播处理是将消息发送到不同的bucket上
	DefaultBroadCastHandler = 10
	// DefaultBroadCastBuffer 对广播消息单独设置缓冲区
	DefaultBroadCastBuffer = 200

	// PluginWTISupport 的参数支持
	PluginWTISupport = false // 是否支持WTI 进行扩展
)

type Option struct {
	// client
	ClientHeartBeatInterval int // 用户心跳间隔
	ClientReaderBufferSize  int // 用户连接读取buffer
	ClientWriteBufferSize   int // 用户连接写入buffer
	ClientBufferSize        int // 用户应用层buffer
	ClientMessageType       int // 用户发送的数据类型
	ClientProtocol          int // 压缩协议

	// bucket
	BucketSize int // bucket用户

	// server
	ServerBucketNumber int // 所有
	ServerRpcPort      string
	ServerHttpPort     string
	ServerValidate     Validate
	ServerReceive      Receive

	//broadcast
	BroadCastBuffer  int
	BroadCastHandler int

	//plugins
	SupportPluginWTI bool // 是否支持wti插件
}

func DefaultOption() *Option {
	return &Option{
		// client
		ClientHeartBeatInterval: DefaultClientHeartBeatInterval,
		ClientReaderBufferSize:  DefaultClientReaderBufferSize,
		ClientWriteBufferSize:   DefaultClientWriteBufferSize,
		ClientBufferSize:        DefaultClientBufferSize,
		ClientMessageType:       DefaultClientMessageType,
		ClientProtocol:          DefaultClientProtocol,
		// bucket
		BucketSize: 			 DefaultBucketSize,
		// server
		ServerBucketNumber: 	 DefaultServerBucketNumber,
		ServerRpcPort:      	 DefaultServerRpcPort,
		ServerHttpPort:     	 DefaultServerHttpPort,
		// broadCast
		BroadCastBuffer:  	 	 DefaultBroadCastBuffer,
		BroadCastHandler: 		 DefaultBroadCastHandler,
		// 插件支持
		SupportPluginWTI: 		 PluginWTISupport,
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

func WithServerHttpPort(ServerHttpPort string) OptionFunc {
	return func(b *Option) {
		b.ServerHttpPort = ServerHttpPort
	}
}

func WithServerRpcPort(ServerRpcPort string) OptionFunc {
	return func(b *Option) {
		b.ServerRpcPort = ServerRpcPort
	}
}

func WithServerValidate(ServerValidate Validate) OptionFunc {
	return func(b *Option) {
		b.ServerValidate = ServerValidate
	}
}

func WithServerBucketNumber(ServerBucketNumber int) OptionFunc {
	return func(b *Option) {
		b.ServerBucketNumber = ServerBucketNumber
	}
}

func WithServerReceive(ServerReceive Receive) OptionFunc {
	return func(b *Option) {
		b.ServerReceive = ServerReceive
	}
}

func WithClientHeartBeatInterval(ClientHeartBeatInterval int) OptionFunc {
	return func(b *Option) {
		b.ClientHeartBeatInterval = ClientHeartBeatInterval
	}
}

func WithClientReaderBufferSize(ClientReaderBufferSize int) OptionFunc {
	return func(b *Option) {
		b.ClientReaderBufferSize = ClientReaderBufferSize
	}
}

func WithClientWriteBufferSize(ClientWriteBufferSize int) OptionFunc {
	return func(b *Option) {
		b.ClientWriteBufferSize = ClientWriteBufferSize
	}
}

func WithClientBufferSize(ClientBufferSize int) OptionFunc {
	return func(b *Option) {
		b.ClientBufferSize = ClientBufferSize
	}
}

func WithClientMessageType(ClientMessageType int) OptionFunc {
	return func(b *Option) {
		b.ClientMessageType = ClientMessageType
	}
}

func WithClientProtocol(ClientProtocol int) OptionFunc {
	return func(b *Option) {
		b.ClientProtocol = ClientProtocol
	}
}

func WithBucketSize(BucketSize int) OptionFunc {
	return func(b *Option) {
		b.BucketSize = BucketSize
	}
}

func WithBroadCastBuffer(BroadCastBuffer int) OptionFunc {
	return func(b *Option) {
		b.BroadCastBuffer = BroadCastBuffer
	}
}

func WithBroadCastHandler(BroadCastHandler int) OptionFunc {
	return func(b *Option) {
		b.BroadCastHandler = BroadCastHandler
	}
}

//设置plugin内容
func WithPluginsWTI(SupportPluginWTI bool) OptionFunc {
	return func(b *Option) {
		if SupportPluginWTI {
			SetSupport()
		}
		b.SupportPluginWTI = SupportPluginWTI
	}
}
