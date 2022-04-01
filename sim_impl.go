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
	"encoding/json"
	"net/http"
	"sim/pkg/logging"
	"time"

	"github.com/zhenjl/cityhash"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	im "sim/api/v1"
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

func (s *sim) initRouter() error {
	s.http.HandleFunc(RouterHealth, handlerHealth)
	s.http.HandleFunc(RouterConnection, s.Connection)
	return nil
}

//Connection  create  connection
func (s *sim) Connection(writer http.ResponseWriter, r *http.Request) {
	now := time.Now()
	defer func() {
		escape := time.Since(now)
		logging.Infof("sim : %s create %s  cost %v  url is %v ", r.RemoteAddr, r.Method, escape, r.URL)
	}()
	res := &Response{
		w:      writer,
		Data:   nil,
		Status: 200,
	}
	if r.ParseForm() != nil {
		res.Status = 400
		res.Data = "connection is bad "
		res.SendJson()
		return
	}

	token := r.Form.Get(ValidateKey)
	if token == "" {
		res.Status = 400
		res.Data = "token validate error"
		res.SendJson()
	}
	// validate token
	bs := s.bucket(token)
	cli, err := bs.CreateConn(writer, r, token)
	if err != nil {
		res.Status = 400
		res.Data = err.Error()
		return
	}
	if err := s.opt.ServerValidate.Validate(token); err != nil {
		s.opt.ServerValidate.ValidateFailed(err, cli)
		return
	} else {
		s.opt.ServerValidate.ValidateSuccess(cli)
	}
	// register to data
	if err := bs.Register(cli, token); err != nil {
		cli.Send([]byte(err.Error()))
		cli.Close(false)
	}
}

type Response struct {
	w      http.ResponseWriter
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func (r *Response) SendJson() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Status)
	return r.w.Write(resp)
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Status: 200,
		Data:   "ok",
	}
	res.SendJson()
}

func Index(token string, size uint32) uint32 {
	return cityhash.CityHash32([]byte(token), uint32(len(token))) % size
}
