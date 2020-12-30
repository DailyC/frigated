package frigate

import (
	"errors"
	"os/exec"

	"github.com/frigated/pkgs/cgroup"
	"github.com/frigated/pkgs/logger"
)

//@author Wang Weiwei
//@since 2020/3/24

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

type Frigate struct {
	// config of task logger
	Log *logger.FLogger
	// control groups of resources，it only use on linux
	ControlGroups []*cgroup.ControlGroup

	ProtectTask *ProtectTask
	Strategy    *Strategy
}

/**
 * 应用配置接口
 * 主要用于讲配置数据应用于cmd
 */
type ApplyConfig interface {
	/**
	 * 应用配置
	 * return: 0 成功
	 * 			其它失败
	 */
	Apply(cmd *exec.Cmd) error
}

/**
 * 应用子进程配置
 */
func (frigate *Frigate) Apply(cmd *exec.Cmd) (err error) {
	err = frigate.Log.Apply(cmd)
	if err != nil {
		return err
	}
	err = frigate.Strategy.Apply(frigate.ProtectTask.Cmd)
	if err != nil {
		return err
	}
	return nil
}

// 启动守护进程
// 启动守护进程时会使用守护策略参数
func (frigate *Frigate) Start() error {
	if frigate.ProtectTask != nil && frigate.ProtectTask.Cmd != nil {
		return frigate.Apply(frigate.ProtectTask.Cmd)
	} else {
		// no args
		return errors.New("no init command")
	}
}
