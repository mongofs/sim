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

package main

import (
	"context"
	"google.golang.org/grpc"
	im "sim/api/v1"
	"sim/pkg/logging"
	"time"
)

const (
	Address           = "ws://127.0.0.1:8080/conn"
	DefaultRpcAddress = "127.0.0.1:8081"
)

func Client() im.BasicClient {
	conn, _ := grpc.Dial(DefaultRpcAddress, grpc.WithInsecure())
	return im.NewBasicClient(conn)
}


func main () {
	clis := Client()
	go monitor(clis)
	select {}
}

func monitor (cli im.BasicClient){
	time.Sleep(5 *time.Second)
	for {
		handleOnline(cli) // 用户在线
		handleWTITargetInfo(cli) //

	}
}


func handleOnline (cli im.BasicClient){
	res ,err := cli.Online(context.Background(),&im.Empty{})
	if err !=nil {
		logging.Error(err)
	}
	logging.Infof(" Number of people currently online %v" ,res.Number)
}

func handleWTITargetInfo (cli im.BasicClient) {
	res ,err := cli.WTITargetInfo(context.Background(),&im.WTITargetInfoReq{Tag: "man"})
	if err !=nil {
		logging.Error(err)
	}
	logging.Infof("")
}