package main

import (
	"time"
	"fmt"
)

func main() {
	config := InitConfig()
	RunServer(config)
	if config.KeepTime == 0 {
		select {}
	}else {
		time.Sleep(time.Duration(config.KeepTime) * time.Second)
	}
	fmt.Println("exit process")
}


// RunServer 启动服务
func RunServer(cof *Config) {
	sb := NewBench(cof)
	sb.Run()
}