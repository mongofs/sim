package sim

import (
	"net/http"
	"sim/pkg/conn"
	"testing"
)

type hook struct{}

func (h hook) Offline(conn conn.Connect, ty int) {
	panic("implement me")
}

func (h hook) Validate(token string) error {
	panic("implement me")
}

func (h hook) ValidateFailed(err error, cli conn.Connect) {
	panic("implement me")
}

func (h hook) ValidateSuccess(cli conn.Connect) {
	panic("implement me")
}

func (h hook) HandleReceive(conn conn.Connect, data []byte) {
	panic("implement me")
}

func (h hook) IdentificationHook(w http.ResponseWriter, r *http.Request) (string, error) {
	panic("implement me")
}

func TestNewSIMServer(t *testing.T) {
	type args struct {
		hooker Hooker
		opts   []OptionFunc
	}
	var (
		// test the right way
		testHookNotNil = hook{}
	)

	tests := []struct {
		name    string
		args    args
		wantErr error
		want    bool
	}{
		// test the right way
		{
			name: "test good situation ",
			args: args{
				hooker: testHookNotNil,
				opts:   []OptionFunc{},
			},
			wantErr: nil,
			want:    true,
		},
		// you can't run this test function along
		// create the object repeated
		{
			name: "create the object repeated ",
			args: args{
				hooker: testHookNotNil,
				opts:   []OptionFunc{},
			},
			wantErr: errInstanceIsExist,
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewSIMServer(tt.args.hooker, tt.args.opts...)
			if err != tt.wantErr {
				t.Errorf("NewSIMServer() error = '%v', wantErr '%v'", err, tt.wantErr)
			}
		})
	}
}

// we can't mock the IO of net , that is too complicate ,if we want to test this case
// we just need mock the interface of Connection witch in sim/pkg/conn/connection.go
//
func TestSendMessage(t *testing.T) {
	NewSIMServer(&hook{})
	go func() {
		if err := Run(); err != nil {
			panic(err)
		}
	}()
	type args struct {
		msg   []byte
		Users []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		want    bool
	}{
		// instance is not nil , test the single person
		{
			name: " send message when instance is not nil , test the single person",
			args: args{
				msg:   []byte(" the instance is nil "),
				Users: []string{"steven", "mike", "mikal"},
			},
			wantErr: nil,
		},

		// instance is not nil ,test the all person
		{
			name: " send message when instance is not nil ,test the all person",
			args: args{
				msg:   []byte(" the instance is nil "),
				Users: []string{"steven", "mike", "mikal"},
			},
			wantErr: nil,
		},

		// instance is not nil ,test the specific person
		{
			name: " send message when instance is not nil ,test the specific person ",
			args: args{
				msg:   []byte(" the instance is nil "),
				Users: []string{},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendMessage(tt.args.msg, tt.args.Users)
			if err != nil && err != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})

	}
}
