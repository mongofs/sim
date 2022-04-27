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
	"errors"
	"net/http"
	im "sim/api/v1"
	"sim/pkg/logging"
)

// 这里原则是初始化路由
func (h *httpserver) initRouter() {

	h.http.HandleFunc(h.validateRouter, h.connection)

	h.http.HandleFunc("/ping", func(writer http.ResponseWriter, request *http.Request) {
		remark := "/ping"
		if err := request.ParseForm(); err != nil {
			handleReturn(writer, nil, err, remark)
			return
		}
		res, err := h.handler.Ping(context.Background(), &im.Empty{})
		handleReturn(writer, res, err, remark)
	})

	h.http.HandleFunc("/online", func(writer http.ResponseWriter, request *http.Request) {
		remark := "/online"
		if err := request.ParseForm(); err != nil {
			handleReturn(writer, nil, err, remark)
			return
		}
		res, err := h.handler.Online(context.Background(), &im.Empty{})
		handleReturn(writer, res, err, remark)
	})

	h.http.HandleFunc("/broadcast", func(writer http.ResponseWriter, request *http.Request) {
		remark := "/broadcast"
		if err := request.ParseForm(); err != nil {
			handleReturn(writer, nil, err, remark)
			return
		}
		content := request.Form.Get("content")
		if content == "" {
			handleReturn(writer, nil, errors.New("param should have 'content' "), remark)
			return
		}
		res, err := h.handler.Broadcast(context.Background(), &im.BroadcastReq{Data: []byte(content)})
		handleReturn(writer, res, err, remark)
	})

	h.http.HandleFunc("/target/info", func(writer http.ResponseWriter, request *http.Request) {
		remark := "/target/info"
		if err := request.ParseForm(); err != nil {
			handleReturn(writer, nil, err, remark)
			return
		}
		tag := request.Form.Get("tag")
		if tag == "" {
			handleReturn(writer, nil, errors.New("param should have 'tag' "), remark)
			return
		}
		res, err := h.handler.WTITargetInfo(context.Background(), &im.WTITargetInfoReq{Tag: tag})
		handleReturn(writer, res, err, remark)
	})

	h.http.HandleFunc("/target/list", func(writer http.ResponseWriter, request *http.Request) {
		remark := "/target/list"
		if err := request.ParseForm(); err != nil {
			handleReturn(writer, nil, err, remark)
			return
		}
		res, err := h.handler.WTITargetList(context.Background(), &im.WTITargetListReq{})
		handleReturn(writer, res, err, remark)
	})

	h.http.HandleFunc("/target/broadcast", func(writer http.ResponseWriter, request *http.Request) {
		remark := "/target/broadcast"
		if err := request.ParseForm(); err != nil {
			handleReturn(writer, nil, err, remark)
			return
		}
		content := request.Form.Get("content")
		tag := request.Form.Get("tag")
		if tag == "" || content == "" {
			handleReturn(writer, nil, errors.New("param should have 'tag && content' "), remark)
			return
		}
		res, err := h.handler.WTIBroadcastByTarget(context.Background(),
			&im.WTIBroadcastReq{Data: map[string][]byte{
				tag: []byte(content),
			}})
		handleReturn(writer, res, err, remark)
	})

}

// todo  handle  http error
func handleReturn(w http.ResponseWriter, returnData interface{}, err error, remark string) {
	res := Response{
		w:      w,
		Desc:   "bad return ",
		Status: 200,
	}
	if err == nil {
		res.Desc = "ok  "
		res.Data = returnData
	} else {
		res.Data = err.Error()
	}

	logging.Infof("sim : http request , HTTPStatus is %v ,router : %v ", res.Status, remark)
	res.SendJson()
}
