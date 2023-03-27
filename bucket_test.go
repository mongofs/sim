package sim

import (
	"fmt"
	"testing"
	"time"

	"sim/pkg/conn"
)

type MockConn struct {
	id        string
	heartTime int64
}

func (m MockConn) Identification() string {
	return m.id
}

func (m MockConn) Send(data []byte) error {
	fmt.Printf("%v received message : %v\n", m.id, string(data))
	return nil
}

func (m MockConn) Close() {
	fmt.Printf("%v Close the connection \n", m.id)
	return
}

func (m MockConn) SetMessageType(messageType conn.MessageType) {
	return
}

func (m MockConn) ReFlushHeartBeatTime() {
	m.heartTime = time.Now().Unix()
}

func (m MockConn) GetLastHeartBeatTime() int64 {
	return m.heartTime
}

// send message to a person
func TestBucket_SendMessage(t *testing.T) {

}
