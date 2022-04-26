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

type TargetStatus int

const (
	TargetStatusNORMAL          = iota // normal
	TargetStatusShouldEXTENSION        // start extension
	TargetStatusEXTENSION              // extension
	TargetStatusShouldSHRINKS          // start shrinks
	TargetStatusSHRINKS                // shrinks
	TargetStatusShouldReBalance        // start reBalance
	TargetStatusREBALANCE              // reBalance
	TargetStatusShouldBeDestroy        // should destroy
)

type Client interface {
	Send([]byte) error

	// HaveTags 判断用户是否存在这批tag，这个函数是用于支持BroadManager的交集推送，如何判断交集中实际
	// 上带来一定的定制化嫌疑，但是目前想到的最快的方法就是这样，可以支持On 的复杂度进行交集下推
	HaveTags([]string) bool
	// 在用户管理中，需要将用户的标识作为key值进行存储，原本打算使用链表进行存储的，但是发现还是存在部分需求
	// 需要进行支持快速查找，比如快速找到组内是否存在某个用户，用户存在就将用户标识对应的内容进行替换，如果
	// 不存在就进行新增。使用链表就不是很有必要，所以决定是用hash表进行存储
	Identification() string
}


type ClientManager interface {
	Add(cli Client)
	Del(token []string) ([]string, int)
	Count() int
}

type TargetManager interface {
	Info() *TargetInfo

	Status() TargetStatus

	// Destroy 判断status 状态为
	Destroy()

	// Expansion target 本身支持扩张，如果用户在某个tag下增长到一定的人数，那么在这个target为了减少锁的粒度
	// 需要进行减小，那么相对应的操作就是增加新的容器进行存放用户，这就是扩容
	Expansion()

	// Shrinks target 本身支持缩容，如果用户在某个tag下缩减到一定的人数，那么在这个target为了减少锁的粒度
	// 需要进行减小，那么相对应的操作就是增加新的容器进行存放用户，这就是缩容
	Shrinks()

	// Balance 重新进行平衡，这个给外部进行调用，目的是为了在合适的时候进行重新平衡用户，如果说在扩张后就进行重
	// 平衡操作，会导致频繁的重平衡，那么需要外部在做一定的判断进行重平衡
	Balance()
}

// 广播器，用于进行广播，交集广播
type BroadCastManager interface {
	// 常规广播，入参为字节数组，将直接写入用户的链接中，返回值为失败的用户的切片
	// 主要报错实际在依赖注入的层面就可以记录，不需要这里进行操作

	// 交集广播，和上面的不同的是，只有拥有多个标签的用户才能进行广播，比如说 man 、 18、178
	// 这三个标签都满足了才能进行广播，我们只能选择广播器所依附的实体对象进行再筛选，一般依附的
	// 实体对象我们选择最少数量原则
	BroadCast(data []byte,tags[]string) []string
}

type TargetInfo struct {
	name       string
	online     int
	limit      int
	createTime int64
	status     int
	numG       int
	change     int //状态变更次数
	GInfo      []*map[string]string
}
