// Package server 提供HTTP服务器实现
package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
)

// GracefulConfig 优雅关闭配置
type GracefulConfig struct {
	ShutdownTimeout time.Duration // 关闭超时时间
	CleanupFuncs    []func()      // 清理函数列表
}

// DefaultGracefulConfig 默认优雅关闭配置
var DefaultGracefulConfig = GracefulConfig{
	ShutdownTimeout: 30 * time.Second,
	CleanupFuncs:    nil,
}

// GracefulShutdown 优雅关闭管理器
type GracefulShutdown struct {
	config       GracefulConfig
	server       *http.Server
	shutdownChan chan struct{}
	doneChan     chan struct{}
}

// NewGracefulShutdown 创建优雅关闭管理器
func NewGracefulShutdown(server *http.Server, config ...GracefulConfig) *GracefulShutdown {
	cfg := DefaultGracefulConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return &GracefulShutdown{
		config:       cfg,
		server:       server,
		shutdownChan: make(chan struct{}),
		doneChan:     make(chan struct{}),
	}
}

// ListenForShutdown 监听关闭信号
func (g *GracefulShutdown) ListenForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		select {
		case sig := <-sigChan:
			logger.Log.Infof("📡 收到信号: %v，开始优雅关闭...", sig)
			g.Shutdown()
		case <-g.shutdownChan:
			// 手动触发关闭
		}
	}()
}

// Shutdown 执行优雅关闭
func (g *GracefulShutdown) Shutdown() {
	defer close(g.doneChan)

	logger.Log.Info("🔄 开始优雅关闭...")

	// 1. 关闭HTTP服务器
	if g.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), g.config.ShutdownTimeout)
		defer cancel()

		logger.Log.Info("⏳ 正在关闭HTTP服务器...")
		if err := g.server.Shutdown(ctx); err != nil {
			logger.Log.Errorf("❌ 关闭HTTP服务器失败: %v", err)
		} else {
			logger.Log.Info("✅ HTTP服务器已关闭")
		}
	}

	// 2. 执行清理函数
	for i, cleanup := range g.config.CleanupFuncs {
		logger.Log.Infof("🧹 执行清理函数 %d/%d", i+1, len(g.config.CleanupFuncs))
		cleanup()
	}

	// 3. 关闭数据库连接
	logger.Log.Info("⏳ 正在关闭数据库连接...")
	if err := database.Close(); err != nil {
		logger.Log.Errorf("❌ 关闭数据库失败: %v", err)
	} else {
		logger.Log.Info("✅ 数据库连接已关闭")
	}

	logger.Log.Info("👋 优雅关闭完成")
}

// Wait 等待关闭完成
func (g *GracefulShutdown) Wait() {
	<-g.doneChan
}

// TriggerShutdown 手动触发关闭
func (g *GracefulShutdown) TriggerShutdown() {
	close(g.shutdownChan)
}

// AddCleanupFunc 添加清理函数
func (g *GracefulShutdown) AddCleanupFunc(fn func()) {
	g.config.CleanupFuncs = append(g.config.CleanupFuncs, fn)
}

// ==================== 便捷函数 ====================

// WaitForShutdownSignal 等待关闭信号（简化版）
func WaitForShutdownSignal(server *http.Server, cleanup ...func()) {
	gs := NewGracefulShutdown(server)
	for _, fn := range cleanup {
		gs.AddCleanupFunc(fn)
	}
	gs.ListenForShutdown()

	// 阻塞等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	gs.Shutdown()
}

// RunWithGracefulShutdown 运行服务器并支持优雅关闭
func RunWithGracefulShutdown(server *http.Server, addr string, cleanup ...func()) error {
	// 启动服务器
	go func() {
		logger.Log.Infof("🚀 服务器启动在 %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatalf("❌ 服务器启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	WaitForShutdownSignal(server, cleanup...)
	return nil
}

// SetupGracefulShutdown 设置优雅关闭（返回等待函数）
func SetupGracefulShutdown(server *http.Server, cleanup ...func()) func() {
	gs := NewGracefulShutdown(server)
	for _, fn := range cleanup {
		gs.AddCleanupFunc(fn)
	}
	gs.ListenForShutdown()

	return func() {
		// 阻塞等待信号
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		gs.Shutdown()
	}
}
