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
	"encoding/json"
	"net"
	"net/http"
	"sim/pkg/logging"
	"time"
)

type httpserver struct {
	http        *http.ServeMux
	port        string
	validateKey string
	validateRouter string

	up      Upgrade
	handler API
}

type Response struct {
	w      http.ResponseWriter
	Desc   string      `json:"desc"`
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func (r *Response) SendJson() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Status)
	return r.w.Write(resp)
}

type Upgrade func(writer http.ResponseWriter, r *http.Request, token string) error

func newHttpServer(validateKey,validateRouter, port string, up Upgrade, api API) *httpserver {
	hs := &httpserver{
		http:        http.NewServeMux(),
		port:        port,
		validateKey: validateKey,
		validateRouter: validateRouter,
		up:          up,
		handler:     api,
	}
	return hs
}

func (s *httpserver) Run() ParallelFunc {
	s.initRouter()
	return s.run
}

func (s *httpserver) connection(writer http.ResponseWriter, r *http.Request) {
	now := time.Now()
	defer func() {
		escape := time.Since(now)
		logging.Infof("sim : %s create %s  cost %v  url is %v ", r.RemoteAddr, r.Method, escape, r.URL)
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

	token := r.Form.Get(s.validateKey)
	if token == "" {
		res.Status = 400
		res.Data = "token validate error"
		res.SendJson()
	}

	if err := s.up(writer, r, token); err != nil {
		logging.Error(err)
	}
}

func (s *httpserver) run() error {
	listen, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	logging.Infof("sim : start HTTP example at %s ", s.port)
	if err := http.Serve(listen, s.http); err != nil {
		return err
	}
	return nil
}
