package logger

//@author Wang Weiwei
//@since 2020/3/24
import (
	"io"
	"gopkg.in/natefinch/lumberjack.v2"
)

/**
 * createWrite 创建一个可自动备份文件的输出流
 * path 文件路径，使用绝对路径
 * return 返回一个输出流
*/
func createWrite(path string) io.Writer {
	return  &lumberjack.Logger{
		Filename:   path, // 日志文件路径
		MaxSize:    128,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,     // 日志文件最多保存多少个备份
		MaxAge:     3,      // 文件最多保存多少天
		Compress:   true,   // 是否压缩
	}
	
}
