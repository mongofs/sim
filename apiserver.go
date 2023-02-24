/* Copyright 2022 steven
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
	"errors"
	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

const (
	RunStatusRunning = 1 + iota
	RunStatusStopped
)

type sim struct {
	// this is the slice of bucket , the bucket implement you can see ./bucket.go
	// or github/mongofs/sim/bucket.go . for avoid the big locker , the specific
	// implement use hash crc13 , so you don't worry about the matter of performance
	bs []bucketInterface

	// this is the counter of online User, there have a goroutine to provide the
	// precision of online people
	num atomic.Int64

	// this is function to notify all goroutine exit
	cancel context.CancelFunc
	ctx    context.Context

	// this parameter is for judge sim status ( running or not )
	running uint


	// this is the option about sim ,you can see ./option.go or github.com/mongofs/sim/option.go
	// you can use the function provided by option.go to set the parameters
	opt *Options
}

var (
	stalk *sim
	once  sync.Once
)

var (
	// make sure the resource not exist
	errInstanceIsExist = errors.New("the instance is existed ")
	// make sure the resource  exist
	errInstanceIsNotExist = errors.New("the instance is not existed ")
	// the server is not Running
	errServerIsNotRunning = errors.New("the server is not running ")
	errServerIsRunning    = errors.New("the server is  running ")

	// hook is nil
	errHookIsNil = errors.New("hook is nil ")
)

func NewSIMServer(hooker Hooker, opts ...OptionFunc) error {
	if hooker == nil {
		panic("hook is nil ")
	}
	if stalk != nil {
		return errInstanceIsExist
	}

	options := LoadOptions(hooker, opts...)
	b := &sim{
		num:     atomic.Int64{},
		opt:     options,
		running: RunStatusStopped,
	}
	// logger
	{
		var loggingOps = []logging.OptionFunc{
			logging.SetLevel(logging.InfoLevel),
			logging.SetLogName("sim"),
			logging.SetLogPath("./log"),
		}
		logging.InitZapLogger(b.opt.debug, loggingOps...)
	}

	b.num.Store(0)
	b.initBucket() // init bucket plugin

	stalk = b
	return nil
}

// This function is
func Run() error {
	if stalk == nil {
		return errInstanceIsNotExist
	}
	if stalk.running != RunStatusStopped {
		// that is mean the sim not run
		return errServerIsRunning
	}
	return stalk.run()
}

func SendMessage(msg []byte, Users []string) error {
	if stalk == nil {
		return errInstanceIsNotExist
	}
	if stalk.running != RunStatusRunning {
		// that is mean the sim not run
		return errServerIsNotRunning
	}
	stalk.sendMessage(msg, Users)
	return nil
}

func Upgrade(w http.ResponseWriter, r *http.Request) error {
	if stalk == nil {
		return errInstanceIsNotExist
	}
	if stalk.running != RunStatusRunning {
		// that is mean the sim not run
		return errServerIsNotRunning
	}
	return stalk.upgrade(w, r)
}

func Stop() error {
	if stalk == nil {
		return errInstanceIsNotExist
	}
	if stalk.running != RunStatusRunning {
		// that is mean the sim not run
		return errServerIsNotRunning
	}
	stalk.close()
	time.Sleep(200 * time.Millisecond)
	return nil
}


type HandleUpgrade func(w http.ResponseWriter, r *http.Request) error

func (s *sim) pprof()error {
	if s.opt.debug {
		if stalk == nil {
			return errInstanceIsNotExist
		}
		if stalk.running != RunStatusRunning {
			// that is mean the sim not run
			return errServerIsNotRunning
		}
		go func() {
			http.ListenAndServe(s.opt.PProfPort, nil)
		}()
	}
	return nil
}

func (s *sim) run() error {
	if s.opt.ServerDiscover != nil {
		s.opt.ServerDiscover.Register()
		defer func() {
			s.opt.ServerDiscover.Deregister()
		}()
	}
	parallelTask, finishChannel := s.Parallel()
	s.running = RunStatusRunning
	go func() {
		defer func() {
			if err := recover(); err != nil {

			}
		}()

		// monitor the channel and log the out information
		for {
			select {
			case <-finishChannel:
				logging.Log.Info("sim : exit -1 ")
				return
			case finishMark := <-parallelTask:
				logging.Log.Info("task finish",zap.String("FINISH_TASK",finishMark))
			}
		}
	}()
	if err :=s.pprof();err!=nil{
		// todo
		panic(err)
	}
	return nil
}

func (s *sim) online() int {
	return int(s.num.Load())
}

// because there is no parallel problem in slice when you read the data
// and there is no any operate action on bucket slice ,so not use locker
func (s *sim) sendMessage(message []byte, users []string) {
	if len(users) != 0 {
		for _, user := range users {
			bs := s.bucket(user)
			bs.SendMessage(message, user)
		}
		return
	}
	// because there is no parallel problem in slice when you read the data
	// and there is no any operate action on bucket slice ,so not use locker
	for _, bucket := range s.bs {
		bucket.SendMessage(message)
	}
	return
}
func (s *sim) upgrade(w http.ResponseWriter, r *http.Request) error {
	// this is plugin need the coder to implement it
	identification, err := s.opt.hooker.IdentificationHook(w, r)
	if err != nil {
		return err
	}
	bs := s.bucket(identification)

	// try to close the same identification device
	bs.Offline(identification)
	sig := bs.SignalChannel()
	cli, err := conn.NewConn(identification, sig, w, r, s.opt.hooker.HandleReceive)
	if err != nil {
		return err
	}
	if err := s.opt.hooker.Validate(identification); err != nil {
		s.opt.hooker.ValidateFailed(err, cli)
		return nil
	} else {
		s.opt.hooker.ValidateSuccess(cli)
	}
	if bucketId, userNum, err := bs.Register(cli); err != nil {
		cli.Send([]byte(err.Error()))
		cli.Close("register to bucket error ")
		return err
	} else {
		logging.Log.Info("upgrade", zap.String("ID", cli.Identification()), zap.String("BUCKET_ID", bucketId), zap.Int64("BUCKET_ONLINE", userNum))
		return nil
	}
}

func (s *sim) close() error {
	s.cancel()
	s.running = RunStatusStopped
	return nil
}

func (s *sim) Parallel() (chan string, chan string) {
	var prepareParallelFunc = []func(ctx context.Context) (string, error){
		s.monitorBucket,
	}
	monitor, closeCn := make(chan string), make(chan string)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logging.Log.Error("Parallel", zap.Any("PANIC", err))
			}
		}()
		wg := sync.WaitGroup{}
		for _, v := range prepareParallelFunc {
			wg.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logging.Log.Error("Parallel", zap.Any("PANIC", err))
					}
				}()
				mark, err := v(s.ctx)
				if err != nil {
					logging.Log.Error("Parallel task", zap.Error(err))
					return
				}
				monitor <- mark
				wg.Add(-1)
				return
			}()
		}
		wg.Wait()
		close(closeCn)
	}()
	return monitor, closeCn
}
