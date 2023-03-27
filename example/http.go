package main

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"net"
	"net/http"
	"sim"
	"sim/pkg/logging"
)

type httpserver struct {
	http *http.ServeMux
	port string
}

type Response struct {
	w      http.ResponseWriter
	Desc   string      `json:"desc"`
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func NewHTTP() *httpserver {
	return &httpserver{
		http: &http.ServeMux{},
		port: ":3306",
	}
}

func (r *Response) SendJson() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Status)
	return r.w.Write(resp)
}

func (s *httpserver) Run(upgrade sim.HandleUpgrade) error {
	s.http.HandleFunc("/conn", func(writer http.ResponseWriter, r *http.Request) {
		//now := time.Now()
		defer func() {
			//escape := time.Since(now)
			//logging.Infof("sim : %s create %s  cost %v  url is %v ", r.RemoteAddr, r.Method, escape, r.URL)
		}()
		res := &Response{
			w:      writer,
			Data:   nil,
			Status: 200,
		}
		if r.ParseForm() != nil {
			res.Status = 400
			res.Data = "connection is bad "
			res.SendJson()
			return
		}

		token := r.Form.Get("token")
		if token == "" {
			res.Status = 400
			res.Data = "token validate error"
			res.SendJson()
		}

		// upgrade connection

		if err := upgrade(writer, r); err != nil {
			res.Status = 400
			res.Data = "upgrade failed"
			res.SendJson()
		}
	})
	return s.run(context.Background())
}

func (s *httpserver) run(ctx context.Context) error {
	listen, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer listen.Close()
	if err != nil {
		logging.Log.Info("run", zap.Error(err))
		return err
	}
	logging.Log.Info("run", zap.String("PORT", s.port))
	if err := http.Serve(listen, s.http); err != nil {
		return err
	}
	return nil
}
