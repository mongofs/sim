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
	"errors"
	"go.uber.org/atomic"
)

//WTI   WebSocket Target Interface ,client 会注册到具体的taget中，所涉及到的taget操作
// action : find ,
type WTI interface {

	// 找到某个具体的target ，然后 cli.SetTag (*result), and the
	Find(tags []string) []*target

	// Update  如果用户下线将会通知调用这个方法
	Update(token ...string)

	// BroadCast 广播到包含标签对象
	BroadCast(content []byte, tag ...string)

	// BroadCastByTarget 广播所有内容
	BroadCastByTarget(targetAndContent map[string][]byte)

	// Distribute 获取到所有tag的用户分布
	Distribute(tags ...string)map[string]*DistributeParam

	// FlushWTI 调用方法的回收房间的策略
	FlushWTI()
}







// 其他地方将调用这个变量，如果自己公司实现tag需要注入在程序中进行注入
var (
	//factoryWTI      WTI = newwti()
	isSupportWTI     = atomic.NewBool(false)
)

func InjectWTI(wti WTI) {
	//factoryWTI = wti
}

func SetSupport (){
	isSupportWTI.Store(true)
}

var (
	ERRNotSupportWTI = errors.New("im/plugins/wti: you should call the SetSupport func")
)

func SetTAG(cli Client, tag ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
	//factory.Set(cli, tag...)
	return nil
}

func DelTAG(cli Client, tag ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
//	factory.Del(cli, tag...)
	return nil
}

func Update(token ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
//	factory.Update(token...)
	return nil
}

func BroadCast(content []byte, tag ...string) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
//	factory.BroadCast(content, tag...)
	return nil
}

func BroadCastByTarget(targetAndContent map[string][]byte) error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
//	factory.BroadCastByTarget(targetAndContent)
	return nil
}

func GetClientTAGs(token string) ([]string, error) {
	if isSupportWTI.Load() == false {
		return nil, ERRNotSupportWTI
	}
//	res := factory.GetClientTAGs(token)
	return nil, nil
}

func GetTAGCreateTime(tag string) (int64, error) {
	if isSupportWTI.Load() == false {
		return 0, ERRNotSupportWTI
	}
//	res := factory.GetTAGCreateTime(tag)
	return 0, nil
}


func Distribute() (map[string]*DistributeParam, error) {
	if isSupportWTI.Load() == false {
		return nil, ERRNotSupportWTI
	}
//	res := factory.Distribute()
	return nil, nil
}


func FlushWTI() error {
	if isSupportWTI.Load() == false {
		return ERRNotSupportWTI
	}
//	factory.FlushWTI()
	return nil
}
