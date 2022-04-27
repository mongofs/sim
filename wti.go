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
	"sim/pkg/target"
)

type WTIManager interface {
	target.TargetManager
	target.ClientManager
	target.BroadCastManager
}


func StartWTIServer() []ParallelFunc {
	return wti.run()
}

func WTIAdd(tag string, client Client) (target.ClientManager, error) {
	if client == nil {
		return nil, nil
	}
	return wti.add(tag, client)
}

func WTIList() []*target.TargetInfo {
	return wti.list()
}

func WTIInfo(tag string) (*target.TargetInfo, error) {
	if len(tag) == 0 {
		return nil, nil
	}
	return wti.info(tag)
}

func WTIBroadCastByTarget(tc map[string][]byte) ([]string, error) {
	if tc == nil {
		return nil, nil
	}
	return wti.broadcastByTag(tc)
}

func WTIBroadCastWithInnerJoinTag(cont []byte, tags []string) []string {
	if cont == nil || tags == nil {
		return nil
	}
	return wti.broadcast(cont,tags...)
}
