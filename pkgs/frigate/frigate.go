package frigate

import (
	"github.com/docker/docker/pkg/reexec"
	"github.com/frigated/pkgs/cgroup"
	"github.com/frigated/pkgs/logger"
	"os"
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

// 创建守护的任务进程
//
// 1. 如果要以golang的函数作为子进程:
// name 代表 frigate 守护进程的标识，frigate 将使用这个名字创建子进程
// 注意如果要使用golang函数作为守护进程，那么函数需要提前注册
// @see RegisterGolangTask
//
// 2. 如果要以可执行程序作为子进程:
// name 代表可执行程序的绝对路径或者相对路径
func Create(name string) *Frigate {
	return &Frigate{
		Log:         logger.DefaultLogger(),
		ProtectTask: NewProtectTask(name),
		Strategy:    defaultStratage(),
	}
}

func Protect() {

}
