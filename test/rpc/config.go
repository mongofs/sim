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
	concurrency int
	requests    int
	keepTime    int
	host        string
)


func init() {
	pflag.IntVarP(&concurrency, "concurrency", "c", 1, "在测试会话中所执行的请求总个数，默认仅执行一个请求")
	pflag.IntVarP(&requests, "requests", "n", 1, "每次请求的并发数，相当于同时模拟多少个人访问url，默认是一次一个")
	pflag.IntVarP(&keepTime, "keepTime", "k", 50000, "测试所进行的最大秒数。其内部隐含值是-n 50000")
	pflag.StringVarP(&host, "host", "h", "", "设置对应的Url ，比如 -h 127.0.0.1:8080,目前展示不能支持配置链接token，后续会加上")
}

type config struct {
	concurrency int
	requests    int
	keepTime    int
	host        string
}

// InitConfig 实例化
func InitConfig() *config {
	return &config{
		concurrency: concurrency,
		requests:    requests,
		keepTime:    keepTime,
		host:        host,
	}
}
