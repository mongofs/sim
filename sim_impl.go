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
	"context"
	"net/http"
)

type ImSrever struct {
	http           *http.ServeMux
	rpc            *grpc.Server
	bs             []bucket.Bucketer
	ps             atomic.Int64

	buffer chan *im.BroadcastReq // 全局广播队列
	cancel func()

	opt *Option
}


func New(opts *Option) *ImSrever {
	b := &ImSrever{
		ps:     atomic.Int64{},
		opt:    opts,
	}
	b.buffer = make(chan *grpc2.BroadcastReq,opts.BroadCastBuffer)
	b.ps.Store(0)
	b.prepareBucketer()
	b.prepareGrpcServer()
	b.prepareHttpServer()
	return b
}



func (h *ImSrever) prepareBucketer() {
	h.bs = make([]bucket.Bucketer, h.opt.ServerBucketNumber)
	_, cancel := context.WithCancel(context.Background())
	h.cancel = cancel

	BucketOptionSet := &bucket.Option{
		HeartBeatInterval: int64(h.opt.ClientHeartBeatInterval),
		ReaderBufferSize:  h.opt.ClientReaderBufferSize,
		WriteBufferSize:   h.opt.ClientWriteBufferSize,
		ClientBufferSize:  h.opt.ClientBufferSize,
		MessageType:       h.opt.ClientMessageType,
		Protocol:          h.opt.ClientProtocol,
		BucketSize:        h.opt.BucketSize,
	}

	for i:= 0 ;i<h.opt.ServerBucketNumber;i ++ {
		h.bs[i] = bucket.New(h.opt.ServerLogger,BucketOptionSet)
	}
}

func (b *ImSrever) prepareGrpcServer() {
	b.rpc = grpc.NewServer()
	grpc2.RegisterBasicServer(b.rpc, b)
}

func (b *ImSrever) prepareHttpServer() {
	b.http = http.NewServeMux()
	b.initRouter()
}


func (s *ImSrever) bucket(token string) bucket.Bucketer {
	idx := Index(token,uint32(s.opt.ServerBucketNumber))
	return s.bs[idx]
}