// Package server 提供HTTP服务器实现
package server

import (
	"context"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Server HTTP服务器结构
type Server struct {
	port   int
	router *gin.Engine
	server *http.Server
	log    *logrus.Logger
}

// GetStaticFSWrapper 从main包获取静态文件系统的包装函数
// 这个函数在main包中声明，因为Go不允许导入main包
var GetStaticFSWrapper func() fs.FS

var GetWebsiteFSWrapper func() fs.FS

func init() {
	// macOS app bundle detection and working directory adjustment
	execPath, err := os.Executable()
	if err == nil && strings.Contains(execPath, "Contents/MacOS") {
		// We're running from a macOS app bundle
		// Set working directory to the executable's directory
		appDir := filepath.Dir(execPath)
		os.Chdir(appDir)
	}
}

// NewServer 创建新的HTTP服务器实例
func NewServer(port int) *Server {
	// 使用日志
	log := logger.GetLogger()

	// 设置gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	router := gin.Default()
	return &Server{
		port:   port,
		router: router,
		log:    log,
	}
}

// AddRoutes 添加路由
func (s *Server) AddRoutes(routesFunc func(*gin.Engine)) {
	routesFunc(s.router)
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	// 配置静态文件服务
	if false && GetStaticFSWrapper != nil && GetWebsiteFSWrapper != nil && !config.IsDev() {
		// 生产环境使用嵌入式文件系统
		staticFS := GetStaticFSWrapper()
		websiteFS := GetWebsiteFSWrapper()

		if staticFS != nil {
			s.log.Info("使用嵌入式文件系统提供静态资源")
			// 添加更多日志以便调试
			s.log.Infof("嵌入式文件系统类型: %T", staticFS)

			// 先尝试列出嵌入文件系统中的文件
			entries, err := fs.ReadDir(staticFS, ".")
			if err != nil {
				s.log.Warnf("无法读取嵌入式文件系统根目录: %v", err)
			} else {
				s.log.Info("嵌入式文件系统中的文件/目录:")
				for _, entry := range entries {
					s.log.Infof("- %s (是目录: %v)", entry.Name(), entry.IsDir())
				}
			}

			// 直接使用原始http.Handler处理静态文件请求
			// 这比StaticFS更直接，避免Gin的路径处理问题
			s.router.Any("/website/*filepath", func(c *gin.Context) {
				// 使用http.FileServer处理请求
				staticServer := http.FileServer(http.FS(websiteFS))
				// 重写请求路径，去掉前缀
				c.Request.URL.Path = c.Param("filepath")
				staticServer.ServeHTTP(c.Writer, c.Request)
			})

			s.router.Any("/static/*filepath", func(c *gin.Context) {
				filePath := "./static/" + c.Param("filepath")
				s.log.Debugf("请求动态静态文件: %s", filePath)
				http.ServeFile(c.Writer, c.Request, filePath)
			})

			// 为方便测试，添加一个简单的静态文件处理器
			s.router.GET("/statictest", func(c *gin.Context) {
				c.String(200, "静态文件系统挂载正常")
			})
		} else {
			// 如果嵌入式文件系统获取失败，回退到使用外部文件系统
			s.router.Static("/static", "./static")
			// 添加uploads目录的访问
			s.router.Static("/upload", "./static/uploads")
		}
	} else {
		// 开发环境使用外部文件系统
		s.log.Info("开发环境：使用外部文件系统提供静态资源")

		// 配置静态文件目录
		s.router.Static("/static", "./static")
		
		// 配置模型库静态文件
		if config.AppConfig.Model.LocalStorageEnabled {
			s.router.Static("/models", config.AppConfig.Model.StorageDir)
			s.log.Infof("模型库静态文件服务: %s", config.AppConfig.Model.StorageDir)
		}
		
		// 配置资产库静态文件
		if config.AppConfig.Asset.LocalStorageEnabled {
			s.router.Static("/assets", config.AppConfig.Asset.StorageDir)
			s.log.Infof("资产库静态文件服务: %s", config.AppConfig.Asset.StorageDir)
		}

		// 配置CDN目录（用于开发环境模拟CDN）
		cdnPath := config.GetResourcePath()
		if _, err := os.Stat(cdnPath); os.IsNotExist(err) {
			// 确保CDN目录存在
			if err := os.MkdirAll(cdnPath, 0755); err != nil {
				s.log.Warnf("创建CDN目录失败: %v", err)
			}
		}

		s.router.Static("/cdn", cdnPath)
		s.log.Infof("CDN服务已配置在: %s，可通过 %s 访问", cdnPath, config.GetCDNBaseURL())

		// 添加CORS头，允许本地前端开发
		s.router.Use(func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
			c.Next()
		})
	}

	// 在goroutine中启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("HTTP服务器启动失败: %v", err))
		}
	}()

	// 等待一小段时间确保服务器启动
	time.Sleep(100 * time.Millisecond)

	return nil
}

// Stop 停止HTTP服务器
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.AppConfig.ServerPort)*time.Second)
	defer cancel()

	s.log.Info("正在关闭 HTTP 服务器...")
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.log.Errorf("关闭 HTTP 服务器出错: %v", err)
		return err
	}

	s.log.Info("HTTP 服务器已成功关闭")
	return nil
}
