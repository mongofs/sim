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
	"sim/pkg/errors"
	"sim/pkg/target"
	"sync"
	"time"
)

var DefaultCapacity = 128

var wti = newSet()

type set struct {
	// mp tagName =>
	mp   map[string]WTIManager // wti => []string
	sort []WTIManager
	rw   *sync.RWMutex

	flag             bool
	limit, watchTime int
	expansion, shrinks, balance, destroy chan WTIManager
}

func newSet() *set {
	return &set{
		mp:        map[string]WTIManager{},
		rw:        &sync.RWMutex{},
		limit:     DefaultCapacity,
		watchTime: 20,
		expansion: make(chan WTIManager),
		shrinks:   make(chan WTIManager),
		balance:   make(chan WTIManager),
		destroy:   make(chan WTIManager),
	}
}

func (s *set) run() []ParallelFunc {
	return s.parallel()
}

// Add 添加用户到某个target 上去，此时用户需要在用户单元保存target内容
func (s *set) add(tag string, client Client) (target.ClientManager, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	s.rw.Lock()
	defer s.rw.Unlock()
	var res WTIManager
	if tg, ok := s.mp[tag]; ok {
		res = tg
		tg.Add(client)
	} else {
		ctag, err := target.NewTarget(tag, s.limit)
		if err != nil {
			return nil, err
		}
		s.mp[tag] = ctag
		res = ctag
	}
	return res, nil
}

func (s *set) list() []*target.TargetInfo {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var res []*target.TargetInfo
	for _, v := range s.mp {
		res = append(res, v.Info())
	}
	return res
}

func (s *set) info(tag string) (*target.TargetInfo, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if v, ok := s.mp[tag]; ok {
		return v.Info(), nil
	}
	return nil, errors.ERRWTITargetNotExist
}

func (s *set) broadcast(cont []byte, tags ...string) (res []string) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if len(tags) != 0 {
		var min int = math.MaxInt32
		var mintg WTIManager
		for _, tag := range tags {
			if v, ok := s.mp[tag]; ok {
				temN := v.Count()
				if v.Count() < min {
					min = temN
					mintg = v
				}
			}
		}
		res = append(res, mintg.BroadCast(cont, tags...)...)
		return
	}
	for _, v := range s.mp {
		res = append(res, v.BroadCast(cont)...)
	}
	return
}

func (s *set) broadcastByTag(msg map[string][]byte) ([]string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var res []string
	for tagN, cont := range msg {
		if tar, ok := s.mp[tagN]; ok {
			res = append(res, tar.BroadCast(cont)...)
		}
	}
	return res, nil
}

func (s *set) parallel() (res []ParallelFunc) {
	res = append(res, s.monitor, s.handleMonitor)
	return
}

func (s *set) check() error {
	return nil
}

func (s *set) monitor() error {
	for {
		time.Sleep(time.Duration(s.watchTime) * time.Second)
		s.rw.RLock()
		for _, r := range s.mp {
			st := r.Status()
			switch st {
			default:
				continue
			case target.TargetStatusShouldEXTENSION:
				s.expansion <- r
			case target.TargetStatusShouldReBalance:
				s.balance <- r
			case target.TargetStatusShouldSHRINKS:
				s.shrinks <- r
			case target.TargetStatusShouldDestroy:

			}
		}
		s.rw.RUnlock()
	}
}

func (s *set) handleMonitor() error {
/*	var duration = 20 * time.Second
	t := time.NewTicker(duration)*/
	for {
		select {
		case t := <-s.expansion:
			t.Expansion()
		case t := <-s.shrinks:
			t.Shrinks()
		case t := <-s.balance:
			t.Balance()
		}
	}
}
