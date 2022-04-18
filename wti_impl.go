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
	"math"
	"sim/pkg/logging"
	"sync"
	"time"
)

var wti = newSet()

type set struct {
	// mp tagName =>
	mp map[string]*target // wti => []string
	rw *sync.RWMutex

	limit     int
	watchTime int
	expansion chan *target
	shrinks   chan *target
	balance   chan *target
}

func newSet() *set {
	return &set{
		mp: map[string]*target{},
		rw: &sync.RWMutex{},
	}
}

// ======================================API =================================

func (s *set) Run()error{
	go s.monitor()
	go s.handleMonitor()
	return nil
}

func (s *set) Add(tag string, client Client) {
	s.rw.Lock()
	defer s.rw.Unlock()
	if target, ok := s.mp[tag]; ok {
		target.Add(client)
	} else {
		ctag, err := NewTarget(tag, s.limit)
		if err != nil {
			logging.Error(err)
		}
		s.mp[tag] = ctag
	}
}

func (s *set) BroadCast(cont []byte) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	for _, v := range s.mp {
		v.BroadCast(cont)
	}
}

func (s *set) BroadCastByTarget(msg map[string][]byte) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	for tag, cont := range msg {
		if tg, ok := s.mp[tag]; ok {
			tg.BroadCast(cont)
		}
	}
}

func (s *set) BroadCastWithInnerJoinTag(cont []byte, tags []string) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var min int = math.MaxInt32
	var mint *target
	for _, tag := range tags {
		if v, ok := s.mp[tag]; ok {
			temN := v.Num()
			if v.Num() < min {
				min = temN
				mint = v
			}
		}
	}
	mint.BroadCastWithInnerJoinTag(cont, tags)
}

func (s *set) monitor() error {
	for {
		s.rw.RLock()
		for _, r := range s.mp {
			switch r.Status() {
			default:
				continue
			case TargetFLAGShouldEXTENSION:
				s.expansion <- r
			case TargetFLAGShouldReBalance:
				s.balance <- r
			case TargetFLAGShouldSHRINKS:
				s.shrinks <- r
			}
		}
		s.rw.RUnlock()
		time.Sleep(time.Duration(s.watchTime) * time.Second)
	}
}

func (s *set) handleMonitor() error {

	for {
		select {
		case t := <-s.expansion:
			t.Expansion()
		case t := <-s.shrinks:
			t.Shrinks()
		case t := <-s.balance:
			since := time.Now()
			t.Balance()
			escape := time.Since(since)
			logging.Infof("sim/wti : rebalance 耗费时间：%v ，在线人数为： %v", escape, t.Num())
		}
	}
}
