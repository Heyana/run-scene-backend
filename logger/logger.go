// Package logger 提供应用程序日志功能
package logger

import (
	"io"
	"os"
	"go_wails_project_manager/config"

	"github.com/sirupsen/logrus"
)

// Log 全局日志实例
var Log *logrus.Logger

// 包初始化
func init() {
	// 初始化日志实例
	Log = logrus.New()
	
	// 设置默认输出格式
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	// 设置默认输出到控制台
	Log.SetOutput(os.Stdout)
	
	// 设置默认日志级别
	Log.SetLevel(logrus.InfoLevel)
}

// Init 初始化日志系统
func Init() {
	// 已经在init函数中初始化了基本配置
	// 这里添加文件输出

	// 创建日志文件
	logFile, err := os.OpenFile("app_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// 同时输出到控制台和文件
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		Log.SetOutput(multiWriter)
	} else {
		// 如果无法创建文件，只输出到控制台
		Log.SetOutput(os.Stdout)
		Log.Warnf("无法创建日志文件: %v", err)
	}

	// 设置日志级别(如果配置已加载)
	if config.AppConfig != nil {
		Log.SetLevel(config.AppConfig.LogLevel)
	}
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return Log
}
