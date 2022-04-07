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

import "sim/pkg/logging"

const (
	// ====================================== Options for only client-side =======================

	DefaultClientHeartBeatInterval = 120
	DefaultClientReaderBufferSize  = 1024
	DefaultClientWriteBufferSize   = 1024
	DefaultClientBufferSize        = 8
	DefaultClientMessageType       = 1
	DefaultClientProtocol          = 1

	// ====================================== Options for only server-side =======================
	DefaultBucketSize         = 1 << 8 // 256
	DefaultServerBucketNumber = 1 << 6 // 64
	DefaultServerRpcPort      = ":8081"
	DefaultServerHttpPort     = ":8080"
	DefaultBroadCastHandler   = 10
	DefaultBroadCastBuffer    = 200

	// PluginWTISupport 的参数支持
	PluginWTISupport = false // 是否支持WTI 进行扩展
)

type Options struct {
	// ====================================== Options for only client-side =======================

	// ClientHeartBeatInterval 用户的心跳间隔时间
	ClientHeartBeatInterval int

	// ClientReaderBufferSize 用户连接读取buffer
	ClientReaderBufferSize int

	// ClientWriteBufferSize 用户连接写入buffer
	ClientWriteBufferSize int

	// ClientBufferSize 用户应用层buffer
	ClientBufferSize int

	//ClientMessageType  用户发送的数据类型
	ClientMessageType MessageType

	// ClientProtocol 压缩协议,这个后期需要舍弃
	ClientProtocol Protocol

	// ====================================== Options for only server-side =======================

	// BucketSize 每个bucket初始值，如果有预估可以减少map后期扩容带来性能开销
	BucketSize int

	// ServerBucketNumber bucket的总数量，预计单机分成多少个bucket
	ServerBucketNumber int

	// ServerRpcPort
	ServerRpcPort string

	// ServerHttpPort
	ServerHttpPort string

	// ServerValidate 服务的validate
	ServerValidate Validate

	// ServerReceive 当某个client收到信息后进行处理
	ServerReceive Receive

	// ServerDiscover 进行服务的发现注册，支持多部署能力
	ServerDiscover Discover

	//BroadCastBuffer 广播缓存的大小
	BroadCastBuffer int

	BroadCastHandler int

	//plugins
	SupportPluginWTI bool // 是否支持wti插件

	Logger logging.Logger

	LogPath string

	LogLevel logging.Level
}

func DefaultOption() *Options {
	return &Options{
		// client
		ClientHeartBeatInterval: DefaultClientHeartBeatInterval,
		ClientReaderBufferSize:  DefaultClientReaderBufferSize,
		ClientWriteBufferSize:   DefaultClientWriteBufferSize,
		ClientBufferSize:        DefaultClientBufferSize,
		ClientMessageType:       DefaultClientMessageType,
		ClientProtocol:          DefaultClientProtocol,
		// server
		BucketSize:         DefaultBucketSize,
		ServerBucketNumber: DefaultServerBucketNumber,
		ServerRpcPort:      DefaultServerRpcPort,
		ServerHttpPort:     DefaultServerHttpPort,
		BroadCastBuffer:    DefaultBroadCastBuffer,
		BroadCastHandler:   DefaultBroadCastHandler,
		SupportPluginWTI:   PluginWTISupport,
	}
}

func LoadOptions(validate Validate, receive Receive, Opt ...OptionFunc) *Options {
	opt := DefaultOption()
	opt.ServerValidate = validate
	opt.ServerReceive = receive
	for _, o := range Opt {
		o(opt)
	}
	return opt
}

type OptionFunc func(b *Options)

func WithServerHttpPort(ServerHttpPort string) OptionFunc {
	return func(b *Options) {
		b.ServerHttpPort = ServerHttpPort
	}
}

func WithServerRpcPort(ServerRpcPort string) OptionFunc {
	return func(b *Options) {
		b.ServerRpcPort = ServerRpcPort
	}
}

func WithServerBucketNumber(ServerBucketNumber int) OptionFunc {
	return func(b *Options) {
		b.ServerBucketNumber = ServerBucketNumber
	}
}

func WithClientHeartBeatInterval(ClientHeartBeatInterval int) OptionFunc {
	return func(b *Options) {
		b.ClientHeartBeatInterval = ClientHeartBeatInterval
	}
}

func WithClientReaderBufferSize(ClientReaderBufferSize int) OptionFunc {
	return func(b *Options) {
		b.ClientReaderBufferSize = ClientReaderBufferSize
	}
}

func WithClientWriteBufferSize(ClientWriteBufferSize int) OptionFunc {
	return func(b *Options) {
		b.ClientWriteBufferSize = ClientWriteBufferSize
	}
}

func WithClientBufferSize(ClientBufferSize int) OptionFunc {
	return func(b *Options) {
		b.ClientBufferSize = ClientBufferSize
	}
}

func WithClientMessageType(ClientMessageType MessageType) OptionFunc {
	return func(b *Options) {
		b.ClientMessageType = ClientMessageType
	}
}

func WithClientProtocol(ClientProtocol Protocol) OptionFunc {
	return func(b *Options) {
		b.ClientProtocol = ClientProtocol
	}
}

func WithBucketSize(BucketSize int) OptionFunc {
	return func(b *Options) {
		b.BucketSize = BucketSize
	}
}

func WithBroadCastBuffer(BroadCastBuffer int) OptionFunc {
	return func(b *Options) {
		b.BroadCastBuffer = BroadCastBuffer
	}
}

func WithBroadCastHandler(BroadCastHandler int) OptionFunc {
	return func(b *Options) {
		b.BroadCastHandler = BroadCastHandler
	}
}

//设置plugin内容
func WithPluginsWTI(SupportPluginWTI bool) OptionFunc {
	return func(b *Options) {
		if SupportPluginWTI {
			SetSupport()
		}
		b.SupportPluginWTI = SupportPluginWTI
	}
}

// WithLogger sets up a customized logger.
func WithLogger(logger logging.Logger) OptionFunc {
	return func(opts *Options) {
		opts.Logger = logger
	}
}



func WithDiscover(discover Discover) OptionFunc {
	return func(opts *Options) {
		opts.ServerDiscover = discover
	}
}
