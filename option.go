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
	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
	"net/http"
)

const (
	DefaultClientHeartBeatInterval = 120
	DefaultBucketSize              = 1 << 8 // 256
	DefaultServerBucketNumber      = 1 << 5 // 32
)

type Options struct {
	ClientHeartBeatInterval int // ClientHeartBeatInterval 用户的心跳间隔时间
	Connection              *conn.Option
	BucketSize              int            // BucketSize 每个bucket初始值
	ServerBucketNumber      int            // ServerBucketNumber 预计单机分成多少个bucket
	Logger                  logging.Logger // Logger 设置logger
	LogPath                 string         // LogPath 设置logger path
	LogLevel                logging.Level  // LogLevel 设置logger level

	// ====================================== Option for hard code ===============================
	ServerDiscover Discover // ServerDiscover 进行服务的发现注册，支持多部署能力
	hooker         Hooker   // 是必须设置的code函数
}

// Discover 可以在服务启动停止的时候自动想注册中心进行注册和注销，这个实现是可选的，具体可
// 查看option的参数，如果没有discover 就是一个单节点，也是可以启动。但是建议你在生产环境
// 使用的时候还是以集群方式启动，一旦存在集群的方式就必须注册这个方法，就可以将所有的内容作为
// 组件方式进行使用，可以通过独立网关进行sim地址下发。
type Discover interface {
	Register()
	Deregister()
}

type Hooker interface {
	// Validate 验证器，用户进行登录的时候需要进行验证，调用层需要注册Validate对象，整体流程
	// 是进行ValidateKey 验证，验证方法也是服务注册的时候实现validate方法，具体见connection
	// 方法，调用validate方法后可能会成功，可能会失败，不论成功或者失败都需要发送业务层的内容
	// 所以需要服务注册时候实现validate整个接口
	Validate(token string) error
	ValidateFailed(err error, cli conn.Connect)
	ValidateSuccess(cli conn.Connect)

	// Receive 是需要用户进行注册，主要是为了接管用户上传的消息内容，在消息处理的时候可以根据
	// 自身的业务需求进行处理
	HandleReceive(conn conn.Connect, data []byte)

	// IdentificationHook 这里是注册通过连接获取连接唯一标识的绑定关系，第一个参数表示返回具体的
	// 标识，第二个参数是具体的错误,当第二个错误出现，创建连接的动作将不再继续
	IdentificationHook(w http.ResponseWriter, r *http.Request) (string, error)
}

func DefaultOption() *Options {
	return &Options{
		// client
		ClientHeartBeatInterval: DefaultClientHeartBeatInterval,
		Connection:              conn.DefaultOption(),
		// server
		BucketSize:         DefaultBucketSize,
		ServerBucketNumber: DefaultServerBucketNumber,
	}
}

func LoadOptions(hooker Hooker, Opt ...OptionFunc) *Options {
	opt := DefaultOption()
	opt.hooker = hooker
	for _, o := range Opt {
		o(opt)
	}
	return opt
}

type OptionFunc func(b *Options)

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

func WithConnectionOption(option *conn.Option) OptionFunc {
	return func(b *Options) {
		b.Connection= option
	}
}

func WithBucketSize(BucketSize int) OptionFunc {
	return func(b *Options) {
		b.BucketSize = BucketSize
	}
}

func WithLogger(logger logging.Logger) OptionFunc {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

func WithLoggerLevel(level logging.Level) OptionFunc {
	return func(b *Options) {
		b.LogLevel = level
	}
}

func WithDiscover(discover Discover) OptionFunc {
	return func(opts *Options) {
		opts.ServerDiscover = discover
	}
}
