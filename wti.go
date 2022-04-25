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

type targetFlag int

const (
	TargetFlagNORMAL          = iota // normal
	TargetFLAGShouldEXTENSION        // start extension
	TargetFLAGEXTENSION              // extension
	TargetFLAGShouldSHRINKS          // start shrinks
	TargetFLAGSHRINKS                // shrinks
	TargetFLAGShouldReBalance        // start reBalance
	TargetFLAGREBALANCE              // reBalance
)

type Monitor interface {
	Monitor() error
}

type Balancer interface {
	Expansion()
	Shrinks()
	Balance()
}

type BroadCaster interface {
	BroadCast([]byte)
	BroadCastByTarget(map[string][]byte)
	BroadCastWithInnerJoinTag([]byte, []string)
}

func StartWTIServer() []ParallelFunc {
	return wti.RegisterParallelFunc()
}

func WTIAdd(tag string, client Client) (*target, error) {
	if client == nil {
		return nil, nil
	}
	return wti.Add(tag, client)
}

func WTIList() []*targetInfo {
	return wti.List()
}

func WTIInfo(tag string) (*targetInfo, error) {
	if len(tag) == 0 {
		return nil, nil
	}
	return wti.Info(tag)
}

func WTIBroadCastByTarget(tc map[string][]byte) ([]string, error) {
	if tc == nil {
		return nil, nil
	}
	return wti.BroadCastByTarget(tc)
}

func WTIBroadCastWithInnerJoinTag(cont []byte, tags []string) ([]string, error) {
	if cont == nil || tags == nil {
		return nil, nil
	}
	return wti.BroadCastWithInnerJoinTag(cont, tags)
}
