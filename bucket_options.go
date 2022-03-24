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

import "context"

const (
	DefaultHeartBeatInterval = 10
	DefaultReaderBufferSize  = 10
	DefaultWriteBufferSize   = 10
	DefaultClientBufferSize  = 10  // 用户的buffersize
	DefaultMessageType       = 10
	DefaultProtocol          = 10
	DefaultBucketSize        = 100
)

type Option struct {
	HeartBeatInterval int64
	ReaderBufferSize  int
	WriteBufferSize   int
	ClientBufferSize  int
	MessageType       int
	Protocol          int
	BucketSize        int

	ctx      context.Context
	callback func()
}

func DefaultOption() *Option {
	return &Option{
		HeartBeatInterval: DefaultHeartBeatInterval,
		ReaderBufferSize:  DefaultReaderBufferSize,
		WriteBufferSize:   DefaultWriteBufferSize,
		ClientBufferSize:  DefaultClientBufferSize,
		MessageType:       DefaultMessageType,
		Protocol:          DefaultProtocol,
		BucketSize:        DefaultBucketSize,

	}
}

func NewOption(optset ...OptionSet) *Option {
	opt := DefaultOption()
	for _, o := range optset {
		o(opt)
	}
	return opt
}

type OptionSet func(option *Option)

func WithHeartBeatInterval(HeartBeatInterval int64) OptionSet {
	return func(option *Option) {
		option.HeartBeatInterval = HeartBeatInterval
	}
}

func WithReaderBufferSize(ReaderBufferSize int) OptionSet {
	return func(option *Option) {
		option.ReaderBufferSize = ReaderBufferSize
	}
}

func WithWriteBufferSize(WriteBufferSize int) OptionSet {
	return func(option *Option) {
		option.WriteBufferSize = WriteBufferSize
	}
}

func WithClientBufferSize(ClientBufferSize int) OptionSet {
	return func(option *Option) {
		option.ClientBufferSize = ClientBufferSize
	}
}

func WithProtocol(Protocol int) OptionSet {
	return func(option *Option) {
		option.Protocol = Protocol
	}
}

func WithMessageType(MessageType int) OptionSet {
	return func(option *Option) {
		option.MessageType = MessageType
	}
}

func WithContext(ctx context.Context) OptionSet {
	return func(h *Option) {
		h.ctx = ctx
	}
}

func WithCallBack(callback func()) OptionSet {
	return func(h *Option) {
		h.callback = callback
	}
}
