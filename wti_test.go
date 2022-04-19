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
	"fmt"
	"net/http"
)

type MockClient struct {
	token string
	tag map[string]string
}

func (m MockClient) SetTag(tag string) error {
	panic("implement me")
}

func (m MockClient) DelTag(tag string) {
	panic("implement me")
}

func (m MockClient) RangeTag() (res []string) {
	panic("implement me")
}

func (m MockClient) HaveTag(tags []string) bool {
	for _,v := range tags {
		if _,ok := m.tag[v];!ok{
			return false
		}
	}
	return true
}


func (m MockClient) Send(bytes []byte) error {
	if m.token == "token3" {return errors.New("bad push ")}
	fmt.Printf("token %s 收到消息： %v \n\r",m.token,string(bytes))
	return nil
}

func (m MockClient) Close(forRetry bool) error {
	panic("implement me")
}

func (m MockClient) ReFlushHeartBeatTime() {
	panic("implement me")
}

func (m MockClient) GetLastHeartBeatTime() int64 {
	panic("implement me")
}

func (m MockClient) Token() string {
	return m.token
}

func (m MockClient) Request() *http.Request {
	panic("implement me")
}

func (m MockClient) SetMessageType(messageType MessageType) {
	panic("implement me")
}

func (m MockClient) SetProtocol(protocol Protocol) {
	panic("implement me")
}

