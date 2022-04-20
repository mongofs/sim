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
)

// Ping 发送ping消息
func (s *sim) Ping(ctx context.Context, empty *im.Empty) (*im.Empty, error) {
	return nil, nil
}

// Online 在线用户
func (s *sim) Online(ctx context.Context, empty *im.Empty) (*im.OnlineReply, error) {
	num := s.ps.Load()
	req := &im.OnlineReply{Number: num}
	return req, nil
}

// Broadcast 给所有在线用户广播
func (s *sim) Broadcast(ctx context.Context, req *im.BroadcastReq) (*im.BroadcastReply, error) {
	fail := s.handlerBroadCast(req.Data,false)
	return &im.BroadcastReply{Fail: fail},nil
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

func (s *sim) WTITargetList(ctx context.Context, req *im.WTITargetListReq) (*im.WTITargetListInfoReply, error) {
	return nil,nil
}

//WTITargetInfo  获取某个TAG 的在线分布情况
func (s *sim) WTITargetInfo(ctx context.Context,req *im.WTITargetInfoReq) (*im.WTITargetInfoReply, error) {
	res,err := WTIInfo(req.Tag)
	if err !=nil {
		return nil,err
	}
	var gInfos []*im.Info
	for _,v := range res.GInfo {
		gInfos = append(gInfos, &im.Info{Info:v})
	}
	result := &im.WTITargetInfoReply{
		Tag:        res.name,
		Online:     int32(res.online),
		Limit:      int32(res.limit),
		CreateTime: res.createTime,
		Status:     int32(res.status),
		NumG:       int32(res.numG),
		GInfos:   gInfos ,
	}
	return result,nil
}

//WTIBroadcastByTarget  通过分组进行广播不同的内容
func (s *sim) WTIBroadcastByTarget(ctx context.Context,req  *im.WTIBroadcastReq) (*im.BroadcastReply, error) {
	res ,err := WTIBroadCastByTarget(req.Data)
	if err != nil {
		return nil,err
	}
	result := &im.BroadcastReply{Fail: res}
	return result,nil
}

//WTIBroadCastWithInnerJoinTag target 交集广播
func (s *sim) WTIBroadCastWithInnerJoinTag(ctx context.Context, req *im.WtiBroadcastWithInnerJoinReq) (*im.BroadcastReply, error) {
	res ,err := WTIBroadCastWithInnerJoinTag(req.Data,req.Tags)
	if err != nil {
		return nil,err
	}
	resutl := & im.BroadcastReply{Fail: res}
	return resutl,nil
}
