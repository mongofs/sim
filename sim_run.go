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
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"sim/pkg/logging"
	"time"
)

func (s *sim) Run() error {
	var prepareParallelFunc = []func() error{
		// 启用单独goroutine 进行监控
		s.monitorOnline,
		s.monitorWTI,
		// 启用单独goroutine 进行运行
		s.runGrpcServer,
		s.runHttpServer,
		s.PushBroadCast,
	}
	return ParallelRun(prepareParallelFunc...)
}

// Close  服务关闭
func (s *sim) Close() error {
	s.rpc.GracefulStop()
	s.cancel()
	return nil
}

// ParallelRun 并行的启动，使用goroutine 来进行管理
func ParallelRun(parallels ...func() error) error {
	wg := errgroup.Group{}
	for _, v := range parallels {
		wg.Go(v)
	}
	return wg.Wait()
}

// monitorOnline 统计用户在线人数
// 监控buffer 长度 并进行报警
func (s *sim) monitorOnline() error {
	for {
		n := int64(0)
		for _, bck := range s.bs {
			n += bck.Online()
		}
		s.ps.Store(n)
		time.Sleep(10 * time.Second)
	}
	return nil
}

// monitorWTI 统计用户在线人数
// 监控buffer 长度 并进行报警
func (s *sim) monitorWTI() error {
	if s.opt.SupportPluginWTI {
		for {
			FlushWTI()
			time.Sleep(20 * time.Second)
		}
	}
	return nil
}

// runGrpcServer 监控rpc 服务
func (s *sim) runGrpcServer() error {
	listen, err := net.Listen("tcp", s.opt.ServerRpcPort)
	if err != nil {
		return err
	}
	logging.Infof("sim : start GRPC server at %s ")
	if err := s.rpc.Serve(listen); err != nil {
		return err
	}
	return nil
}

// runHttpServer 运行httpserver
func (s *sim) runHttpServer() error {
	listen, err := net.Listen("tcp", s.opt.ServerHttpPort)
	if err != nil {
		return err
	}
	logging.Infof("im/run : start HTTP server at %s ", s.opt.ServerHttpPort)
	if err := http.Serve(listen, s.http); err != nil {
		return err
	}
	return nil
}

// PushBroadCast 单独处理广播业务 todo 后续会将此方法进行修改，主要方向会将用户句柄按照优先级进行优化
func (s *sim) PushBroadCast() error {
	wg := errgroup.Group{}
	for i := 0; i < s.opt.BroadCastHandler; i++ {
		wg.Go(func() error {
			for {
				data := <-s.buffer
				for _, v := range s.bs {
					err := v.BroadCast(data, false)
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
