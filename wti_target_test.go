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
	."github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewTarget(t *testing.T) {
	Convey("create a target",t, func() {
		Convey("create should  fail", func() {
			_,err := NewTarget("",20)
			So(err != nil ,ShouldBeTrue)
			_,err1:= NewTarget("demo",0)
			So(err1 != nil ,ShouldBeTrue)
		})
		Convey("create should success ", func() {
			tg,err := NewTarget("demo",20)
			if err !=nil {t.Fatal(err)}
			status := tg.Status()
			So(status== TargetFlagNORMAL,ShouldBeTrue)
			So(tg.num == 0,ShouldBeTrue)
			So(tg.numG == 1,ShouldBeTrue)
			So(tg.createTime==0,ShouldBeFalse)
		})
	})
}

func  TestTarget_Add(t *testing.T) {
	tg,err := NewTarget("demo",20)
	if err !=nil {t.Fatal(err)}
	Convey("test target Add client ",t, func() {
		tg.Add(&MockClient{"aaaa"})
		tg.Add(&MockClient{"bbbb"})
		So(tg.num == 2 ,ShouldBeTrue)
		// test for retry conn
		// 创建连接，实际上是一个比较严重的bug，最好要避免
		newAconn := &MockClient{"aaaa"}
		tg.Add(newAconn)
		So()

	})
}