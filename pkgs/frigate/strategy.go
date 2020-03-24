package frigate

import (
	"os/user"
	"time"
)

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
	// frigated 被时，是否强杀掉该进程
	Kill bool
	// 启动用户
	User *user.User
}

// 程序自动重启策略
type AutoRestartStrategy string

const (
	// 未知异常时重启
	UNEXPECTED = "unexpected"
	// 总是自动重启
	TRUE = "true"
	// 不自动重启
	FALSE = "false"
)
