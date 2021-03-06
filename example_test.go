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

import "sim/pkg/logging"

type MockReceive struct{}

func (m MockReceive) Handle(conn Connect, data []byte) {
	conn.ReFlushHeartBeatTime()
}

type MockValidate struct{}

func (m MockValidate) Validate(token string) error {
	return nil
}

func (m MockValidate) ValidateFailed(err error, cli Client) {
	panic("implement me")
}

func (m MockValidate) ValidateSuccess(cli Client) {
	return
}

type mockDiscovery struct {
}

func (m mockDiscovery) Register() {
	logging.Infof("sim : start Discover.Register success")
}

func (m mockDiscovery) Deregister() {
	logging.Infof("sim : start Discover.Deregister success")
}

func ExampleRun() {
	optionfunc := []OptionFunc{
		WithServerRpcPort(":8089"),
		WithServerHttpPort(":8081"),
		WithLoggerLevel(logging.DebugLevel),
		WithLabelManager(),
		WithDiscover(mockDiscovery{}),
	}
	Run(&MockValidate{}, &MockReceive{}, optionfunc...)
}
