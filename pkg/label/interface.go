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

package label

type TargetStatus int

const (
	TargetStatusNORMAL          = iota // normal
	TargetStatusShouldEXTENSION        // start extension
	TargetStatusShouldSHRINKS          // start shrinks
	TargetStatusShouldReBalance        // start reBalance
	TargetStatusShouldDestroy          // should destroy
)

const DefaultCapacity = 20

type Parallel func() error

// Manager Label管理器
type Manager interface {
	Run ()[] func()error

	// AddClient  添加一个用户到Label服务中，需要指出具体某个tag ,如果tag 不存在需要创建出新的tag
	// forClient 是需要用户将label 做一下本地保存，一个用户可以存储多个label，所以需要在用户操作的时候
	// 去协助调配label的增删改查
	AddClient(tag string, client Client) (ForClient, error)

	// List 获取当前存在的label ，获取label 列表信息
	List(limit, page int) []*LabelInfo

	// LabelInfo 获取具体label相信信息
	LabelInfo(tag string) (*LabelInfo, error)

	// BroadCastByLabel 通过label来进行广播，可以利用这个接口做版本广播，比如v1的用户传输内容格式是基于json
	// v2版本的用户是基于protobuf 可以通过这个api 非常便捷就可以完成
	BroadCastByLabel(tc map[string][]byte) ([]string, error)

	// BroadCastWithInnerJoinLabel 通过label的交集发布，比如要找到 v1版本、room1 、man 三个标签都满足
	// 才发送广播，此时可以通过这个接口
	BroadCastWithInnerJoinLabel(cont []byte, tags []string) ([]string,error)
}

// Client 存储单元的标准，每一个用户应该支持这几个方法才能算作一个客户，send 主要用做数据下发，HaveTags
// 做组内消息下推的筛选，Identification 是获取到用户的链接标示做存储，方法上有具体为什么使用哈希表做存储
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

// ForClient 用户使用工具，当用户被删除的时候就是，用户可以调用此方法，将用户标签在targetManager对象中删除
type ForClient interface {
	Delete(token []string) ([]string, int)
}

// Actor 是Label单元的行为集合，可以操作Actor实现者的创建
type Actor interface {

	// Destroy 判断status 状态为shouldDestroy的时候可以调用此方法
	Destroy()

	// Expansion label 本身支持扩张，如果用户在某个tag下增长到一定的人数，那么在这个target为了减少锁的粒度
	// 需要进行减小，那么相对应的操作就是增加新的容器进行存放用户，这就是扩容
	Expansion()

	// Shrinks label 本身支持缩容，如果用户在某个tag下缩减到一定的人数，那么在这个target为了减少锁的粒度
	// 需要进行减小，那么相对应的操作就是增加新的容器进行存放用户，这就是缩容
	Shrinks()

	// Balance 重新进行平衡，这个给外部进行调用，目的是为了在合适的时候进行重新平衡用户，如果说在扩张后就进行重
	// 平衡操作，会导致频繁的重平衡，那么需要外部在做一定的判断进行重平衡
	Balance()
}

// 广播器，用于进行广播，交集广播
type BroadCast interface {
	// 常规广播，入参为字节数组，将直接写入用户的链接中，返回值为失败的用户的切片
	// 主要报错实际在依赖注入的层面就可以记录，不需要这里进行操作
	// 交集广播，和上面的不同的是，只有拥有多个标签的用户才能进行广播，比如说 man 、 18、178
	// 这三个标签都满足了才能进行广播，我们只能选择广播器所依附的实体对象进行再筛选，一般依附的
	// 实体对象我们选择最少数量原则
	BroadCast(data []byte, tags ...string) []string
}

// label 标签管理单元，相同的标签会放在同样的标签实现中，标签是整个wti的管理单元，具有相同的标签的用户将会
// 放在这里
type Label interface {

	Actor

	BroadCast

	ForClient

	// Add 往标签中添加一个用户
	Add(cli Client) ForClient

	// Count 获取到label中的所有用户
	Count() int

	// Info 获取到label相关消息
	Info() *LabelInfo

	// Status 每次调用获取到当前target的相关状态，调用manager本身的其他暴露的端口进行耗时的操作，比如进行
	// 重平衡、进行扩容、进行缩容等操作，
	Status() TargetStatus
}

type LabelInfo struct {
	Name       string
	Online     int
	Limit      int
	CreateTime int64
	Status     int
	NumG       int
	Change     int //状态变更次数
	GInfo      []*map[string]string
}
