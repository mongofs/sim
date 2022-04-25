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
	"sync"
	"time"
)

var wti = newSet()

type set struct {
	// mp tagName =>
	mp   map[string]*target // wti => []string
	sort []*target
	rw   *sync.RWMutex

	flag      bool
	limit     int
	watchTime int
	expansion chan *target
	shrinks   chan *target
	balance   chan *target
}

func newSet() *set {
	return &set{
		mp:        map[string]*target{},
		rw:        &sync.RWMutex{},
		limit:     DefaultCapacity,
		watchTime: 20,
		expansion: make(chan *target),
		shrinks:   make(chan *target),
		balance:   make(chan *target),
	}
}

// ======================================API =================================

func (s *set) RegisterParallelFunc() []ParallelFunc {
	return s.parallel()
}

func (s *set) Add(tag string, client Client) (*target, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	s.rw.Lock()
	defer s.rw.Unlock()
	var res *target
	if target, ok := s.mp[tag]; ok {
		res = target
		target.Add(client)
	} else {
		ctag, err := NewTarget(tag, s.limit)
		if err != nil {
			return nil, err
		}
		s.mp[tag] = ctag
		res = ctag
	}
	return res, nil
}

func (s *set) Info(tag string) (*targetInfo, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	return s.info(tag)
}

func (s *set) List() []*targetInfo {
	return s.list()
}

func (s *set) BroadCastByTarget(msg map[string][]byte) ([]string, error) {
	if err := s.check(); err != nil {
		return nil, err
	}
	return s.broadcastByTag(msg)
}

func (s *set) BroadCastWithInnerJoinTag(cont []byte, tags []string) ([]string, error) {
	if err := s.check(); err != nil {
		return nil, err
	}

	res := s.broadcast(cont, tags...)
	return res, nil
}

// ====================================helper ==================================

func (s *set) list() []*targetInfo {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var res []*targetInfo
	for _, v := range s.mp {
		res = append(res, v.Info())
	}
	return res
}

func (s *set) info(tag string) (*targetInfo, error) {
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
		res = append(res, mint.BroadCastWithInnerJoinTag(cont, tags)...)
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
			case TargetFLAGShouldEXTENSION:
				s.expansion <- r
			case TargetFLAGShouldReBalance:
				s.balance <- r
			case TargetFLAGShouldSHRINKS:
				s.shrinks <- r
			}
		}
		s.rw.RUnlock()

	}
}

func (s *set) handleMonitor() error {
	var duration = 20 * time.Second
	t := time.NewTicker(duration)
	for {
		select {
		case <-t.C: // clear掉没有用户，且创建时间超过10min 的组
			s.clear()
		case t := <-s.expansion:
			t.Expansion()
		case t := <-s.shrinks:
			t.Shrinks()
		case t := <-s.balance:
			t.Balance()
		}
	}
}

func (s *set) clear() {
	s.rw.Lock()
	defer s.rw.Unlock()
	for k, v := range s.mp {
		if v.num == 0 && time.Now().Unix()-v.createTime > 60*2 {
			v.Destroy()
			delete(s.mp, k)
		}
	}
}
