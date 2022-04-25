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
	. "github.com/smartystreets/goconvey/convey"
	"sim/pkg/errors"
	"testing"
)

func TestNewGroup(t *testing.T) {
	Convey("测试创建group", t, func() {
		g := NewGroup(10)
		So(g.cap == 10, ShouldBeTrue)
		So(g.createTime != 0, ShouldBeTrue)

		g1 := NewGroup(0)
		So(g1.cap == DefaultCapacity, ShouldBeTrue)
	})
}

func TestGroup_Add(t *testing.T) {
	Convey("测试添加Group获取返回值", t, func() {
		Convey("测试添加超额的内容，查看返回是否正确", func() {
			g := NewGroup(2)
			{
				same := g.Add(&MockClient{token: "token1"})
				So(same == false, ShouldBeTrue)
			}
			{
				same := g.Add(&MockClient{token: "token2"})
				So(same == false, ShouldBeTrue)
			}
			{
				same := g.Add(&MockClient{token: "token3"})
				So(same == false, ShouldBeTrue)
			}
		})
		Convey("测试添加相同内容，查看返回是否正确", func() {
			g := NewGroup(20)
			g.Add(&MockClient{token: "token1"})
			g.Add(&MockClient{token: "token1"})
			g.Add(&MockClient{token: "token1"})
			So(g.num == 1, ShouldBeTrue)
		})
	})
}

func TestGroup_Del(t *testing.T) {
	Convey("测试删除用户的group", t, func() {
		g := NewGroup(2)
		g.Add(&MockClient{token: "token1"})
		g.Add(&MockClient{token: "token2"})
		So(g.num == 2, ShouldBeTrue)
		So(g.load == 0, ShouldBeTrue)

		stop, _, _ := g.Del([]string{"token1"})
		So(stop == true, ShouldBeTrue)
		So(g.num == 1, ShouldBeTrue)
		So(g.load == 1, ShouldBeTrue)

		stop1, _, _ := g.Del([]string{"token3"})
		So(stop1 == false, ShouldBeTrue)
	})
}

func TestGroup_Move(t *testing.T) {
	Convey("测试批量移除用户", t, func() {
		g := NewGroup(2)
		g.Add(&MockClient{token: "token1"})
		g.Add(&MockClient{token: "token2"})
		g.Add(&MockClient{token: "token3"})
		g.Add(&MockClient{token: "token4"})
		So(g.num == 4 && g.load == -2, ShouldBeTrue)

		_, err := g.Move(20)
		So(err == errors.ErrGroupBadParam, ShouldBeTrue)

		_, err = g.Move(-500)
		So(err == errors.ErrGroupBadParam, ShouldBeTrue)

		res, err := g.Move(3)
		So(err == nil, ShouldBeTrue)
		So(len(res) == 3, ShouldBeTrue)
		So(g.num == 1, ShouldBeTrue)
		So(g.load == 1, ShouldBeTrue)
	})
}

func TestGroup_BatchAdd(t *testing.T) {
	Convey("测试批量添加用户", t, func() {
		tests := []Client{
			&MockClient{token: "1111"},
			&MockClient{token: "2222"},
			&MockClient{token: "3333"},
			&MockClient{token: "1111"},
		}
		g := NewGroup(3)
		g.BatchAdd(tests)

		So(g.num == 3, ShouldBeTrue)
		So(g.load == 0, ShouldBeTrue)
	})
}

func TestGroup_BroadCast(t *testing.T) {
	Convey("测试广播", t, func() {
		g := NewGroup(2)
		g.Add(&MockClient{token: "token1"})
		g.Add(&MockClient{token: "token2"})
		g.Add(&MockClient{token: "token3"})
		g.Add(&MockClient{token: "token4"})

		fail := g.BroadCast([]byte("hello old baby"))
		So(len(fail) == 1 && fail[0] == "token3", ShouldBeTrue)
	})
}

func TestGroup_BroadCastWithOtherTag(t *testing.T) {
	Convey("测试交集广播", t, func() {
		g := NewGroup(2)
		g.Add(&MockClient{token: "token1", tag: map[string]string{"roomA": "1"}})
		g.Add(&MockClient{token: "token2", tag: map[string]string{"roomB": "1"}})
		g.Add(&MockClient{token: "token3", tag: map[string]string{"roomA": "1"}})
		g.Add(&MockClient{token: "token4", tag: map[string]string{"roomA": "1"}})
		g.Add(&MockClient{token: "token5", tag: map[string]string{"roomB": "1", "roomC": "1"}})

		fail := g.BroadCastWithOtherTag([]byte("hello old baby"), []string{"roomA"})
		So(len(fail) == 1 && fail[0] == "token3", ShouldBeTrue)
		// output :
		// token token1 收到消息： hello old baby
		// token token4 收到消息： hello old baby
		fail = g.BroadCastWithOtherTag([]byte("hello old baby"), []string{"roomB", "roomC"})
		// output :
		// token token5 收到消息： hello old baby
	})
}

