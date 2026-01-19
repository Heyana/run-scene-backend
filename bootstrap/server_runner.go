// Package bootstrap 提供应用程序启动相关的功能
package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"go_wails_project_manager/core"
	"go_wails_project_manager/logger"
)

// 捕获panic的工具函数
func recoverPanic() {
	if r := recover(); r != nil {
		logger.Log.Errorf("服务器启动过程中发生panic: %v", r)
		// 记录到专门的错误日志文件
		errorFile, err := os.OpenFile("server_error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			defer errorFile.Close()
			fmt.Fprintf(errorFile, "[%s] 服务器启动过程中发生panic: %v\n", 
				time.Now().Format("2006-01-02 15:04:05"), r)
		}
	}
}

// SafeStartHTTPServer 安全地启动HTTP服务器
func SafeStartHTTPServer(appCore *core.AppCore) error {
	// 用defer和recover捕获任何可能的panic
	defer recoverPanic()
	
	// 尝试启动HTTP服务器
	return appCore.StartServer()
}

// RunServer 启动应用服务器
func RunServer() {
	// 确保最外层有panic恢复，防止整个应用崩溃
	defer recoverPanic()
	
	// 确保数据目录存在
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			logger.Log.Errorf("创建数据目录失败: %v", err)
			return
		}
	}

	// 设置环境变量
	os.Setenv("DB_PATH", filepath.Join(dataDir, "app.db"))
	os.Setenv("PRODUCT_DB_PATH", filepath.Join(dataDir, "product.db"))

	// 创建应用核心实例
	appCore, err := core.NewAppCore()
	if err != nil {
		logger.Log.Errorf("初始化应用核心失败: %v", err)
		return
	}

	// 初始化数据库
	if err := appCore.InitDatabases(); err != nil {
		logger.Log.Errorf("初始化数据库失败: %v", err)
		return
	}

	// 安全地启动HTTP服务器
	if err := SafeStartHTTPServer(appCore); err != nil {
		logger.Log.Errorf("启动HTTP服务器失败: %v", err)
		// 记录到专门的错误日志文件
		errorFile, err := os.OpenFile("server_error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			defer errorFile.Close()
			fmt.Fprintf(errorFile, "[%s] 启动HTTP服务器失败: %v\n", 
				time.Now().Format("2006-01-02 15:04:05"), err)
		}
		return
	}

	// 打印服务器状态
	status := appCore.GetServerStatus()
	logger.Log.Infof("服务器状态: 运行=%v, 端口=%d", status["running"], status["port"])

	// 阻塞主线程，保持服务器运行
	select {}
} 