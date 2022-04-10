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
	"net"
	"net/http"
	"time"

	im "sim/api/v1"
	"sim/pkg/logging"

	"github.com/zhenjl/cityhash"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type sim struct {
	http *http.ServeMux
	rpc  *grpc.Server
	bs   []*bucket
	ps   atomic.Int64

	buffer chan []byte
	cancel func()
	opt    *Options
}

func (s *sim) bucket(token string) *bucket {
	idx := s.Index(token, uint32(s.opt.ServerBucketNumber))
	return s.bs[idx]
}

func (s *sim) Index(token string, size uint32) uint32 {
	return cityhash.CityHash32([]byte(token), uint32(len(token))) % size
}

func (s *sim) connection(writer http.ResponseWriter, r *http.Request) {
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

	token := r.Form.Get(s.opt.RouterValidateKey)
	if token == "" {
		res.Status = 400
		res.Data = "token validate error"
		res.SendJson()
	}
	// validate token
	bs := s.bucket(token)
	cli, err := bs.createConn(writer, r, token)
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

func (s *sim) handlerHealth(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Status: 200,
		Data:   "ok",
	}
	res.SendJson()
}

func (s *sim) monitorOnline() error {
	var interval = 10
	logging.Infof("sim : start monitor online server , interval is %v second",interval)
	for {
		n := int64(0)
		for _, bck := range s.bs {
			n += bck.Online()
		}
		s.ps.Store(n)
		time.Sleep(time.Duration(interval) * time.Second)
	}
	return nil
}

func (s *sim) monitorWTI() error {
	var interval = 20
	logging.Infof("sim : start monitor WIT  server , interval is %v second",interval)
	if s.opt.SupportPluginWTI {
		for {
			FlushWTI()
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
	return nil
}

func (s *sim) runGrpcServer() error {
	listen, err := net.Listen("tcp", s.opt.ServerRpcPort)
	if err != nil {
		return err
	}
	logging.Infof("sim : start GRPC server at %s ",s.opt.ServerRpcPort)
	if err := s.rpc.Serve(listen); err != nil {
		return err
	}
	return nil
}

func (s *sim) runHttpServer() error {
	listen, err := net.Listen("tcp", s.opt.ServerHttpPort)
	if err != nil {
		return err
	}
	logging.Infof("sim : start HTTP server at %s ", s.opt.ServerHttpPort)
	if err := http.Serve(listen, s.http); err != nil {
		return err
	}
	return nil
}

func (s *sim) handlerBroadCast() error {
	wg := errgroup.Group{}
	logging.Infof("sim : start handlerBroadCast ，number is %v  ", s.opt.BroadCastHandler)
	for i := 0; i < s.opt.BroadCastHandler; i++ {
		wg.Go(func() error {
			for {
				data := <-s.buffer
				for _, v := range s.bs {
					err := v.broadCast(data, false)
					if err != nil {
						logging.Error(err)
					}
				}
			}
			return nil
		})
	}
	return wg.Wait()
}

// todo
func (s *sim) handlerTargetBraodCastAsync ()error{
	wg := errgroup.Group{}
	logging.Infof("sim : start handlerBroadCast ，number is %v  ", s.opt.BroadCastHandler)
	for i := 0; i < s.opt.BroadCastHandler; i++ {
		wg.Go(func() error {
			for {
				data := <-s.buffer
				for _, v := range s.bs {
					err := v.broadCast(data, false)
					if err != nil {
						logging.Error(err)
					}
				}
			}
			return nil
		})
	}
	return wg.Wait()
}

func (s *sim) close() error {
	s.rpc.GracefulStop()
	s.cancel()
	return nil
}


func initSim(opts *Options) *sim {
	b := &sim{
		ps:  atomic.Int64{},
		opt: opts,
	}
	b.buffer = make(chan []byte, opts.BroadCastBuffer)
	b.ps.Store(0)

	// prepare buckets
	b.bs = make([]*bucket, b.opt.ServerBucketNumber)
	_, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	for i := 0; i < b.opt.ServerBucketNumber; i++ {
		b.bs[i] = newBucket(b.opt)
	}
	logging.Infof("sim : INIT_BUCKET_NUMBER is %v ",b.opt.ServerBucketNumber)
	logging.Infof("sim : INIT_BUCKET_SIZE  is %v ",b.opt.BucketSize)

	// prepare grpc
	b.rpc = grpc.NewServer()
	im.RegisterBasicServer(b.rpc, b)

	// prepare http
	b.http = http.NewServeMux()
	b.http.HandleFunc(b.opt.RouterHealth, b.handlerHealth)
	b.http.HandleFunc(b.opt.RouterConnection, b.connection)
	logging.Infof("sim : INIT_ROUTER_CONNECTION  is %s ",b.opt.RouterConnection)
	logging.Infof("sim : INIT_ROUTER_HEALTH  is %s ",b.opt.RouterHealth)
	logging.Infof("sim : INIT_VALIDATE_KEY is %s",b.opt.RouterValidateKey)
	return b
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
