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
	"sim/pkg/conn"
	"sim/pkg/logging"
)

const (
	DefaultClientHeartBeatInterval    = 120
	DefaultBucketSize                 = 1 << 9 // 512
	DefaultServerBucketNumber         = 1 << 4 // 16
	DefaultBucketBuffer               = 1 << 5 // 32
	DefaultBucketSendMessageGoroutine = 1 << 2 // 4

	// to show some pprof
	DefaultPProfPort = ":6060"
)

const (
	OfflineBySqueezeOut = iota + 1
	OfflineByLogic
)

type Options struct {
	ClientHeartBeatInterval    int // ClientHeartBeatInterval
	Connection                 *conn.Option
	BucketSize                 int           // BucketSize bucket size
	BucketBuffer               int           // BucketBuffer the buffer of cache bucket data , it will lose data when service dead
	BucketSendMessageGoroutine int           // BucketSendMessageGoroutine bucket goroutine witch use to send data
	ServerBucketNumber         int           // ServerBucketNumber
	LogPath                    string        // LogPath
	LogLevel                   logging.Level // LogLevel
	PProfPort                  string        // PProfPort

	// when user Offline by some reason  , you must to know that , so there may have some operate
	// you need to do , so you can implement this function , but i suggest you don't
	Offline func(conn conn.Connect, ty int)

	// ====================================== Option for hard code ===============================
	ServerDiscover Discover // ServerDiscover
	debug          bool
}

type Discover interface {
	Register()
	Deregister()
}

func DefaultOption() *Options {
	return &Options{
		// client
		ClientHeartBeatInterval: DefaultClientHeartBeatInterval,
		Connection:              conn.DefaultOption(),
		// server
		BucketSize:                 DefaultBucketSize,
		BucketBuffer:               DefaultBucketBuffer,
		BucketSendMessageGoroutine: DefaultBucketSendMessageGoroutine,
		ServerBucketNumber:         DefaultServerBucketNumber,
		PProfPort:                  DefaultPProfPort,

		debug: false,
	}
}

func LoadOptions(hooker Hooker, Opt ...OptionFunc) *Options {
	opt := DefaultOption()
	for _, o := range Opt {
		o(opt)
	}
	return opt
}

type OptionFunc func(b *Options)

func WithServerDebug() OptionFunc {
	return func(b *Options) {
		b.debug = true
	}
}

func WithPprofPort(pprof string) OptionFunc {
	return func(b *Options) {
		b.PProfPort = pprof
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

func WithConnectionOption(option *conn.Option) OptionFunc {
	return func(b *Options) {
		b.Connection = option
	}
}

func WithBucketSize(BucketSize int) OptionFunc {
	return func(b *Options) {
		b.BucketSize = BucketSize
	}
}

func WithBucketBuffer(BucketBuffer int) OptionFunc {
	return func(b *Options) {
		b.BucketBuffer = BucketBuffer
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
