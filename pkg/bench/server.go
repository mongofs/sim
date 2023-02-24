package main

import (
	"fmt"
	print2 "github.com/mongofs/sim/pkg/print"

	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"math/rand"
	"time"
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
)

const (
	StageOFNewConnection = iota + 1
	StageOfFinishCreateConnection
)

type Bench struct {

	// metric count
	success atomic.Int32
	fail    atomic.Int32
	online  atomic.Int32
	retry   atomic.Int32

	// message count
	singleMessageCount  atomic.Int64
	allUserMessageCount atomic.Int64
	heartBeat           int
	stage               int

	oToken string //output token
	config *Config

	closeMonitor chan string
}

var url string

func NewBench(conf *Config) *Bench {
	return &Bench{
		success:             atomic.Int32{},
		fail:                atomic.Int32{},
		online:              atomic.Int32{},
		retry:               atomic.Int32{},
		singleMessageCount:  atomic.Int64{},
		allUserMessageCount: atomic.Int64{},
		heartBeat:           180,
		stage:               StageOFNewConnection,
		oToken:              "BaseToken",
		config:              conf,
		closeMonitor:        make(chan string, 10),
	}
}

func (s *Bench) Run() {
	tokens := s.getValidateKey()
	url = fmt.Sprintf("%s?%s=", s.config.Host, s.config.Identification)
	print2.PrintWithColor(fmt.Sprintf("CONFIG-PRINT-URL : url is '%s'", url), print2.FgGreen)
	print2.PrintWithColor(fmt.Sprintf("CONFIG-PRINT-HEARTBEAT : heartbeat interval  is  '%vs' ", s.heartBeat), print2.FgGreen)
	print2.PrintWithColor(fmt.Sprintf("CONFIG-PRINT-KEEPTIME : keep time is  '%vs' ", s.heartBeat), print2.FgGreen)
	print2.PrintWithColor(fmt.Sprintf("CONFIG_PRINT_ONLINE_TEST : '%s' BaseToken is online", identification), print2.FgGreen)
	print2.PrintWithColorAndSpace(fmt.Sprintf("=====================================CONFIG_IS_UP ================================"), print2.FgYellow, 1, 1)

	// create the based connection
	go func() {
		if err := s.CreateClient(s.oToken); err != nil {
			panic(err)
		}
	}()
	s.Batch(tokens)
	s.monitor()
}

var limiter = time.NewTicker(50 * time.Microsecond)

func (s *Bench) Batch(tokens []string) {
	for k, v := range tokens {
		if k%s.config.Concurrency == 0 && k != 0 {
			time.Sleep(1000 * time.Millisecond)
			if s.success.Load() != 0 {
				print2.PrintWithColor(fmt.Sprintf("Current_Status: Online %v ,Fail %v ", s.success.Load(), s.fail.Load()), print2.FgGreen)
			}
		}
		go s.CreateClient(v)
	}

	// because the interval of for loop exist the sleep 1 millisecond
	time.Sleep(1001 * time.Millisecond)
	print2.PrintWithColor(fmt.Sprintf("Current_Status: Online %v ,Fail %v ", s.success.Load(), s.fail.Load()), print2.FgGreen)

	s.stage = StageOfFinishCreateConnection
	print2.PrintWithColorAndSpace(fmt.Sprintf(" =====================================Created_Connection_Situation_IS_UP==========================================="), print2.FgYellow, 1, 1)
}

func (s *Bench) CreateClient(identification string) error {
	<-limiter.C
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url+identification, nil)
	if err != nil {
		s.fail.Inc()
		fmt.Printf("error occurs during runtime id : %v, url : %s ,err :%s\r\n", "ddd", url, err.Error())
		return err
	} else {
		s.success.Inc()
		s.online.Inc()
	}
	defer conn.Close()

	go func() {
		for {
			time.Sleep(time.Duration(s.heartBeat) * time.Second)
			conn.WriteJSON("{test: 1}")
		}
	}()

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic Occre  : %v \n", err)
		}
	}()

	for {
		messageType, messageData, err := conn.ReadMessage()
		if err != nil {
			s.closeMonitor <- identification
			fmt.Printf("Connection_Read_Err : %v\r\n", err)
		}
		s.allUserMessageCount.Inc()
		if identification == s.oToken {
			switch messageType {
			case websocket.TextMessage:
				if s.stage == StageOfFinishCreateConnection {
					s.singleMessageCount.Inc()
					print2.PrintWithColor(fmt.Sprintf("BaseToken_Receive  : %v", string(messageData)), print2.FgBlue)
				}
			case websocket.BinaryMessage:
			default:
			}
		}
	}
	return nil
}

func (s *Bench) monitor() {
	go func() {
		t := time.NewTicker(time.Duration(s.config.Monitor) * time.Second)
		for {
			select {
			case <-t.C:
				str := fmt.Sprintf("Current_Status: Online_%v ,Retry_%v, Msg_Count_%v ,All_Msg_Count %v",
					s.online.Load(), s.retry.Load(), s.singleMessageCount.Load(), s.allUserMessageCount.Load())
				print2.PrintWithColor(str, print2.FgGreen)
			case token := <-s.closeMonitor:
				fmt.Printf("Client_Offline:  %v is closed \r\n", token)
				s.retry.Inc()
				s.online.Dec()
				// go s.CreateClient(s.config.Identification)
			}
		}
	}()
}

func (s *Bench) getValidateKey() []string {
	var tokens []string
	for i := 0; i < s.config.Number; i++ {
		tokens = append(tokens, RandString(20))
	}
	return tokens
}

func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}
