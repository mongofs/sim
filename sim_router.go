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
	"errors"
	"net/http"
	"time"
)



func (s *ImSrever) initRouter()error{
	//分组创建路由
	s.http.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		res := &Response{
			w:      writer,
			Status: 403,
			Data:   "ok",
		}
		res.SendJson()
	})
	s.http.HandleFunc("/conn", s.Connection)
	return nil
}

// create  connection
func (s *ImSrever) Connection(writer http.ResponseWriter, request *http.Request){
	now :=time.Now()
	defer func() {
		escape := time.Since(now)
		s.opt.ServerLogger.Infof("im/router : %s create %s  cost %v  url is %v ", request.RemoteAddr,request.Method,escape,request.URL)
	}()

	res := &Response{
		w:      writer,
		Data:   nil,
		Status: 200,
	}
	if request.ParseForm() != nil {
		res.Status = 400
		res.Data = "connection is bad "
		res.SendJson()
		return
	}

	token:= request.Form.Get("token")
	if token == "" {
		res.Status=400
		res.Data = "token validate error"
		res.SendJson()
	}
	// validate token
	bs:= s.bucket(token)
	cli,err := bs.CreateConn(writer,request,token,s.opt.ServerReceive)
	if err !=nil {
		res.Status=400
		res.Data = err.Error()
		return
	}
	// validate failed
	if err := s.opt.ServerValidate.Validate(token);err !=nil {
		s.opt.ServerValidate.ValidateFailed(err,cli)
		return
	}else {
		//validate success
		s.opt.ServerValidate.ValidateSuccess(cli)
	}


	// register to data
	if err := bs.Register(cli,token);err !=nil {
		cli.Send([]byte(err.Error()))
		cli.Offline()
	}
}
