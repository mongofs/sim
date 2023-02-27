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
	"fmt"
	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
	"github.com/zhenjl/cityhash"
	"go.uber.org/zap"
	"time"
)

//  init the bucket
func (s *sim) initBucket() {
	// prepare buckets
	s.bs = make([]bucketInterface, s.opt.ServerBucketNumber)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	for i := 0; i < s.opt.ServerBucketNumber; i++ {
		s.bs[i] = NewBucket(s.opt, i, s.ctx)
	}

	logging.Log.Info("initBucket", zap.Int("BUCKET_NUMBER", s.opt.ServerBucketNumber))
	logging.Log.Info("initBucket", zap.Int("BUCKET_SIZE", s.opt.BucketSize))
}

func (s *sim) bucket(token string) bucketInterface {
	idx := s.routeBucket(token, uint32(s.opt.ServerBucketNumber))
	return s.bs[idx]
}

func (s *sim) monitorBucket(ctx context.Context) (string, error) {
	var interval = 10
	var dataMonitorInterval = 60
	dataMonitorTimer := time.NewTimer(time.Duration(dataMonitorInterval) * time.Second)
	timer := time.NewTicker(time.Duration(interval) * time.Second)
	logging.Log.Info("monitorBucket ", zap.Int("MONITOR_ONLINE_INTERVAL", interval))
	for {
		select {
		case <-ctx.Done():
			return "monitorBucket", nil
		case <-timer.C:
			var sum int64 = 0
			for _, v := range s.bs {
				sum += int64(v.Count())
			}
			s.num.Store(sum)
			if s.opt.debug == true {
				// you get get the pprof ,
				pprof := fmt.Sprintf("http://127.0.0.1%v/debug/pprof", s.opt.PProfPort)

				logging.Log.Info("monitorBucket ", zap.Int64("ONLINE", s.num.Load()), zap.String("PPROF", pprof))
			} else {
				logging.Log.Info("monitorBucket ", zap.Int64("ONLINE", s.num.Load()))
			}
		case <-dataMonitorTimer.C:
			content, loseContent, contentLength := conn.SwapSendData()
			logging.Log.Info("monitorBucket",
				zap.Int64("COUNT_LOSE_CONTENT", loseContent),
				zap.Int64("COUNT_CONTENT", content),
				zap.Int64("COUNT_CONTENT_LEN(Byte)", contentLength),
				zap.Int64("COUNT_CONTENT_LEN(KB)", contentLength/1024),
				zap.Int64("COUNT_CONTENT_LEN(MB)", contentLength/1024/1024))
		}
	}
}

func (s *sim) routeBucket(token string, size uint32) uint32 {
	return cityhash.CityHash32([]byte(token), uint32(len(token))) % size
}
