package frigate

import (
	"context"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"
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
			if tt.args.f.ProtectTask.Process != nil {
				t.Error("子进程无法正常结束")
			}
		})
	}
}

// TestunsafePointModiffyPrivite 非安全的修改os.Cmd 内部变量
func TestUnsafePointModiffyPrivite(t *testing.T) {
	cmd := exec.Command("/usr/bin/top")
	// cmd 开始指针
	ps := unsafe.Pointer(cmd)
	type AlignS struct {
		ProcessState    *os.ProcessState
		ctx             context.Context // nil means none
		lookPathErr     error           // LookPath error, if any.
		finished        bool            // when Wait was called
		childFiles      []*os.File
		closeAfterStart []io.Closer
		closeAfterWait  []io.Closer
		goroutine       []func() error
		errch           chan error // one send per goroutine
		waitDone        chan struct{}
	}

	as := &AlignS{}

	t.Run("测试非安全修改结构体私有变量", func(t *testing.T) {
		t.Logf("cmd start address is %d", uintptr(ps))
		t.Logf("cmd size is %d", unsafe.Sizeof(*cmd))
		t.Logf("cmd align size is %d", unsafe.Alignof(*cmd))
		// t.Logf("cmd value attr offset addr is %d", unsafe.Offsetof((*cmd).ProcessState))

		// 测试计算偏移量的占用内存大小关系
		if unsafe.Offsetof(cmd.ProcessState) + unsafe.Sizeof(*as) != unsafe.Sizeof(*cmd) {
			t.Logf("cmd attr ProcessState addr is %d", unsafe.Offsetof(cmd.ProcessState))
			t.Logf("size of AlignS %d", unsafe.Sizeof(*as))
			t.Errorf("state addr offset + size of AlignS is not eq size of cmd")
		}

		// alignF := unsafe.Offsetof(as.finished)
		
		
		// t.Logf("size of AlignS %d", unsafe.Sizeof(as.SysProcAttr))
	})
}
