// Package database 提供数据库迁移和升级工具
package database

import (
	"go_wails_project_manager/logger"
	// "sync"
)

var (
// migrationOnce sync.Once
)

// RunOnceUpgrade 运行一次性升级任务
func RunOnceUpgrade() error {
	logger.Log.Info("开始执行一次性升级任务...")

	logger.Log.Info("所有升级任务执行完成")
	return nil
}
