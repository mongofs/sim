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
	im "sim/api/v1"
	"sim/pkg/errors"
)

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
		return nil, errors.ErrUserBufferIsFull
	}
	s.buffer <- req.Data
	return &im.BroadcastReply{
		Size: int64(len(s.buffer)),
	}, nil
}

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

func (s *sim) WTIBroadcast(ctx context.Context, req *im.BroadcastByWTIReq) (
	*im.BroadcastReply, error) {
	var err error
	err = BroadCastByTarget(req.Data)
	return &im.BroadcastReply{
		Size: int64(len(s.buffer)),
	}, err
}

func (s *sim) SendMessageToMultiple(ctx context.Context, req *im.SendMsgReq) (
	*im.Empty, error) {
	var err error
	for _, token := range req.Token {
		bs := s.bucket(token)
		err = bs.Send(req.Data, token, false)
	}
	return &im.Empty{}, err
}
