package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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


func (h hooker) Offline(cli conn.Connect,ty int )  {
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
	fmt.Println(string(data))
	return
}

func (h hooker) IdentificationHook(w http.ResponseWriter, r *http.Request) (string, error) {
	return r.Form.Get("token"), nil

}

func main() {
	sim.NewSIMServer(hooker{})
	tk := &talk{http: NewHTTP()}
	if err := sim.Run(); err != nil {
		panic(err)
	}
	go func() {
		if err := tk.http.Run(sim.Upgrade);err!=nil {
			panic(err)
		}
	}()
	sig := make(chan os.Signal, 1)

	// Free up resources by monitoring server interrupt to instead of killing the process id
	// graceful shutdown
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case sig := <-sig:
		logging.Infof("sim : close signal : %v", sig)
		if err := sim.Stop(); err != nil {
			panic(err)
		}
		break
	}
}
