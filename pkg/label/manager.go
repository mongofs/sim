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
	"math"
	"sim/pkg/errors"
	"sync"
	"time"
)

type manager struct {
	// mp tagName =>
	mp   map[string]Label // wti => []string
	sort []Label
	rw   *sync.RWMutex

	flag                                 bool
	limit, watchTime                     int
	expansion, shrinks, balance, destroy chan Label
}


var m Manager


func Add (label string ,cli Client)(ForClient,error){
	return m.AddClient(label,cli)
}


func NewManager() Manager {
	if m ==nil {
		m = &manager{
			mp:        map[string]Label{},
			rw:        &sync.RWMutex{},
			limit:     DefaultCapacity,
			watchTime: 20,
			expansion: make(chan Label, 5),
			shrinks:   make(chan Label, 5),
			balance:   make(chan Label, 5),
		}
	}
	return m
}

// Run 将target的需要长时间运行的内容返回出去执行
func (s *manager) Run() []func() error{
	return s.parallel()
}

func (s *manager) AddClient(tag string, client Client) (ForClient, error) {
	if tag == "" || client == nil {
		return nil, errors.ErrBadParam
	}
	return s.add(tag, client)
}

func (s *manager) List(limit, page int) []*LabelInfo {
	return s.list()
}

func (s *manager) LabelInfo(label string) (*LabelInfo, error) {
	if label == "" {
		return nil, errors.ErrBadParam
	}
	return s.LabelInfo(label)
}

func (s *manager) BroadCastByLabel(tc map[string][]byte) ([]string, error) {
	if len(tc) == 0 {
		return nil, errors.ErrBadParam
	}
	return s.broadcastByLabel(tc)
}

func (s *manager) BroadCastWithInnerJoinLabel(cont []byte, tags []string) ([]string, error) {
	if len(cont) == 0 || len(tags) == 0 {
		return nil, errors.ErrBadParam
	}
	return s.broadcast(cont, tags...), nil
}

// Add 添加用户到某个target 上去，此时用户需要在用户单元保存target内容
func (s *manager) add(tag string, client Client) (ForClient, error) {
	s.rw.Lock()
	defer s.rw.Unlock()
	var res ForClient
	if tg, ok := s.mp[tag]; ok {
		res = tg
		tg.Add(client)
	} else {
		ctag, err := NewLabel(tag, s.limit)
		if err != nil {
			return nil, err
		}
		s.mp[tag] = ctag
		res = ctag
	}
	return res, nil
}

func (s *manager) list() []*LabelInfo {
	s.rw.RLock()
	defer s.rw.RUnlock()
	var res []*LabelInfo
	for _, v := range s.mp {
		res = append(res, v.Info())
	}
	return res
}

func (s *manager) info(tag string) (*LabelInfo, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if v, ok := s.mp[tag]; ok {
		return v.Info(), nil
	}
	return nil, errors.ERRWTITargetNotExist
}

func (s *manager) broadcast(cont []byte, tags ...string) (res []string) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	if len(tags) != 0 {
		var min int = math.MaxInt32
		var mintg Label
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

func (s *manager) broadcastByLabel(msg map[string][]byte) ([]string, error) {
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

func (s *manager) parallel()(res []func() error) {
	res = append(res, s.monitor, s.handleMonitor)
	return
}

func (s *manager) monitor() error {
	for {
		time.Sleep(time.Duration(s.watchTime) * time.Second)
		s.rw.RLock()
		for k, r := range s.mp {
			st := r.Status()
			switch st {
			default:
				continue
			case TargetStatusShouldEXTENSION:
				s.expansion <- r
			case TargetStatusShouldReBalance:
				s.balance <- r
			case TargetStatusShouldSHRINKS:
				s.shrinks <- r
			case TargetStatusShouldDestroy:
				delete(s.mp,k)
				r.Destroy()
			}
		}
		s.rw.RUnlock()
	}
}

func (s *manager) handleMonitor() error {
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



