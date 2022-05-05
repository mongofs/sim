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
	"net"
	"net/http"
	"strconv"
	"time"

	im "github.com/mongofs/sim/api/v1"
	"github.com/mongofs/sim/pkg/label"
	"github.com/mongofs/sim/pkg/logging"

	"github.com/zhenjl/cityhash"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

type sim struct {
	http  *httpserver
	rpc   *grpc.Server
	label label.Manager
	bs    []*bucket
	ps    atomic.Int64

	cancel context.CancelFunc
	ctx    context.Context
	opt    *Options
}

func (s *sim) Ping(ctx context.Context, empty *im.Empty) (*im.Empty, error) {
	return nil, nil
}

func (s *sim) Online(ctx context.Context, empty *im.Empty) (*im.OnlineReply, error) {
	num := s.ps.Load()
	req := &im.OnlineReply{Number: num}
	return req, nil
}

func (s *sim) Broadcast(ctx context.Context, req *im.BroadcastReq) (*im.BroadcastReply, error) {
	fail := s.handlerBroadCast(req.Data, false)
	return &im.BroadcastReply{Fail: fail}, nil
}

func (s *sim) SendMsg(ctx context.Context, req *im.SendMsgReq) (*im.SendMsgResp, error) {
	var err error
	fail := map[string]string{}
	var success []string
	for _, token := range req.Token {
		bs := s.bucket(token)
		err = bs.send(req.Data, token, false)
		if err != nil {
			fail[token] = err.Error()
		} else {
			success = append(success, token)
		}
	}

	result := &im.SendMsgResp{
		MsgID:   "",
		Filed:   &im.Load{Token: fail},
		Success: success,
	}
	return result, err
}

//

func (s *sim) LabelList(ctx context.Context, req *im.LabelListReq) (*im.LabelListReply, error) {
	res := s.label.List(0, 0)
	var result []*im.Info
	for _, v := range res {
		result = append(result, &im.Info{Info: map[string]string{
			"tag":        v.Name,
			"online":     strconv.Itoa(v.Online),
			"limit":      strconv.Itoa(v.Limit),
			"createTime": strconv.Itoa(int(v.CreateTime)),
			"status":     strconv.Itoa(v.Status),
			"change":     strconv.Itoa(v.Change),
			"numG":       strconv.Itoa(v.NumG),
		}})
	}
	return &im.LabelListReply{
		Count: int32(len(result)),
		Info:  result,
	}, nil
}
func (s *sim) LabelInfo(ctx context.Context, req *im.LabelInfoReq) (*im.LabelInfoReply, error) {
	res, err := s.label.LabelInfo(req.Tag)
	if err != nil {
		return nil, err
	}
	var gInfos []*im.Info
	for _, v := range res.GInfo {
		gInfos = append(gInfos, &im.Info{Info: *v})
	}
	result := &im.LabelInfoReply{
		Tag:        res.Name,
		Online:     int32(res.Online),
		Limit:      int32(res.Limit),
		CreateTime: res.CreateTime,
		Status:     int32(res.Status),
		NumG:       int32(res.NumG),
		GInfos:     gInfos,
	}
	return result, nil
}
func (s *sim) BroadCastByLabel(ctx context.Context, req *im.BroadCastByLabelReq) (*im.BroadcastReply, error) {
	res, err := s.label.BroadCastByLabel(req.Data)
	if err != nil {
		return nil, err
	}
	result := &im.BroadcastReply{Fail: res}
	return result, nil
}
func (s *sim) BroadCastByLabelWithInJoin(ctx context.Context, req *im.BroadCastByLabelWithInJoinReq) (*im.BroadcastReply, error) {
	res, err := s.label.BroadCastWithInnerJoinLabel(req.Data, req.Tags)
	if err != nil {
		return nil, err
	}
	resutl := &im.BroadcastReply{Fail: res}
	return resutl, nil
}

func (s *sim) initHttp() {
	s.http = newHttpServer(s.opt.RouterValidateKey, s.opt.RouterConnection, s.opt.ServerHttpPort, s.upgrade, s)
}

func (s *sim) bucket(token string) *bucket {
	idx := s.Index(token, uint32(s.opt.ServerBucketNumber))
	return s.bs[idx]
}

func (s *sim) Index(token string, size uint32) uint32 {
	return cityhash.CityHash32([]byte(token), uint32(len(token))) % size
}

func (s *sim) upgrade(writer http.ResponseWriter, r *http.Request, token string) error {
	// validate token
	bs := s.bucket(token)
	cli, err := bs.createConn(writer, r, token)
	if err != nil {
		return err
	}
	if err := s.opt.ServerValidate.Validate(token); err != nil {
		s.opt.ServerValidate.ValidateFailed(err, cli)
		return nil
	} else {
		s.opt.ServerValidate.ValidateSuccess(cli)
	}
	// register to data
	if err := bs.Register(cli, token); err != nil {
		cli.Send([]byte(err.Error()))
		cli.Close(false)
		return err
	}
	return nil
}

func (s *sim) handlerHealth(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Status: 200,
		Data:   "ok",
	}
	res.SendJson()
}

func (s *sim) monitorOnline(ctx context.Context) error {
	var interval = 10
	logging.Infof("sim : monitor of sim online  starting , interval is %v second", interval)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		case <-ticker.C:
			n := int64(0)
			for _, bck := range s.bs {
				n += bck.Online()
			}
			s.ps.Store(n)
		case <-ctx.Done():
			goto loop
		}
	}
loop:
	logging.Infof("sim : monitor of sim online is closed  ")
	return nil
}

func (s *sim) runGrpcServer(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.opt.ServerRpcPort)
	if err != nil {
		return err
	}
	defer listen.Close()
	logging.Infof("sim : start GRPC example at %s ", s.opt.ServerRpcPort)
	if err := s.rpc.Serve(listen); err != nil {
		return err
	}
	return nil
}

func (s *sim) handlerBroadCast(data []byte, ack bool) []string {
	var res []string
	for _, v := range s.bs {
		res = append(res, v.broadCast(data, false)...)
	}
	return res
}

func (s *sim) close() error {
	s.cancel()
	s.rpc.GracefulStop()
	logging.Infof("sim : server is closed ")
	return nil
}

func initSim(opts *Options) *sim {
	b := &sim{
		ps:  atomic.Int64{},
		opt: opts,
	}
	b.ps.Store(0)

	// prepare buckets
	b.bs = make([]*bucket, b.opt.ServerBucketNumber)
	b.ctx, b.cancel = context.WithCancel(context.Background())

	for i := 0; i < b.opt.ServerBucketNumber; i++ {
		b.bs[i] = newBucket(b.opt)
	}
	logging.Infof("sim : INIT_BUCKET_NUMBER is %v ", b.opt.ServerBucketNumber)
	logging.Infof("sim : INIT_BUCKET_SIZE  is %v ", b.opt.BucketSize)

	// prepare grpc
	b.rpc = grpc.NewServer()
	im.RegisterBasicServer(b.rpc, b)

	// prepare http
	b.initHttp()

	logging.Infof("sim : INIT_ROUTER_CONNECTION  is %s ", b.opt.RouterConnection)
	logging.Infof("sim : INIT_ROUTER_HEALTH  is %s ", b.opt.RouterHealth)
	logging.Infof("sim : INIT_VALIDATE_KEY is %s", b.opt.RouterValidateKey)
	return b
}
