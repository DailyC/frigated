package logger

import (
	"io"
	"os"
	"os/exec"
)

//@author Wang Weiwei
//@since 2020/3/24

type ByteSize int64

const (
	B  = ByteSize(1)
	KB = ByteSize(1024)
	MB = KB * 1024
	GB = MB * 1024
)

// Frigate 的日志数据
type FLogger struct {
	// 标准输出流日志
	Stdout io.Writer
	Stderr io.Writer
	// 日志备份数量
	Backups int
	// 日志文件大小
	Maxbytes ByteSize
	// 把 stderr 重定向到 stdout，默认 false,错误日志也会写进Stdout中
	Redirect bool
}

// 基于标准输出流，构造一个日志控制器
func DefaultLogger() *FLogger {
	return &FLogger{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Backups:  0,
		Maxbytes: 0,
		Redirect: false,
	}
}

/**
 * 应用logger配置，包括重定向输出流及启用文件清理相关功能
 */
func (f *FLogger)Apply(cmd *exec.Cmd) error {
		cmd.Stdin = nil
		cmd.Stdout = f.Stdout
		cmd.Stderr = f.Stderr
		return nil
}