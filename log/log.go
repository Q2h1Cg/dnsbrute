package log

import "log"

const (
	// INFO log level: INFO
	INFO = 0
	// DEBUG log level: DEBUG
	DEBUG = iota
)

var logLevel = INFO

var (
	// Info Wrapper for log.Println
	Info = log.Println
	// Infof Wrapper for log.Printf
	Infof = log.Printf
	// Fatal Wrapper for log.Fatalln
	Fatal = log.Fatalln
	// Fatalf Wrapper for log.Fatalf
	Fatalf = log.Fatalf
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
		log.Println(msg...)
	}
}

// Debugf 显示调试信息，fotmated
func Debugf(format string, v ...interface{}) {
	if logLevel >= DEBUG {
		log.Printf(format, v...)
	}
}
