package main

import (
	"fmt"
	"github.com/mongofs/sim"
	"github.com/mongofs/sim/pkg/conn"
	"net/http"
)

type talk struct {
	im   sim.ApiServer
	http * httpserver
}

func NewTalk() *talk {
	hk := &hooker{}
	im, err := sim.NewSIM(hk)
	if err != nil {
		panic(err)
	}
	ht := NewHTTP()
	return &talk{
		im:   im,
		http: ht,
	}
}

type hooker struct {
}

func (h hooker) Validate(token string) error {
	return nil
}

func (h hooker) ValidateFailed(err error, cli conn.Connect) {
	panic("implement me")
}

func (h hooker) ValidateSuccess(cli conn.Connect) {
	return
}

func (h hooker) HandleReceive(conn conn.Connect, data []byte) {
	return
}

func (h hooker) IdentificationHook(w http.ResponseWriter, r *http.Request) (string, error) {
	return r.Form.Get("token"),nil

}


func main(){
	tk := NewTalk()
	go func() {
		err := tk.im.Run()
		if err !=nil {
			panic(err)
		}
	}()


	err := tk.http.Run(tk.im.Upgrade)
	fmt.Println(err)
	panic(err)

}