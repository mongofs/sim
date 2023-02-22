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
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
	"go.uber.org/atomic"
)

const (
	RunStatusRunning = 1 + iota
	RunStatusStopped

)

type sim struct {
	// this is the slice of bucket , the bucket implement you can see ./bucket.go
	// or github/mongofs/sim/bucket.go . for avoid the big locker , the specific
	// implement use hash crc13 , so you don't worry about the matter of performance
	bs     []bucketInterface

	// this is the counter of online User, there have a goroutine to provide the
	// precision of online people
	num     atomic.Int64

	// this is function to notify all goroutine exit
	cancel context.CancelFunc
	ctx    context.Context

	// this parameter is for judge sim status ( running or not )
	stat uint

	// this is the option about sim ,you can see ./option.go or github.com/mongofs/sim/option.go
	// you can use the function provided by option.go to set the parameters
	opt    *Options
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
	errServerIsNotRunning =errors.New("the server is not running ")
	errServerIsRunning =errors.New("the server is  running ")

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
	var (
		logger logging.Logger
	)
	if options.LogPath != "" {
		logging.FlushLogPath(options.LogPath, "test", logging.OutputStdout)
	} else {
		logger = logging.GetDefaultLogger()
	}
	if options.Logger == nil {
		options.Logger = logger
	}

	b := &sim{
		num:  atomic.Int64{},
		opt: options,
		stat: RunStatusStopped,
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
	if stalk.stat != RunStatusStopped {
		// that is mean the sim not run
		return errServerIsRunning
	}
	return stalk.run()
}

func SendMessage(msg []byte, Users []string) error {
	if stalk == nil {
		return errInstanceIsNotExist
	}
	if stalk.stat != RunStatusRunning {
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
	if stalk.stat != RunStatusRunning {
		// that is mean the sim not run
		return errServerIsNotRunning
	}
	return stalk.upgrade(w, r)
}

type HandleUpgrade func(w http.ResponseWriter, r *http.Request) error


func (s *sim) run() error {
	if s.opt.ServerDiscover != nil {
		s.opt.ServerDiscover.Register()
		defer func() {
			s.opt.ServerDiscover.Deregister()
		}()
	}

	parallelTask, finishChannel := s.Parallel()
	sigs := make(chan os.Signal, 1)
	s.stat = RunStatusRunning
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for {
		select {
		case <-finishChannel:
			logging.Infof("sim : exit -1 ")
			return nil
		case finishMark := <-parallelTask:
			logging.Infof("sim : %v parallel task is out ", finishMark)
		case <-sigs:
			s.close()
		}
	}
	return nil
}
func (s *sim) online() int {
	return 1
}
func (s *sim) sendMessage(message []byte, users []string) {
	s.bs[0].SendMessage([]byte("123"))
	return
}
func (s *sim) upgrade(w http.ResponseWriter, r *http.Request) error {
	// this is plugin need the coder to implement it
	identification, err := s.opt.hooker.IdentificationHook(w, r)
	if err != nil {
		return err
	}
	bs := s.bucket(identification)
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
		cli.Close()
		return err
	} else {
		logging.Infof("sim : %v connected ,bucket : %v ,number : %v", cli.Identification(), bucketId, userNum)
		return nil
	}
}

func (s *sim) close() error {
	s.cancel()
	s.stat = RunStatusStopped
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
				logging.Errorf("sim/apiServer.go : parallel task occurred Panic %v", err)
			}
		}()
		wg := sync.WaitGroup{}
		for _, v := range prepareParallelFunc {
			wg.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logging.Errorf("sim/apiServer.go : parallel task occurred Panic %v", err)
					}
				}()
				mark, err := v(s.ctx)
				if err != nil {
					logging.Errorf("sim/apiServer.go : parallel task occurred error %v", err)
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
