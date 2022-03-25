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

	im "sim/api/v1"

	"github.com/zhenjl/cityhash"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

type sim struct {
	http *http.ServeMux
	rpc  *grpc.Server
	bs   []Bucket
	ps   atomic.Int64

	buffer chan []byte
	cancel func()
	opt    *Option
}

func New(opts *Option) *sim {
	b := &sim{
		ps:  atomic.Int64{},
		opt: opts,
	}
	b.buffer = make(chan []byte, opts.BroadCastBuffer)
	b.ps.Store(0)
	// 准备创建bucket
	b.prepareBucket()
	// 创建grpcServer
	b.prepareGrpcServer()
	// 创建httpServer
	b.prepareHttpServer()
	return b
}

// prepareBucket 构建bucket
func (h *sim) prepareBucket() {
	h.bs = make([]Bucket, h.opt.ServerBucketNumber)
	_, cancel := context.WithCancel(context.Background())
	h.cancel = cancel
	for i := 0; i < h.opt.ServerBucketNumber; i++ {
		h.bs[i] = newBucket(h.opt)
	}
}

// prepareGrpcServer 构建grpc 服务注册
func (b *sim) prepareGrpcServer() {
	b.rpc = grpc.NewServer()
	im.RegisterBasicServer(b.rpc, b)
}

// prepareHttpServer 构建http服务
func (b *sim) prepareHttpServer() {
	b.http = http.NewServeMux()
	b.initRouter()
}

func (s *sim) bucket(token string) Bucket {
	idx := Index(token, uint32(s.opt.ServerBucketNumber))
	return s.bs[idx]
}

func Index(token string, size uint32) uint32 {
	return cityhash.CityHash32([]byte(token), uint32(len(token))) % size
}

// Ping 发送ping消息
func (s *sim) Ping(ctx context.Context, empty *im.Empty) (*im.Empty, error) {
	return nil, nil
}

// Online 在线用户
func (s *sim) Online(ctx context.Context, empty *im.Empty) (*im.OnlineReply,
	error) {
	num := s.ps.Load()
	req := &im.OnlineReply{Number: num}
	return req, nil
}



// Broadcast 给所有在线用户广播
func (s *sim) Broadcast(ctx context.Context, req *im.BroadcastReq) (
	*im.BroadcastReply, error) {
	if len(s.buffer)*10 > 8*cap(s.buffer) {
		return nil,ErrUserBufferIsFull
	}
	s.buffer <- req.Data
	return &im.BroadcastReply{
		Size: int64(len(s.buffer)),
	}, nil
}

// WTIDistribute 获取每个版本多少人
func (s *sim) WTIDistribute(ctx context.Context, req *im.Empty) (
	*im.WTIDistributeReply, error) {
	distribute, err := Distribute()
	if err != nil {
		return nil, err
	}

	var result = map[string]*im.WTIDistribute{}
	for k, v := range distribute {
		data := &im.WTIDistribute{
			Tag:        v.TagName,
			Number:     v.Onlines,
			CreateTime: v.CreateTime,
		}
		result[k] = data
	}
	return &im.WTIDistributeReply{
		Data: result,
	}, nil
}

// WTIBroadcast 在开发过程中存在IM需要版本共存的需求，比如我的协议替换了，但是如果im应用在App上面如何
// 进行切换，这就是协议定制不合理的地方，但也需要.IM 服务器在这个过程中做配合。
// IM 存在给用户分组的需求，所以我们在进行Broadcast 就必须进行用户的状态区分，所以前台需要对内容进行分
// 组，传入的内容也需要对应分组比如 v1 => string ，v2 => []byte，那么v1，v2 就是不相同的两个版本内
// 容。在client上面可以设置用户的连接版本Version，建议在使用用户
func (s *sim) WTIBroadcast(ctx context.Context, req *im.BroadcastByWTIReq) (
	*im.BroadcastReply, error) {
	var err error
	err = BroadCastByTarget(req.Data)
	return &im.BroadcastReply{
		Size: int64(len(s.buffer)),
	}, err
}

// SendMessageToMultiple 发送消息给多个用户
func (s *sim) SendMessageToMultiple(ctx context.Context, req *im.SendMsgReq)(
	*im.Empty, error) {
	var err error
	for _, token := range req.Token {
		bs := s.bucket(token)
		err = bs.Send(req.Data, token, false)
	}
	return &im.Empty{}, err
}
