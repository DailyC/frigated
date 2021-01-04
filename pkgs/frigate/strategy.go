package frigate

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"strconv"
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
	// 程序退出后自动重启,可选值：[true,false]，默认为 true
	AutoRestart bool
	//启动n秒后没有异常退出，就表示进程正常启动了，默认为10秒
	// 如果这个值为0，则没有启动失败的概念，即只要发现进程未在运行，就触发重启
	Startsecs time.Duration
	// 启动失败自动重试次数，默认是3
	StartRetries int
	// todo 启动进程前，是否尝试关闭同名其它进程
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

func defaultStrategy() *Strategy {
	return &Strategy{
		true,
		true,
		10 * time.Second,
		3,
		false,
		30 * time.Second,
		false,
		currentUser,
		"",
	}
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

	// 设置新进程用户
	if s.User != nil && s.User != currentUser {
		// 重新查找正确的用户
		if s.User.Uid != "" {
			s.User, err = user.LookupId(s.User.Uid)
			if err != nil {
				return err
			}
		} else if s.User.Name != "" {
			s.User, err = user.Lookup(s.User.Name)
			if err != nil {
				return err
			}
		} else {
			// 没有设置用户，默认继承当前用户
			s.User = currentUser
		}
		// todo 如果没有找到用户，且守护进程有root权限，则应创建相关用户
		uid, err := strconv.Atoi(s.User.Uid)
		if err != nil {
			return errors.New("find uid error " + err.Error())
		}
		gid, err := strconv.Atoi(s.User.Gid)
		if err != nil {
			return errors.New("find gid error " + err.Error())
		}
		credential := &syscall.Credential{}
		credential.Uid = uint32(uid)
		credential.Gid = uint32(gid)
		cmd.SysProcAttr.Credential = credential
	}
	return nil
}

/**
 * tryRestart 测试策略性的重启
 * 如程序异常结束，则应允许应用重启
 * 而当应用连续尝试启动失败后，不应该一直重启，因为这会造成一直浪费资源，且无法正常启动
 */
func (s *Strategy) tryRestart(t time.Duration) bool {
	if s.Startsecs != 0 && s.Startsecs > t {
		if s.StartRetries > 0 {
			s.StartRetries--
			return true
		} else {
			return false
		}
	}
	return s.AutoRestart
}
