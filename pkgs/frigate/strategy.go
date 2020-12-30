package frigate

import (
	"os"
	"os/exec"
	"os/user"
	"syscall"
	"time"
)

var pid int = 0
var pgid int = 0
var currentUser *user.User

func init() {
	pid = os.Getpid()
	pgid, _ = syscall.Getpgid(syscall.Getpid())
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	currentUser = u

}

//@author Wang Weiwei
//@since 2020/3/24
// 子任务守护策略
type Strategy struct {
	//在 Frigated 启动的时候也自动启动
	AutoStart bool
	// 程序退出后自动重启,可选值：[unexpected,true,false]，默认为 unexpected
	AutoRestart bool
	//启动10秒后没有异常退出，就表示进程正常启动了，默认为1秒
	Startsecs time.Duration
	// 启动失败自动重试次数，默认是3
	StartRetries int
	// 启动进程前，是否尝试关闭同名其它进程
	GraceClose bool
	// 关闭同名进程最大等待时间
	GraceCloseWait time.Duration
	// frigated 被关闭时，是否强杀掉该进程
	Kill bool
	// 启动用户
	User *user.User
	// chroot
	Chroot string
}

type AutoRestartStrategy string

const (
	// 未知异常时重启
	UNEXPECTED = "unexpected"
	// 总是自动重启
	TRUE = "true"
	// 不自动重启
	FALSE = "false"
)

func defaultStratage() *Strategy {

}

/**
 * 在创建子进程时使用守护策略
 * 在创建进程时，子进程默认属于守护进程相同的进程组
 * 子进程默认与父进程是相同用户
 */
func (s *Strategy) Apply(cmd *exec.Cmd) (err error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pgid = pgid
	if s.User != nil {
		// 重新查找正确的用户
		if s.User.Uid != ""{
			s.User, err = user.LookupId(s.User.Uid)
			if err != nil {
				return err
			}
		}else if s.User.Name != "" {
			s.User, err = user.Lookup(s.User.Name)
			if err != nil {
				return err
			}
		} 
		// todo 如果没有找到用户，且守护进程有root权限，则应创建相关用户
		
		cmd.SysProcAttr.Credential = &syscall.Credential{}
	
	}
}
