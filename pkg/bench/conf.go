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

import "github.com/spf13/pflag"

var (
	concurrency    int    // 并发请求数量 ，比如 -c 100 ，代表每秒创建100个链接
	number         int    // 总的请求数量 ，比如 -n 10000,代表总共建立链接10000个
	keepTime       int    // 总的在线时长，比如 -k 100 ,代表在线时间为100秒，100秒后就会释放
	heartBeat      int    // 是否上报心跳，如果设置存在值，就会按照默认的结构体进行{test :1 } 进行心跳发送
	monitorPrint   int    // 设置当前状态打印间隔，默认10s
	host           string // 设置对应的Url ，比如 -h 127.0.0.1:8080,目前展示不能支持配置链接token，后续会加上
	identification string // 设置对应的identification的key值，比如 identification =token,表示你的token 才是唯一标识
)

// 整体流程： connection  -c -
func init() {
	pflag.IntVarP(&concurrency, "concurrency", "c", 100, "并发请求数量 ，比如 -c 100 ，代表每秒创建100个链接")
	pflag.IntVarP(&number, "number", "n", 10000, "总的请求数量 ，比如 -n 10000,代表总共建立链接10000个")
	pflag.IntVarP(&heartBeat, "heartBeat", "b", 50, "是否上报心跳，如果设置存在值，就会按照默认的结构体{test :1}按设定时间进行心跳发送 ")
	pflag.IntVarP(&keepTime, "keepTime", "k", 0, "总的在线时长，比如 -k 100s ,代表在线时间为100秒，100秒后就会释放，为0不释放")
	pflag.IntVarP(&monitorPrint, "monitor", "m", 10, "设置当前状态打印间隔，默认10s")
	pflag.StringVarP(&host, "host", "h", "ws://127.0.0.1:3306/conn", "设置对应的Url")
	pflag.StringVarP(&identification, "identification", "i", "token", "标识，比如 identification =token,表示你的token 才是唯一标识")
	pflag.Parse()
}

type Config struct {
	Concurrency    int
	Number         int
	KeepTime       int
	Host           string
	HeartBeat      int
	Monitor        int
	Identification string
}

// InitConfig 实例化
func InitConfig() *Config {
	return &Config{
		Concurrency:    concurrency,
		Number:         number,
		KeepTime:       keepTime,
		Host:           host,
		HeartBeat:      heartBeat,
		Monitor: monitorPrint,
		Identification: identification,
	}
}
