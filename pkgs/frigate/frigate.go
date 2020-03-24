package frigate

import (
	"github.com/frigated/pkgs/cgroup"
	"github.com/frigated/pkgs/logger"
	"os"
	"os/exec"
	"time"
)

//@author Wang Weiwei
//@since 2020/3/24
type Frigate struct {
	// config of task logger
	Log *logger.FLogger
	// control groups of resources，it only use on linux
	ControlGroups []*cgroup.ControlGroup

	ProtectTask *ProtectTask
	Strategy    *Strategy
}

// 受保护任务定义
type ProtectTask struct {
	// the task run in child process
	Cmd       *exec.Cmd
	Name      string
	Process   *os.Process
	StartTime time.Time
}

// 注册要守护的任务进程
// name 代表 frigate 守护进程的标识，frigate 将使用这个名字创建子进程
// task 是要作为子进程启动的函数。注意不要忘记对函数引用的变量做初始化。
func Register(name string, task func()) *Frigate {
	return &Frigate{}
}

// 注册要守护的可执行文件
// path 代表可执行文件的绝对路径
func RegisterExec(path string) *Frigate {
	return &Frigate{}
}

func Protect() {

}
