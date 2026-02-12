// Package database 提供数据库迁移和升级工具
package database

import (
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"os"
	"strings"

	"gorm.io/gorm"
)

// RunOnceUpgrade 运行一次性升级任务
func RunOnceUpgrade(db *gorm.DB) error {
	logger.Log.Info("开始执行一次性升级任务...")

	// 检查并执行版本化升级
	if err := runVersionedUpgrades(db); err != nil {
		logger.Log.Errorf("执行版本化升级失败: %v", err)
	}

	logger.Log.Info("所有升级任务执行完成")
	return nil
}

// runVersionedUpgrades 执行版本化的升级任务
func runVersionedUpgrades(db *gorm.DB) error {
	// 读取配置文件中的目标版本
	targetVersion := getTargetVersionFromConfig()
	logger.Log.Infof("配置文件目标版本: %d", targetVersion)
	
	// 获取上次执行的版本（从文件读取）
	lastVersion := getLastExecutedVersion()
	logger.Log.Infof("上次执行版本: %d", lastVersion)

	// 如果已经是最新版本，跳过
	if lastVersion >= targetVersion {
		logger.Log.Info("数据库已是最新版本，跳过升级")
		return nil
	}

	// 版本 1: 基础清理和修复
	if lastVersion < 1 && targetVersion >= 1 {
		logger.Log.Info("执行版本 1 升级...")
		
		// 1. 删除 cdn_path 为空的文件记录及其关联
		if err := cleanEmptyCDNPathFiles(db); err != nil {
			logger.Log.Errorf("清理空 CDN 路径失败: %v", err)
		}

		// 2. 更新 cdn_path 为相对路径（移除 base_url 前缀）
		if err := updateCDNPathToRelative(db); err != nil {
			logger.Log.Errorf("更新 CDN 路径失败: %v", err)
		}

		// 3. 删除重复的文件记录（保留最新的）
		if err := cleanDuplicateFiles(db); err != nil {
			logger.Log.Errorf("清理重复文件失败: %v", err)
		}

		// 4. 更新材质同步状态（根据文件下载情况）
		if err := updateTextureSyncStatus(db); err != nil {
			logger.Log.Errorf("更新材质同步状态失败: %v", err)
		}

		// 5. 修复 CDN 路径中的 textures/textures/ 重复问题
		if err := fixDuplicateTexturesInCDNPath(db); err != nil {
			logger.Log.Errorf("修复 CDN 路径失败: %v", err)
		}

		saveLastExecutedVersion(1)
		logger.Log.Info("版本 1 升级完成")
	}

	// 版本 2: 贴图类型分析和映射（2026-01-26）
	if lastVersion < 2 && targetVersion >= 2 {
		logger.Log.Info("执行版本 2 升级: 重新分析所有贴图类型...")
		
		// 强制重新分析所有文件的贴图类型
		if err := analyzeTextureTypes(db); err != nil {
			logger.Log.Errorf("分析贴图类型失败: %v", err)
		}

		// 强制重新更新所有材质的贴图类型列表
		if err := updateTextureTypesList(db); err != nil {
			logger.Log.Errorf("更新材质贴图类型列表失败: %v", err)
		}

		// 记录贴图类型分析结果到日志
		if err := logTextureTypeAnalysis(db); err != nil {
			logger.Log.Errorf("记录贴图类型分析失败: %v", err)
		}

		saveLastExecutedVersion(2)
		logger.Log.Info("版本 2 升级完成")
	}

	// 版本 3: 改进 AmbientCG 贴图类型提取（2026-01-26）
	if lastVersion < 3 && targetVersion >= 3 {
		logger.Log.Info("执行版本 3 升级: 改进 AmbientCG 贴图类型提取...")
		
		// 强制重新分析所有文件的贴图类型（使用新的 AmbientCG 提取逻辑）
		if err := analyzeTextureTypes(db); err != nil {
			logger.Log.Errorf("分析贴图类型失败: %v", err)
		}

		// 强制重新更新所有材质的贴图类型列表
		if err := updateTextureTypesList(db); err != nil {
			logger.Log.Errorf("更新材质贴图类型列表失败: %v", err)
		}

		// 记录贴图类型分析结果到日志
		if err := logTextureTypeAnalysis(db); err != nil {
			logger.Log.Errorf("记录贴图类型分析失败: %v", err)
		}

		saveLastExecutedVersion(3)
		logger.Log.Info("版本 3 升级完成")
	}

	// 版本 4: 清理丢失的文件记录（2026-01-26）
	if lastVersion < 4 && targetVersion >= 4 {
		logger.Log.Info("执行版本 4 升级: 清理丢失的文件记录...")
		
		// 检查并删除物理文件不存在的记录
		if err := cleanMissingFiles(db); err != nil {
			logger.Log.Errorf("清理丢失文件失败: %v", err)
		}

		// 更新材质同步状态
		if err := updateTextureSyncStatus(db); err != nil {
			logger.Log.Errorf("更新材质同步状态失败: %v", err)
		}

		saveLastExecutedVersion(4)
		logger.Log.Info("版本 4 升级完成")
	}

	// 版本 5: 修复贴图类型解析（2026-01-26）
	if lastVersion < 5 && targetVersion >= 5 {
		logger.Log.Info("执行版本 5 升级: 修复贴图类型解析...")
		
		// 重新分析所有文件的贴图类型（使用统一的解析函数）
		if err := analyzeTextureTypes(db); err != nil {
			logger.Log.Errorf("分析贴图类型失败: %v", err)
		}

		// 重新更新所有材质的贴图类型列表
		if err := updateTextureTypesList(db); err != nil {
			logger.Log.Errorf("更新材质贴图类型列表失败: %v", err)
		}

		saveLastExecutedVersion(5)
		logger.Log.Info("版本 5 升级完成")
	}

	// 版本 6: 添加文件夹递归统计字段（2026-02-12）
	if lastVersion < 6 && targetVersion >= 6 {
		logger.Log.Info("执行版本 6 升级: 添加文件夹递归统计字段...")
		
		// 添加新字段（GORM 会自动处理）
		if err := db.AutoMigrate(&models.Document{}); err != nil {
			logger.Log.Errorf("自动迁移失败: %v", err)
		} else {
			logger.Log.Info("成功添加 total_size, total_count, stats_updated_at 字段")
		}
		
		// 计算所有文件夹的统计信息
		if err := models.RecalculateAllFolderStats(db); err != nil {
			logger.Log.Errorf("计算文件夹统计失败: %v", err)
		} else {
			logger.Log.Info("成功计算所有文件夹的递归统计")
		}

		saveLastExecutedVersion(6)
		logger.Log.Info("版本 6 升级完成")
	}

	// 版本 7: 创建审计日志表（2026-02-12）
	if lastVersion < 7 && targetVersion >= 7 {
		logger.Log.Info("执行版本 7 升级: 创建审计日志表...")
		
		// 创建审计日志表
		if err := db.AutoMigrate(&models.AuditLog{}); err != nil {
			logger.Log.Errorf("创建审计日志表失败: %v", err)
		} else {
			logger.Log.Info("成功创建 audit_logs 表")
		}

		saveLastExecutedVersion(7)
		logger.Log.Info("版本 7 升级完成")
	}

	return nil
}

// getTargetVersionFromConfig 从配置文件读取目标版本
func getTargetVersionFromConfig() int {
	if config.DatabaseVersion == nil {
		logger.Log.Warn("数据库版本配置未加载，使用默认版本 0")
		return 0
	}
	return config.DatabaseVersion.GetTargetVersion()
}

// getLastExecutedVersion 获取上次执行的版本（从文件读取）
func getLastExecutedVersion() int {
	data, err := os.ReadFile("data/.db_version")
	if err != nil {
		// 文件不存在，返回 0
		return 0
	}

	versionStr := strings.TrimSpace(string(data))
	version := 0
	for _, c := range versionStr {
		if c >= '0' && c <= '9' {
			version = version*10 + int(c-'0')
		}
	}
	return version
}

// saveLastExecutedVersion 保存已执行的版本号到文件
func saveLastExecutedVersion(version int) error {
	// 确保 data 目录存在
	os.MkdirAll("data", 0755)

	// 将版本号转换为字符串
	versionStr := ""
	if version == 0 {
		versionStr = "0"
	} else {
		temp := version
		digits := []rune{}
		for temp > 0 {
			digits = append([]rune{rune('0' + temp%10)}, digits...)
			temp /= 10
		}
		versionStr = string(digits)
	}

	// 写入文件
	err := os.WriteFile("data/.db_version", []byte(versionStr), 0644)
	if err != nil {
		logger.Log.Errorf("保存版本号失败: %v", err)
		return err
	}

	logger.Log.Infof("已保存版本号: %d", version)
	return nil
}

// cleanEmptyCDNPathFiles 删除 cdn_path 为空的文件记录及其关联
func cleanEmptyCDNPathFiles(db *gorm.DB) error {
	logger.Log.Info("开始清理空 CDN 路径的文件...")

	// 查找所有 cdn_path 为空的文件
	var emptyFiles []models.File
	if err := db.Where("cdn_path = ? OR cdn_path IS NULL", "").Find(&emptyFiles).Error; err != nil {
		return err
	}

	if len(emptyFiles) == 0 {
		logger.Log.Info("没有找到空 CDN 路径的文件")
		return nil
	}

	logger.Log.Infof("找到 %d 个空 CDN 路径的文件", len(emptyFiles))

	// 删除关联的 texture_files 记录
	for _, file := range emptyFiles {
		// 删除 texture_files 关联
		if err := db.Where("file_id = ?", file.ID).Delete(&models.TextureFile{}).Error; err != nil {
			logger.Log.Warnf("删除文件关联失败 (file_id=%d): %v", file.ID, err)
		}

		// 如果是贴图的主文件，删除贴图记录
		if file.RelatedType == "Texture" && file.FileType == "thumbnail" {
			if err := db.Where("id = ?", file.RelatedID).Delete(&models.Texture{}).Error; err != nil {
				logger.Log.Warnf("删除贴图记录失败 (texture_id=%d): %v", file.RelatedID, err)
			}
		}
	}

	// 删除文件记录
	if err := db.Where("cdn_path = ? OR cdn_path IS NULL", "").Delete(&models.File{}).Error; err != nil {
		return err
	}

	logger.Log.Infof("成功清理 %d 个空 CDN 路径的文件", len(emptyFiles))
	return nil
}

// updateCDNPathToRelative 更新 cdn_path 为相对路径
func updateCDNPathToRelative(db *gorm.DB) error {
	logger.Log.Info("开始更新 CDN 路径为相对路径...")

	// 查找所有包含完整 URL 的文件
	var files []models.File
	if err := db.Where("cdn_path LIKE ?", "http%").Find(&files).Error; err != nil {
		return err
	}

	if len(files) == 0 {
		logger.Log.Info("没有找到需要更新的文件")
		return nil
	}

	logger.Log.Infof("找到 %d 个需要更新的文件", len(files))

	// 更新每个文件的 cdn_path
	for _, file := range files {
		// 提取相对路径（移除 base_url 前缀）
		// 例如: http://192.168.3.39:23359/textures/leather_white/nor_dx_2k.jpg -> leather_white/nor_dx_2k.jpg
		relativePath := file.CDNPath
		
		// 查找 /textures/ 后的路径
		if idx := findLastIndex(relativePath, "/textures/"); idx != -1 {
			relativePath = relativePath[idx+10:] // 跳过 "/textures/"
		} else if idx := findLastIndex(relativePath, "/"); idx != -1 {
			// 如果没有 /textures/，尝试提取最后两段路径
			parts := splitPath(relativePath)
			if len(parts) >= 2 {
				relativePath = parts[len(parts)-2] + "/" + parts[len(parts)-1]
			}
		}

		// 更新数据库
		if err := db.Model(&file).Update("cdn_path", relativePath).Error; err != nil {
			logger.Log.Warnf("更新文件 CDN 路径失败 (id=%d): %v", file.ID, err)
		}
	}

	logger.Log.Infof("成功更新 %d 个文件的 CDN 路径", len(files))
	return nil
}

// cleanDuplicateFiles 删除重复的文件记录（保留最新的）
func cleanDuplicateFiles(db *gorm.DB) error {
	logger.Log.Info("开始清理重复文件...")

	// 查找所有重复的文件（根据 related_id, related_type, file_name）
	type DuplicateFile struct {
		RelatedID   uint
		RelatedType string
		FileName    string
		Count       int
	}

	var duplicates []DuplicateFile
	if err := db.Model(&models.File{}).
		Select("related_id, related_type, file_name, COUNT(*) as count").
		Group("related_id, related_type, file_name").
		Having("COUNT(*) > 1").
		Find(&duplicates).Error; err != nil {
		return err
	}

	if len(duplicates) == 0 {
		logger.Log.Info("没有找到重复文件")
		return nil
	}

	logger.Log.Infof("找到 %d 组重复文件", len(duplicates))

	totalDeleted := 0
	for _, dup := range duplicates {
		// 查找该组的所有文件，按 ID 降序排列（保留最新的）
		var files []models.File
		if err := db.Where("related_id = ? AND related_type = ? AND file_name = ?",
			dup.RelatedID, dup.RelatedType, dup.FileName).
			Order("id DESC").
			Find(&files).Error; err != nil {
			logger.Log.Warnf("查询重复文件失败: %v", err)
			continue
		}

		// 保留第一个（最新的），删除其他的
		for i := 1; i < len(files); i++ {
			file := files[i]

			// 删除 texture_files 关联
			if err := db.Where("file_id = ?", file.ID).Delete(&models.TextureFile{}).Error; err != nil {
				logger.Log.Warnf("删除文件关联失败 (file_id=%d): %v", file.ID, err)
			}

			// 删除文件记录
			if err := db.Delete(&file).Error; err != nil {
				logger.Log.Warnf("删除文件记录失败 (id=%d): %v", file.ID, err)
			} else {
				totalDeleted++
			}
		}
	}

	logger.Log.Infof("成功清理 %d 个重复文件", totalDeleted)
	return nil
}

// findLastIndex 查找字符串最后一次出现的位置
func findLastIndex(s, substr string) int {
	idx := -1
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			idx = i
		}
	}
	return idx
}

// splitPath 分割路径
func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, ch := range path {
		if ch == '/' || ch == '\\' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}


// updateTextureSyncStatus 更新材质同步状态
// 根据文件下载情况自动设置 sync_status 和 download_completed
func updateTextureSyncStatus(db *gorm.DB) error {
	logger.Log.Info("开始更新材质同步状态...")

	// 按需下载模式：只更新状态不一致的材质
	// 查找状态可能不一致的材质（有文件但状态为0，或无文件但状态为2）
	var textures []models.Texture
	if err := db.Where("sync_status IN (0, 3)").Find(&textures).Error; err != nil {
		return err
	}

	if len(textures) == 0 {
		logger.Log.Info("没有找到需要更新状态的材质")
		return nil
	}

	logger.Log.Infof("找到 %d 个需要检查的材质...", len(textures))

	updatedCount := 0
	batchSize := 100
	
	for i := 0; i < len(textures); i += batchSize {
		end := i + batchSize
		if end > len(textures) {
			end = len(textures)
		}
		
		batch := textures[i:end]
		
		for _, texture := range batch {
			// 查询该材质的文件数量（优化：只查数量）
			var thumbnailCount int64
			var textureCount int64
			
			db.Model(&models.File{}).
				Where("related_id = ? AND related_type = ? AND file_type = ?", texture.ID, "Texture", "thumbnail").
				Count(&thumbnailCount)
			
			db.Model(&models.File{}).
				Where("related_id = ? AND related_type = ? AND file_type = ?", texture.ID, "Texture", "texture").
				Count(&textureCount)

			// 按需下载模式的状态判断
			newSyncStatus := texture.SyncStatus
			newDownloadCompleted := texture.DownloadCompleted

			if thumbnailCount > 0 && textureCount > 0 {
				// 有缩略图和贴图，已完全下载
				newSyncStatus = 2
				newDownloadCompleted = true
			} else if thumbnailCount > 0 {
				// 只有缩略图，元数据已同步，等待按需下载
				newSyncStatus = 2
				newDownloadCompleted = false
			} else if textureCount > 0 {
				// 只有贴图没有缩略图（异常情况）
				newSyncStatus = 1
				newDownloadCompleted = false
			} else {
				// 没有文件，未同步
				newSyncStatus = 0
				newDownloadCompleted = false
			}

			// 如果状态有变化，更新数据库
			if newSyncStatus != texture.SyncStatus || newDownloadCompleted != texture.DownloadCompleted {
				if err := db.Model(&texture).Updates(map[string]interface{}{
					"sync_status":        newSyncStatus,
					"download_completed": newDownloadCompleted,
				}).Error; err != nil {
					logger.Log.Warnf("更新材质状态失败 (id=%d): %v", texture.ID, err)
				} else {
					updatedCount++
				}
			}
		}
		
		if (i+batchSize) < len(textures) {
			logger.Log.Infof("已处理 %d/%d 个材质...", i+batchSize, len(textures))
		}
	}

	logger.Log.Infof("成功更新 %d 个材质的同步状态", updatedCount)
	return nil
}

// analyzeTextureTypes 分析并更新文件的贴图类型
func analyzeTextureTypes(db *gorm.DB) error {
	logger.Log.Info("开始分析文件贴图类型...")

	// 查询所有贴图文件（texture 类型）
	var files []models.File
	if err := db.Where("file_type = ? AND related_type = ?", "texture", "Texture").Find(&files).Error; err != nil {
		return err
	}

	if len(files) == 0 {
		logger.Log.Info("没有找到贴图文件")
		return nil
	}

	logger.Log.Infof("找到 %d 个贴图文件，开始分析类型...", len(files))

	// 统计所有发现的类型
	typeStats := make(map[string]int)
	updatedCount := 0

	for _, file := range files {
		// 提取贴图类型
		textureType := models.ExtractTextureType(file.FileName)
		
		// 强制更新所有文件的 texture_type（使用最新规则）
		if textureType != "" {
			if err := db.Model(&file).Update("texture_type", textureType).Error; err != nil {
				logger.Log.Warnf("更新文件贴图类型失败 (id=%d): %v", file.ID, err)
				continue
			}
			updatedCount++
			typeStats[textureType]++
			
			if updatedCount%100 == 0 {
				logger.Log.Infof("已处理 %d/%d 个文件...", updatedCount, len(files))
			}
		}
	}

	logger.Log.Infof("成功更新 %d 个文件的贴图类型", updatedCount)
	logger.Log.Info("发现的贴图类型统计:")
	for textureType, count := range typeStats {
		logger.Log.Infof("  - %s: %d 个", textureType, count)
	}

	return nil
}

// updateTextureTypesList 更新材质的贴图类型列表
func updateTextureTypesList(db *gorm.DB) error {
	logger.Log.Info("开始更新材质的贴图类型列表...")

	// 查询所有材质（强制重新更新）
	var textures []models.Texture
	if err := db.Find(&textures).Error; err != nil {
		return err
	}

	if len(textures) == 0 {
		logger.Log.Info("没有找到材质记录")
		return nil
	}

	logger.Log.Infof("找到 %d 个材质，开始更新贴图类型列表...", len(textures))

	updatedCount := 0
	batchSize := 100
	
	for i := 0; i < len(textures); i += batchSize {
		end := i + batchSize
		if end > len(textures) {
			end = len(textures)
		}
		
		batch := textures[i:end]
		
		for _, texture := range batch {
			// 查询该材质的所有文件的贴图类型（优化：只查询 texture_type 字段）
			var textureTypes []string
			if err := db.Model(&models.File{}).
				Where("related_id = ? AND related_type = ? AND file_type = ? AND texture_type != ''", 
					texture.ID, "Texture", "texture").
				Distinct("texture_type").
				Pluck("texture_type", &textureTypes).Error; err != nil {
				logger.Log.Warnf("查询材质文件失败 (texture_id=%d): %v", texture.ID, err)
				continue
			}

			// 构建新的 texture_types 字符串
			var textureTypesStr string
			if len(textureTypes) > 0 {
				textureTypesStr = joinStrings(textureTypes, ",")
			} else {
				textureTypesStr = "" // 没有贴图文件的材质设为空
			}
			
			// 强制更新材质的 texture_types 字段
			if err := db.Model(&texture).Update("texture_types", textureTypesStr).Error; err != nil {
				logger.Log.Warnf("更新材质贴图类型列表失败 (id=%d): %v", texture.ID, err)
				continue
			}
			
			updatedCount++
		}
		
		if (i+batchSize) < len(textures) {
			logger.Log.Infof("已处理 %d/%d 个材质...", i+batchSize, len(textures))
		}
	}

	logger.Log.Infof("成功更新 %d 个材质的贴图类型列表", updatedCount)
	return nil
}

// joinStrings 连接字符串数组
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// fixDuplicateTexturesInCDNPath 修复 CDN 路径中的 textures/textures/ 重复问题
func fixDuplicateTexturesInCDNPath(db *gorm.DB) error {
	logger.Log.Info("开始修复 CDN 路径中的重复 textures/ 前缀...")

	// 查找所有包含 textures/textures/ 的文件记录
	var files []models.File
	if err := db.Where("cdn_path LIKE ?", "textures/textures/%").Find(&files).Error; err != nil {
		return err
	}

	if len(files) == 0 {
		logger.Log.Info("没有找到需要修复的文件")
		return nil
	}

	logger.Log.Infof("找到 %d 个需要修复的文件，开始修复...", len(files))

	fixedCount := 0
	for _, file := range files {
		// 移除第一个 textures/ 前缀
		// textures/textures/AssetID/file.png -> textures/AssetID/file.png
		newCDNPath := file.CDNPath[9:] // 跳过前 9 个字符 "textures/"
		
		// 更新数据库
		if err := db.Model(&file).Update("cdn_path", newCDNPath).Error; err != nil {
			logger.Log.Errorf("更新文件 %d 失败: %v", file.ID, err)
			continue
		}
		
		fixedCount++
		if fixedCount%100 == 0 {
			logger.Log.Infof("已修复 %d/%d 个文件...", fixedCount, len(files))
		}
	}

	logger.Log.Infof("成功修复 %d 个文件的 CDN 路径", fixedCount)
	return nil
}

// logTextureTypeAnalysis 记录贴图类型分析结果到日志
func logTextureTypeAnalysis(db *gorm.DB) error {
	logger.Log.Info("开始记录贴图类型分析...")

	// 检查表是否存在
	if !db.Migrator().HasTable(&models.File{}) || !db.Migrator().HasTable(&models.Texture{}) {
		logger.Log.Info("相关表不存在，跳过贴图类型分析")
		return nil
	}

	analysis, err := AnalyzeAllTextureTypes(db)
	if err != nil {
		return err
	}

	if len(analysis) == 0 {
		logger.Log.Info("暂无贴图数据可分析")
		return nil
	}

	logger.Log.Info("========== 贴图类型分析报告 ==========")
	logger.Log.Infof("共发现 %d 种贴图类型", len(analysis))
	
	// 按数据源分组
	polyhavenTypes := 0
	ambientcgTypes := 0
	
	logger.Log.Info("\n--- PolyHaven 数据源 ---")
	for _, item := range analysis {
		if item.Source == "polyhaven" {
			polyhavenTypes++
			logger.Log.Infof("  %s: %d 个文件 → 建议类型: %s", 
				item.OriginalType, item.Count, item.SuggestedType)
		}
	}
	
	logger.Log.Info("\n--- AmbientCG 数据源 ---")
	for _, item := range analysis {
		if item.Source == "ambientcg" {
			ambientcgTypes++
			logger.Log.Infof("  %s: %d 个文件 → 建议类型: %s", 
				item.OriginalType, item.Count, item.SuggestedType)
		}
	}
	
	logger.Log.Infof("\nPolyHaven: %d 种类型, AmbientCG: %d 种类型", 
		polyhavenTypes, ambientcgTypes)
	logger.Log.Info("========================================")
	
	return nil
}

// cleanMissingFiles 清理物理文件不存在的数据库记录
func cleanMissingFiles(db *gorm.DB) error {
	logger.Log.Info("开始检查文件完整性...")

	// 查询所有文件记录
	var files []models.File
	if err := db.Find(&files).Error; err != nil {
		return err
	}

	if len(files) == 0 {
		logger.Log.Info("没有找到文件记录")
		return nil
	}

	logger.Log.Infof("找到 %d 个文件记录，开始检查物理文件...", len(files))

	deletedCount := 0
	checkedCount := 0
	batchSize := 100
	affectedTextureIDs := make(map[uint]bool) // 记录受影响的材质 ID

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		if end > len(files) {
			end = len(files)
		}

		batch := files[i:end]

		for _, file := range batch {
			checkedCount++

			// 跳过没有 local_path 的记录
			if file.LocalPath == "" {
				continue
			}

			// 检查物理文件是否存在
			fileExists := checkFileExists(file.LocalPath)

			if !fileExists {
				// 文件不存在，删除数据库记录
				logger.Log.Warnf("文件不存在，删除记录: %s (ID: %d)", file.LocalPath, file.ID)

				// 记录受影响的材质 ID
				if file.RelatedType == "Texture" && file.RelatedID > 0 {
					affectedTextureIDs[file.RelatedID] = true
				}

				// 删除文件记录
				if err := db.Delete(&file).Error; err != nil {
					logger.Log.Errorf("删除文件记录失败 (ID: %d): %v", file.ID, err)
					continue
				}

				deletedCount++
			}

			// 每处理 100 个文件输出一次进度
			if checkedCount%100 == 0 {
				logger.Log.Infof("已检查 %d/%d 个文件，删除 %d 个丢失记录...", 
					checkedCount, len(files), deletedCount)
			}
		}
	}

	logger.Log.Infof("文件完整性检查完成: 检查 %d 个，删除 %d 个丢失记录", 
		checkedCount, deletedCount)

	// 更新受影响材质的下载状态
	if len(affectedTextureIDs) > 0 {
		logger.Log.Infof("开始更新 %d 个受影响材质的下载状态...", len(affectedTextureIDs))
		
		updatedCount := 0
		for textureID := range affectedTextureIDs {
			// 查询该材质的文件数量
			var thumbnailCount int64
			var textureCount int64
			
			db.Model(&models.File{}).
				Where("related_id = ? AND related_type = ? AND file_type = ?", textureID, "Texture", "thumbnail").
				Count(&thumbnailCount)
			
			db.Model(&models.File{}).
				Where("related_id = ? AND related_type = ? AND file_type = ?", textureID, "Texture", "texture").
				Count(&textureCount)

			// 更新材质状态
			var texture models.Texture
			if err := db.First(&texture, textureID).Error; err != nil {
				logger.Log.Warnf("查询材质失败 (ID: %d): %v", textureID, err)
				continue
			}

			// 计算新的状态
			newDownloadCompleted := false
			newSyncStatus := texture.SyncStatus

			if thumbnailCount > 0 && textureCount > 0 {
				// 有缩略图和贴图，已完全下载
				newDownloadCompleted = true
				newSyncStatus = 2
			} else if thumbnailCount > 0 {
				// 只有缩略图，元数据已同步，等待按需下载
				newDownloadCompleted = false
				newSyncStatus = 2
			} else {
				// 没有文件，未同步
				newDownloadCompleted = false
				newSyncStatus = 0
			}

			// 如果状态有变化，更新数据库
			if newDownloadCompleted != texture.DownloadCompleted || newSyncStatus != texture.SyncStatus {
				if err := db.Model(&texture).Updates(map[string]interface{}{
					"download_completed": newDownloadCompleted,
					"sync_status":        newSyncStatus,
				}).Error; err != nil {
					logger.Log.Warnf("更新材质状态失败 (ID: %d): %v", textureID, err)
				} else {
					updatedCount++
					logger.Log.Infof("更新材质状态: ID=%d, download_completed=%v, sync_status=%d", 
						textureID, newDownloadCompleted, newSyncStatus)
				}
			}
		}

		logger.Log.Infof("成功更新 %d 个材质的下载状态", updatedCount)
	}

	return nil
}

// checkFileExists 检查文件是否存在（支持多种路径格式）
func checkFileExists(localPath string) bool {
	// 从配置获取 NAS 路径
	nasPath := ""
	if config.AppConfig != nil && config.AppConfig.Texture.NASEnabled {
		nasPath = config.AppConfig.Texture.NASPath
	}

	// 尝试多种路径格式
	paths := []string{}

	// 1. 原始路径
	paths = append(paths, localPath)

	// 2. 如果配置了 NAS 路径，使用 NAS 路径
	if nasPath != "" {
		// local_path 格式: static\textures\xxx\yyy.jpg
		// 需要提取 textures 后面的部分
		relativePath := extractRelativePath(localPath)
		if relativePath != "" {
			// 拼接 NAS 路径
			fullNASPath := joinPath(nasPath, relativePath)
			paths = append(paths, fullNASPath)
		}
	}

	// 3. 尝试当前目录的相对路径
	paths = append(paths, convertToUnixPath(localPath))

	// 检查所有可能的路径
	for _, p := range paths {
		if fileExistsAtPath(p) {
			return true
		}
	}

	return false
}

// extractRelativePath 从 local_path 提取相对路径
// 例如: static\textures\xxx\yyy.jpg -> xxx/yyy.jpg
//      static/textures/xxx/yyy.jpg -> xxx/yyy.jpg
func extractRelativePath(localPath string) string {
	// 统一转换为 Unix 格式
	path := convertToUnixPath(localPath)

	// 查找 textures/ 的位置
	texturesIndex := -1
	searchStr := "textures/"
	for i := 0; i <= len(path)-len(searchStr); i++ {
		if path[i:i+len(searchStr)] == searchStr {
			texturesIndex = i + len(searchStr)
			break
		}
	}

	if texturesIndex > 0 && texturesIndex < len(path) {
		return path[texturesIndex:]
	}

	return ""
}

// joinPath 拼接路径
func joinPath(basePath, relativePath string) string {
	// 统一转换为 Unix 格式
	base := convertToUnixPath(basePath)
	rel := convertToUnixPath(relativePath)

	// 移除 base 末尾的斜杠
	if len(base) > 0 && base[len(base)-1] == '/' {
		base = base[:len(base)-1]
	}

	// 移除 rel 开头的斜杠
	if len(rel) > 0 && rel[0] == '/' {
		rel = rel[1:]
	}

	return base + "/" + rel
}

// fileExistsAtPath 检查指定路径的文件是否存在
func fileExistsAtPath(path string) bool {
	if path == "" {
		return false
	}

	// 使用 os.Stat 检查文件
	_, err := os.Stat(path)
	return err == nil
}

// convertToUnixPath 转换为 Unix 路径格式
func convertToUnixPath(path string) string {
	// 替换反斜杠为正斜杠
	result := ""
	for _, ch := range path {
		if ch == '\\' {
			result += "/"
		} else {
			result += string(ch)
		}
	}
	return result
}
