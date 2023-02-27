package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mongofs/sim"
	"github.com/mongofs/sim/pkg/conn"
	"github.com/mongofs/sim/pkg/logging"
)

type talk struct {
	http *httpserver
}

type hooker struct {
}

func (h hooker) Validate(token string) error {
	return nil
}

func (h hooker) Offline(cli conn.Connect, ty int) {
	if ty == sim.OfflineBySqueezeOut {
		cli.Send([]byte("您已经被挤掉了"))
	}
}

func (h hooker) ValidateFailed(err error, cli conn.Connect) {
	panic("implement me")
}

func (h hooker) ValidateSuccess(cli conn.Connect) {
	return
}

func (h hooker) HandleReceive(conn conn.Connect, data []byte) {

	conn.Send([]byte("你好呀"))

	conn.ReFlushHeartBeatTime()
	//fmt.Println(string(data))
	return
}

func (h hooker) IdentificationHook(w http.ResponseWriter, r *http.Request) (string, error) {
	return r.Form.Get("token"), nil

}
func test() {
	fmt.Println(len([]byte("1234567890"))) // 10
}

func main() {
	sim.NewSIMServer(hooker{}, sim.WithServerDebug())
	tk := &talk{http: NewHTTP()}
	if err := sim.Run(); err != nil {
		panic(err)
	}
	go func() {
		if err := tk.http.Run(sim.Upgrade); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			time.Sleep(5000 * time.Millisecond)
			err := sim.SendMessage([]byte("1234567890"), []string{})
			if err != nil {
				fmt.Println(err)
			}
			//fmt.Println("一个循环")
		}
	}()
	sig := make(chan os.Signal, 1)

	// Free up resources by monitoring server interrupt to instead of killing the process id
	// graceful shutdown
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case sig := <-sig:
		logging.Log.Info("main", zap.Any("SIG", sig))
		if err := sim.Stop(); err != nil {
			panic(err)
		}
		break
	}
}
