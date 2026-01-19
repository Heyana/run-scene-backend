//go:build dev
// +build dev

package main

import (
	"embed"
	"flag"
	"go_wails_project_manager/config"
	"go_wails_project_manager/core"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/server"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

// 使用 ** 通配符确保递归嵌入所有子目录和文件
//
//go:embed static/**
var StaticFiles embed.FS

//go:embed website/**
var WebsiteFiles embed.FS

// GetStaticFS 返回嵌入的静态文件系统
func GetStaticFS() fs.FS {
	return StaticFiles
}

// GetWebsiteFS 返回嵌入的网站文件系统
func GetWebsiteFS() fs.FS {
	// 打印当前工作目录
	cwd, err := os.Getwd()
	if err == nil {
		log.Println("当前工作目录:", cwd)
	}

	// 打印可执行文件路径
	execPath, err := os.Executable()
	if err == nil {
		log.Println("可执行文件路径:", execPath)
		log.Println("可执行文件目录:", filepath.Dir(execPath))
	}

	return WebsiteFiles
}

func main() {
	// 解析命令行参数
	port := flag.Int("port", 0, "服务器端口（覆盖配置文件）")
	flag.Parse()

	// 设置文件系统包装器
	server.GetStaticFSWrapper = GetStaticFS
	server.GetWebsiteFSWrapper = GetWebsiteFS

	// 确保数据目录存在
	dataDir := "./data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			logger.Log.Fatalf("创建数据目录失败: %v", err)
		}
	}

	// 设置环境变量
	os.Setenv("DB_PATH", filepath.Join(dataDir, "app.db"))
	os.Setenv("PRODUCT_DB_PATH", filepath.Join(dataDir, "product.db"))

	// 创建应用核心实例
	appCore, err := core.NewAppCore()
	if err != nil {
		logger.Log.Fatalf("初始化应用核心失败: %v", err)
	}

	// 如果指定了端口参数，覆盖配置
	if *port > 0 {
		config.AppConfig.ServerPort = *port
		logger.Log.Infof("使用命令行指定端口: %d", *port)
	}

	// 初始化数据库
	if err := appCore.InitDatabases(); err != nil {
		logger.Log.Fatalf("初始化数据库失败: %v", err)
	}

	// 启动服务器
	if err := appCore.StartServer(); err != nil {
		logger.Log.Fatalf("启动服务器失败: %v", err)
	}

	// 打印服务器状态
	status := appCore.GetServerStatus()
	logger.Log.Infof("服务器状态: 运行=%v, 端口=%d", status["running"], status["port"])

	// 等待关闭信号（优雅停机）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Log.Infof("📡 收到信号: %v，开始优雅关闭...", sig)

	// 执行优雅停机
	appCore.Shutdown()
}
