package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger.Out = file
	} else {
		fmt.Println("Failed to log to file, using default stderr")
	}
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

// SetLogLevel 设置日志级别
func SetLogLevel(level logrus.Level) {
	logger.SetLevel(level)
}

// Warn 级别日志
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error 级别日志
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Debug 级别日志
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func SetLogFormat(format string) {
	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else if format == "text" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true, // 输出完整时间戳
		})
	}
}
