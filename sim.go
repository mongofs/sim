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

const (
	RouterConnection = "/conn"
	RouterHealth     = "/health"

	ValidateKey = "token"
)

type Sim interface {
	Ping(context.Context, *im.Empty) (*im.Empty, error)
	Online(context.Context, *im.Empty) (*im.OnlineReply, error)
	SendMessageToMultiple(context.Context, *im.SendMsgReq) (*im.Empty, error)
	Broadcast(context.Context, *im.BroadcastReq) (*im.BroadcastReply, error)
	WTIBroadcast(context.Context, *im.BroadcastByWTIReq) (*im.BroadcastReply, error)
	WTIDistribute(context.Context, *im.Empty) (*im.WTIDistributeReply, error)
}

// Discover 可以在服务启动停止的时候自动想注册中心进行注册和注销，这个实现是可选的，具体可
// 查看option的参数，如果没有discover 就是一个单节点，也是可以启动。但是建议你在生产环境
// 使用的时候还是以集群方式启动，一旦存在集群的方式就必须注册这个方法，就可以将所有的内容作为
// 组件方式进行使用，可以通过独立网关进行sim地址下发。
type Discover interface {
	Register()
	Deregister()
}

// Validate 验证器，用户进行登录的时候需要进行验证，调用层需要注册Validate对象，整体流程
// 是进行ValidateKey 验证，验证方法也是服务注册的时候实现validate方法，具体见connection
// 方法，调用validate方法后可能会成功，可能会失败，不论成功或者失败都需要发送业务层的内容
// 所以需要服务注册时候实现validate整个接口
type Validate interface {
	// Validate 用户创建链接的时候需要验证令牌，令牌如何获取可以通过
	Validate(token string) error
	ValidateFailed(err error, cli Client)
	ValidateSuccess(cli Client)
}

// Receive 是需要用户进行注册，主要是为了接管用户上传的消息内容，在消息处理的时候可以根据
// 自身的业务需求进行处理
type Receive interface {
	Handle(Client Client, data <- chan []byte)
}
