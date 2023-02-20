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
	"context"
	"github.com/mongofs/sim/pkg/logging"
	"github.com/zhenjl/cityhash"
	"time"
)

//  init the bucket
func (s *sim) initBucket() {
	// prepare buckets
	s.bs = make([]bucketInterface, s.opt.ServerBucketNumber)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for i := 0; i < s.opt.ServerBucketNumber; i++ {
		s.bs[i] = NewBucket(s.opt,i)
	}
	logging.Infof("sim : init_bucket_number is %v ", s.opt.ServerBucketNumber)
	logging.Infof("sim : init_bucket_size  is %v ", s.opt.BucketSize)
}

func (s *sim) bucket(token string) bucketInterface {
	idx := s.routeBucket(token, uint32(s.opt.ServerBucketNumber))
	return s.bs[idx]
}


func (s *sim) monitorBucket(ctx context.Context) (string, error) {
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return "monitorBucket", nil
		case <-timer.C:
			var sum int64 = 0
			for _, v := range s.bs {
				sum += int64(v.Count())
			}
			s.ps.Store(sum)
			logging.Infof("sim : Current online number %v ", s.ps.Load())
		}
	}
}

func (s *sim) routeBucket(token string, size uint32) uint32 {
	return cityhash.CityHash32([]byte(token), uint32(len(token))) % size
}
