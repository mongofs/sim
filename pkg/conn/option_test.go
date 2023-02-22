package conn

import "testing"

func TestSetOption(t *testing.T) {
	type args struct {
		option *Option
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "no Option",
			args: args{
				option: &Option{
					Buffer:                0,
					MessageType:           1,
					ConnectionWriteBuffer: 0,
					ConnectionReadBuffer:  0,
				}},
			want: ErrBufferParam,
		},
		{
			name: "bad Option",
			args: args{option: &Option{
				Buffer:                1,
				MessageType:           1,
				ConnectionWriteBuffer: 0,
				ConnectionReadBuffer:  -20,
			}},
			want: ErrConnReadBufferParam,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetOption(tt.args.option)
			if err !=tt.want {
				t.Errorf("SetOption() error = '%v', wantErr '%v' \n", err, tt.want)
			}
		})
	}
}

