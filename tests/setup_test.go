package tests

import (
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/database"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/middleware"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	TestDB     *gorm.DB
	TestRouter *gin.Engine
	TestJWT    *middleware.JWTAuth
)

// SetupTestEnvironment 初始化测试环境
func SetupTestEnvironment() {
	// 设置测试模式
	gin.SetMode(gin.TestMode)
	
	// 设置测试数据库路径（在加载配置之前）
	testDBPath := "./data/test.db"
	os.Setenv("DB_PATH", testDBPath)
	
	// 初始化日志（logger 包的 init 函数会自动初始化 Log）
	logger.Init()
	
	// 加载配置
	config.LoadConfig()
	config.LoadRequirementConfig()
	
	// 确保需求管理功能启用并设置默认值
	if config.RequirementCfg == nil {
		config.RequirementCfg = &config.RequirementConfig{}
	}
	config.RequirementCfg.Requirement.Enabled = true
	
	// 确保任务配置有默认值
	if config.RequirementCfg.Requirement.Mission.DefaultStatus == "" {
		config.RequirementCfg.Requirement.Mission.DefaultStatus = "todo"
	}
	if config.RequirementCfg.Requirement.Mission.DefaultPriority == "" {
		config.RequirementCfg.Requirement.Mission.DefaultPriority = "P2"
	}
	
	// 初始化数据库
	if err := database.Init(); err != nil {
		panic("初始化测试数据库失败: " + err.Error())
	}
	TestDB = database.MustGetDB()
	
	// 初始化JWT
	TestJWT = middleware.NewJWTAuth()
}

// CleanupTestEnvironment 清理测试环境
func CleanupTestEnvironment() {
	if TestDB != nil {
		sqlDB, _ := TestDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
	
	// 删除测试数据库
	testDBPath := "./data/test.db"
	os.Remove(testDBPath)
}

// TestMain 测试入口
func TestMain(m *testing.M) {
	SetupTestEnvironment()
	
	// 初始化测试报告器
	GlobalReporter = NewTestReporter()
	
	code := m.Run()
	
	// 保存测试结果
	if err := GlobalReporter.SaveResults(); err != nil {
		fmt.Printf("保存测试结果失败: %v\n", err)
	}
	
	CleanupTestEnvironment()
	os.Exit(code)
}
