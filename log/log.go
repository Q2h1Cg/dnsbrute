package log

import _log "log"

const (
	INFO  = 0
	DEBUG = iota
)

var logLevel = INFO

var (
	Info   = _log.Println
	Infof  = _log.Printf
	Fatal  = _log.Fatalln
	Fatalf = _log.Fatalf
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
		_log.Println(msg...)
	}
}

// Debugf 显示调试信息，fotmated
func Debugf(format string, v ...interface{}) {
	if logLevel >= DEBUG {
		_log.Printf(format, v...)
	}
}
