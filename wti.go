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
	"sim/pkg/logging"
)

type targetFlag int

const (
	TargetFlagNORMAL          = iota + 1 // normal
	TargetFLAGShouldEXTENSION            // start extension
	TargetFLAGEXTENSION                  // extension
	TargetFLAGShouldSHRINKS              // start shrinks
	TargetFLAGSHRINKS                    // shrinks
	TargetFLAGShouldReBalance            // start reBalance
	TargetFLAGREBALANCE                  // reBalance
)

type Adder interface {
	Add(Client)
}

type Deleter interface {
	Del([]string) ([]string, int)
}

type Expander interface {
	Expansion()
}

type Shrinker interface {
	Shrinks()
}

type Balancer interface {
	Balance()
}

type Monitor interface {
	Monitor() error
}

type BroadCaster interface {
	BroadCast([]byte)
	BroadCastByTarget(map[string][]byte)
	BroadCastWithInnerJoinTag([]byte, []string)
}

// =================================== API ==============

func WTIAdd(adder Adder,tag string, client Client) {
	if adder == nil || client == nil {
		return
	}
	adder.Add(client)
}

func WTIDel(del Deleter, token []string) error {
	if token == nil || del == nil {
		return nil
	}
	tokens, current := del.Del(token)
	logging.Infof("sim/wti :  下线用户 %v ,剩余在线人数 ： %v", tokens, current)
	return nil
}

func WTIBroadCast(cas BroadCaster, cont []byte) {
	if cas == nil || cont == nil {
		return
	}
	cas.BroadCast(cont)
}

func WTIBroadCastWithInnerJoinTag(cas BroadCaster, cont []byte, tags []string) {
	if cas == nil || cont == nil || tags == nil {
		return
	}
	cas.BroadCastWithInnerJoinTag(cont, tags)
}

func WTIBroadCastByTarget(cas BroadCaster, tc map[string][]byte) {
	if tc == nil || cas == nil {
		return
	}
	cas.BroadCastByTarget(tc)
}

func WTIMonitor(monitor Monitor) error{
	return monitor.Monitor()
}



