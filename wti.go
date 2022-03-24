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

import "errors"

//  WebSocket Target Interface
type WTI interface {
	// 给用户打上标签
	SetTAG(cli Client, tags ...string)

	// 删除用户的标签
	DelTAG(cli Client, tags ...string)

	// 如果用户下线将会通知调用这个方法
	Update(token ...string)

	// 广播到包含标签对象
	BroadCast(content []byte, tag ...string)

	// 广播所有内容
	BroadCastByTarget(targetAndContent map[string][]byte)

	// 获取某个用户的所有的标签
	GetClienterTAGs(token string) []string

	// 获取到标签的创建时间
	GetTAGCreateTime(tag string) int64

	// 获取到所有tag的用户分布
	Distribute(tags ...string)map[string]*DistributeParam

	// 调用方法的回收房间的策略
	FlushWTI()
}




type DistributeParam struct {
	TagName string
	Onlines int64
	CreateTime int64
}


// 其他地方将调用这个变量，如果自己公司实现tag需要注入在程序中进行注入
var (
	factory WTI = newwti()
	isSupportWTI = atomic.NewBool(false)
)

func Inject(wti WTI) {
	factory = wti
}

func SetSupport (){
	isSupportWTI.Store(true)
}

var (
	ERRNotSupportWTI = errors.New("im/plugins/wti: you should call the SetSupport func")
)

func SetTAG(cli client.Clienter, tag ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	factory.SetTAG(cli, tag...)
	return nil
}

func DelTAG(cli client.Clienter, tag ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	factory.DelTAG(cli, tag...)
	return nil
}

func Update(token ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	factory.Update(token...)
	return nil
}

func BroadCast(content []byte, tag ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	factory.BroadCast(content, tag...)
	return nil
}

func BroadCastByTarget(targetAndContent map[string][]byte) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	factory.BroadCastByTarget(targetAndContent)
	return nil
}

func GetClienterTAGs(token string) ([]string, error) {
	if isSupportWTI.Load() == false {
		return nil, ERRNotSupportWTI
	}
	res := factory.GetClienterTAGs(token)
	return res, nil
}

func GetTAGCreateTime(tag string) (int64, error) {
	if isSupportWTI.Load() == false {
		return 0, ERRNotSupportWTI
	}
	res := factory.GetTAGCreateTime(tag)
	return res, nil
}


func Distribute() (map[string]*DistributeParam, error) {
	if isSupportWTI.Load() == false {
		return nil, ERRNotSupportWTI
	}
	res := factory.Distribute()
	return res, nil
}


func FlushWTI() error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	factory.FlushWTI()
	return nil
}
