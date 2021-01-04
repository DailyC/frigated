package frigate

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/frigated/pkgs/cgroup"
	"github.com/frigated/pkgs/logger"
)

//@author Wang Weiwei
//@since 2020/3/24

// Create 创建守护的任务进程
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
		Strategy:    defaultStrategy(),
	}
}

type Frigate struct {
	// config of task logger
	Log *logger.FLogger
	// control groups of resources，it only use on linux
	ControlGroups []*cgroup.ControlGroup

	ProtectTask *ProtectTask
	Strategy    *Strategy
	// 进程信号通道
	SignalChan chan error
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

// Runable 可运行程序接口
/*  可运行的任务接口
 */
type Runable interface {
	// 启动
	Start() (err error)
	// 停止
	Stop(d time.Duration) (err error)
}

// Apply 应用配置
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

// Start 启动守护进程
// 启动守护进程时会使用守护策略参数
// 启动进程后，会监听进程状态
func (frigate *Frigate) Start() (err error) {
	if frigate.ProtectTask != nil && frigate.ProtectTask.Cmd != nil {
		err = frigate.Apply(frigate.ProtectTask.Cmd)
		if err != nil {
			return err
		}

		frigate.Log.Stderr.Write([]byte(fmt.Sprintf("[DEBUG] %s frigated.go start %s task by frigate\n", time.Now().String(), frigate.ProtectTask.Name)))
		err = frigate.ProtectTask.Start()
		if err != nil {
			frigate.Log.Stderr.Write([]byte(fmt.Sprintf("[ERROR] %s frigated.go %s task can not start case by %s\n",
				time.Now().String(), frigate.ProtectTask.Name, err.Error())))
			return err
		}

		// 进程退出信号监听
		go func() {
			for e := range frigate.ProtectTask.Done() {
				// 用户主动关闭进程
				if e.Error() == CANCEL_PROCESS {
					frigate.Log.Stderr.Write([]byte(fmt.Sprintf("[WARN] %s frigated.go cancel %s task by frigate\n", time.Now().String(), frigate.ProtectTask.Name)))
				} else {
					// case 2： 尝试异常重启
					if frigate.Strategy.tryRestart(frigate.ProtectTask.StartTime.Sub(time.Now())) {
						frigate.Log.Stderr.Write([]byte(
							fmt.Sprintf("[ERROR] %s frigated.go %s task exit %s, try restart task and remaining %d times for retry start fail case\n",
								time.Now().String(), frigate.ProtectTask.Name, e.Error(), frigate.Strategy.StartRetries)))
						frigate.Start()
					} else {
						// case 3 无法正常启动
						frigate.Log.Stderr.Write([]byte(fmt.Sprintf("[ERROR] %s frigated.go  %s task start fail %s, and beyond the max restart times\n", 
							time.Now().String(), frigate.ProtectTask.Name, e.Error())))
					}
				}
			}
		}()

	} else {
		// no args
		return errors.New("no init command")
	}
	return nil
}

// Stop 用户主动调用，关闭守护进程逻辑
func (frigate *Frigate) Stop(d time.Duration) (err error) {

	err = frigate.ProtectTask.Stop(frigate.Strategy.GraceCloseWait)
	frigate.Log.Stderr.Write([]byte(fmt.Sprintf("[WARNING] %s frigated.go  %s task stop return %s\n", time.Now().String(), frigate.ProtectTask.Name, err.Error())))
	return err
}
