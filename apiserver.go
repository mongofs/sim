package sim

import (
	"context"
	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
	"go.uber.org/atomic"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ApiServer interface {
	Run() error
	Ping() string
	Online() int
	SendMessage(message []byte, users []string)
	Upgrade(w http.ResponseWriter, r *http.Request) error
}

type sim struct {
	bs     []bucketInterface
	ps     atomic.Int64
	cancel context.CancelFunc
	ctx    context.Context
	opt    *Options
}

func NewSIM(hooker Hooker, opts ...OptionFunc) (ApiServer ApiServer, err error) {
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
		ps:  atomic.Int64{},
		opt: options,
	}
	b.ps.Store(0)
	b.initBucket() // init bucket plugin
	return b, nil
}

type HandleUpgrade func(w http.ResponseWriter, r *http.Request) error

func (s *sim) Run() error {
	if s.opt.ServerDiscover != nil {
		s.opt.ServerDiscover.Register()
		defer func() {
			s.opt.ServerDiscover.Deregister()
		}()
	}

	parallelTask, finishChannel := s.Parallel()
	sigs := make(chan os.Signal, 1)
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
func (s *sim) Ping() string {
	return ""
}
func (s *sim) Online() int {
	return 1
}
func (s *sim) Upgrade(w http.ResponseWriter, r *http.Request) error {
	return s.upgrade(w, r)
}
func (s *sim) SendMessage(message []byte, users []string) {
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
	if bucketId,userNum ,err := bs.Register(cli); err != nil {
		cli.Send([]byte(err.Error()))
		cli.Close()
		return err
	}else{
		logging.Infof("sim : %v connected ,bucket : %v ,number : %v", cli.Identification(),bucketId,userNum)
		return nil
	}
}

func (s *sim) close() error {
	s.cancel()
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
