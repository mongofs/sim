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

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockClient struct {
	token string
}

func (m MockClient) Send(bytes []byte) error {
	fmt.Printf("mockeClient : %v \n\r", string(bytes))
	return nil
}

func (m MockClient) HaveTags(strings []string) bool {
	return true
}

func (m MockClient) Identification() string {
	return m.token
}

func TestLabel_Status(t *testing.T) {
	Convey("进行各种状态测试，查看状态是否判断正确", t, func() {
		Convey("测试扩容过程中状态判断是否正确", func() {
			// 容量为2
			tg, err := NewLabel("example", 2)
			if err != nil {
				t.Fatal(err)
			}
			tg.Add(&MockClient{token: "1111"})
			tgStatus1 := tg.Status()
			So(tgStatus1 == TargetStatusNORMAL && tg.targetG == 1, ShouldBeTrue)
			tg.Add(&MockClient{token: "1222"})
			tgStatus2 := tg.Status()
			So(tgStatus2 == TargetStatusShouldEXTENSION && tg.targetG == 2, ShouldBeTrue)
			tg.Add(&MockClient{token: "1333"})
			tg.Add(&MockClient{token: "1334"})
			tgStatus3 := tg.Status()
			So(tgStatus3 == TargetStatusShouldEXTENSION && tg.targetG == 3, ShouldBeTrue)
		})

		Convey("测试需要缩容的情况是否正确", func() {
			tg, err := NewLabel("example", 2)
			if err != nil {
				t.Fatal(err)
			}
			tg.expansion(10)
			tgStatus1 := tg.Status()
			So(tgStatus1 == TargetStatusShouldSHRINKS && tg.targetG == 1 && tg.numG == 11, ShouldBeTrue)
			tg.expansion(2)
			tgStatus2 := tg.Status()
			So(tgStatus2 == TargetStatusShouldSHRINKS && tg.targetG == 1 && tg.numG == 13, ShouldBeTrue)
			tg.Add(&MockClient{token: "1333"})
			tg.Add(&MockClient{token: "1334"})
			tgStatus3 := tg.Status()
			So(tgStatus3 == TargetStatusShouldSHRINKS && tg.targetG == 2 && tg.numG == 13, ShouldBeTrue)
			tg.Add(&MockClient{token: "1335"})
			tg.Add(&MockClient{token: "1336"})
			tg.Add(&MockClient{token: "1337"})
			tgStatus4 := tg.Status()
			So(tgStatus4 == TargetStatusShouldSHRINKS && tg.targetG == 3 && tg.numG == 13, ShouldBeTrue)
			tg.shrinks(5) // numG = 8
			So(tgStatus4 == TargetStatusShouldSHRINKS && tg.targetG == 3 && tg.numG == 8, ShouldBeTrue)
		})


		Convey("测试在状态错误下能否自行校正", func() {
			tg, err := NewLabel("example", 2)
			if err != nil {
				t.Fatal(err)
			}
			tg.Add(&MockClient{token: "1111"})
			tg.Add(&MockClient{token: "1112"})
			tg.Add(&MockClient{token: "1113"})
			tg.Add(&MockClient{token: "1114"})
			tgStatus1 := tg.Status()
			So(tgStatus1 == TargetStatusShouldEXTENSION && tg.targetG == 3 && tg.numG == 1, ShouldBeTrue)
			tg.expansion(2)
			// 状态回归正常后
			tgStatus2 := tg.Status()
			So(tgStatus2 == TargetStatusShouldReBalance && tg.targetG == 3 && tg.numG == 3, ShouldBeTrue)
		})
	})
}

func TestTarget_Balance(t *testing.T) {
	Convey("测试重平衡", t, func() {
		tg, err := NewLabel("example", 20,)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 200; i++ {
			tg.add(&MockClient{token: fmt.Sprintf("aaa_%d", i)})
		}
		tg.expansion(7)
		fmt.Println("------------------expansion to 8 group",tg.distribute())
		tg.Balance()
		fmt.Println("------------------after expansion to 8 group, Balance",tg.distribute())
		tg.shrinks(5)
		fmt.Println("------------------shrinks to 3 group",tg.distribute())
		tg.Balance()
		fmt.Println("------------------after shrinks to 3 group, Balance",tg.distribute())
		tg.expansion(128)
		fmt.Println("------------------expansion to 131 group",tg.distribute())
		tg.Balance()
		fmt.Println("------------------after expansion to 131 group",tg.distribute())
		tg.shrinks(127)
		fmt.Println("------------------shrinks expansion to 4 group",tg.distribute())
		tg.Balance()
		fmt.Println("------------------after shrinks to 4 group",tg.distribute())
	})
}
