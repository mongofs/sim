package sim

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mongofs/sim/pkg/conn"
)

type MockConn struct {
	id string
	heartTime int64
}

func (m MockConn) Identification() string {
	return m.id
}

func (m MockConn) Send(data []byte) error {
	fmt.Printf("%v received message : %v\n",m.id ,string(data))
	return nil
}

func (m MockConn) Close() {
	fmt.Printf("%v Close the connection \n",m.id )
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

func TestNewBucket(t *testing.T) {
	type args struct {
		option *Options
		id     int
		buffer int
	}
	tests := []struct {
		name string
		args args
		want *bucket
	}{
		{
			name :"test for new bucket with the right parameter",
			args: args{
				option: DefaultOption(),
				id:     1001,
				buffer: 20 ,
			},
			want: nil,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBucket(tt.args.option, tt.args.id, tt.args.buffer); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

