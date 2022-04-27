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
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewTarget(t *testing.T) {
	Convey("create a target", t, func() {
		Convey("create should  fail", func() {
			_, err := NewTarget("", 20)
			So(err != nil, ShouldBeTrue)
			_, err1 := NewTarget("demo", 0)
			So(err1 != nil, ShouldBeTrue)
		})
		Convey("create should success ", func() {
			tg, err := NewTarget("demo", 20)
			if err != nil {
				t.Fatal(err)
			}
			status := tg.Status()
			So(status == TargetFlagNORMAL, ShouldBeTrue)
			So(tg.num == 0, ShouldBeTrue)
			So(tg.numG == 1, ShouldBeTrue)
			So(tg.createTime == 0, ShouldBeFalse)
		})
	})
}

func TestTarget_Add(t *testing.T) {
	Convey("test target Add client ", t, func() {
		Convey("测试创建相同token", func() {
			tg, err := NewTarget("demo", 20)
			if err != nil {
				t.Fatal(err)
			}
			tg.Add(&MockClient{token: "aaaa"})
			tg.Add(&MockClient{token: "bbbb"})
			So(tg.num == 2, ShouldBeTrue)
			// test for retry conn
			// 创建连接，实际上是一个比较严重的bug，最好要避免
			newAconn := &MockClient{token: "aaaa"}
			tg.Add(newAconn)
			So(tg.num == 2, ShouldBeTrue)
		})
		Convey("测试创建超过单个组容量的客户端", func() {
			tg, err := NewTarget("demo", 2)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 25; i++ {
				token := fmt.Sprintf("%d%d", i, i)
				tg.Add(&MockClient{token: token})
			}
			So(tg.num == 25, ShouldBeTrue)
		})
	})
}

func TestTarget_Del(t *testing.T) {
	Convey("测试target中删除某个用户", t, func() {
		tg, err := NewTarget("demo", 20)
		if err != nil {
			t.Fatal(err)
		}
		tg.Add(&MockClient{token: "1111"})
		tg.Add(&MockClient{token: "1222"})
		tg.Add(&MockClient{token: "1333"})
		tg.Add(&MockClient{token: "1444"})
		res1, cur1 := tg.Del([]string{"1111"})
		So(cur1 == 3, ShouldBeTrue)
		So(res1[0] == "1111", ShouldBeTrue)
		res2, cur2 := tg.Del([]string{"1111"})
		So(len(res2) == 0, ShouldBeTrue)
		So(cur2 == 3, ShouldBeTrue)
	})
}

func TestTarget_Expansion(t *testing.T) {
	Convey("test expansion ", t, func() {
		tg, err := NewTarget("demo", 2)
		if err != nil {
			t.Fatal(err)
		}
		tg.Add(&MockClient{token: "1111"})
		tg.Add(&MockClient{token: "1222"})
		So(tg.flag == TargetFlagNORMAL, ShouldBeTrue)
		So(tg.num == 2, ShouldBeTrue)
		tg.Add(&MockClient{token: "1333"})
		So(tg.flag == TargetFLAGShouldEXTENSION, ShouldBeTrue)
		tg.Expansion()
		So(tg.numG == 2, ShouldBeTrue)
		So(tg.flag == TargetFlagNORMAL, ShouldBeTrue)
		tg.Add(&MockClient{token: "1334"})
		So(tg.li.Back() == tg.offset, ShouldBeTrue)
	})
}

func TestTarget_Shrinks(t *testing.T) {

	Convey("测试缩容", t, func() {
		tg, err := NewTarget("demo", 10)
		if err != nil {
			t.Fatal(err)
		}
		tg.Expansion()
		tg.Expansion()
		tg.Expansion()
		tg.Expansion()
		tg.add(&MockClient{token: "1223"})
		tg.add(&MockClient{token: "1224"})
		tg.add(&MockClient{token: "1225"})
		tg.Shrinks()
		So(tg.flag == TargetFLAGShouldReBalance, ShouldBeTrue)
	})
}

func TestTarget_Balance(t *testing.T) {
	Convey("测试重平衡", t, func() {
		Convey("测试容量为20", func() {
			tg, err := NewTarget("demo", 20)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 200; i++ {
				tg.add(&MockClient{token: fmt.Sprintf("aaa_%d", i)})
			}
			fmt.Println(tg.distribute())
			tg.Expansion()
			tg.reBalance()
			fmt.Println(tg.Distribute())
		})
		Convey("测试容量为200", func() {
			tg, err := NewTarget("demo", 200)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 200; i++ {
				tg.add(&MockClient{token: fmt.Sprintf("aaa_%d", i)})
			}
			fmt.Println(tg.Distribute())
			tg.Expansion()
			tg.reBalance()
			fmt.Println(tg.Distribute())
		})
		Convey("测试容量为300", func() {
			tg, err := NewTarget("demo", 300)
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 200; i++ {
				tg.add(&MockClient{token: fmt.Sprintf("aaa_%d", i)})
			}
			fmt.Println(tg.Distribute())
			tg.Expansion()
			tg.reBalance()
			fmt.Println(tg.Distribute())
		})

	})
}


