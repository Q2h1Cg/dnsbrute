package log

import syslog "log"

const (
	INFO  = 0
	DEBUG = iota
)

var logLevel = INFO

var (
	Info   = syslog.Println
	Infof  = syslog.Printf
	Fatal  = syslog.Fatalln
	Fatalf = syslog.Fatalf
)

// SetLevel 设置日志级别
func SetLevel(level int) {
	if level < 2 {
		logLevel = level
	}
}

// Debug 显示调试信息
func Debug(msg ...interface{}) {
	if logLevel >= DEBUG {
		syslog.Println(msg...)
	}
}

// Debugf 显示调试信息，fotmated
func Debugf(format string, v ...interface{}) {
	if logLevel >= DEBUG {
		syslog.Printf(format, v...)
	}
}
