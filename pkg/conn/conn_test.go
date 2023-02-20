package conn

import (
	"fmt"
	"sync"
	"testing"
	"unsafe"
)

func TestNewConn(t *testing.T) {

	// the  limit is not the struct of connection or any other struct
	// 72 bit  , 10,000,000 connections need 1GB memory to load ,is very
	// small
	fmt.Println(unsafe.Sizeof(conn{
		once:           sync.Once{},
		con:            nil,
		identification: "",
		buffer:         make(chan []byte, 8),
		heartBeatTime:  0,
		closeChan:      nil,
		messageType:    0,
	}))
}
