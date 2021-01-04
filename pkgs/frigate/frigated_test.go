package frigate

import (
	"testing"
	"time"
)

func TestProtect(t *testing.T) {
	type args struct {
		f *Frigate
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"test child process use top",
			args{Create("/usr/bin/top")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Protect(tt.args.f)
			if tt.args.f.ProtectTask.isCloseSC {
				t.Error("子进程异常结束")
			}
			time.Sleep(time.Second * 40)
			tt.args.f.Stop(0)
			time.Sleep(time.Second * 15)
			if tt.args.f.ProtectTask.Process != nil {
				t.Error("子进程无法正常结束")
			}
		})
	}
}
