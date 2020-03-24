package logger

//@author Wang Weiwei
//@since 2020/3/24
import (
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func createWrite() {
	hook := lumberjack.Logger{
		Filename:   "path", // 日志文件路径
		MaxSize:    128,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,     // 日志文件最多保存多少个备份
		MaxAge:     3,      // 文件最多保存多少天
		Compress:   true,   // 是否压缩
	}
}
