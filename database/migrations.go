// Package database 提供数据库迁移和升级工具
package database

import (
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"

	"gorm.io/gorm"
)

// RunOnceUpgrade 运行一次性升级任务
func RunOnceUpgrade(db *gorm.DB) error {
	logger.Log.Info("开始执行一次性升级任务...")

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

	// 5. 分析并更新文件的贴图类型
	if err := analyzeTextureTypes(db); err != nil {
		logger.Log.Errorf("分析贴图类型失败: %v", err)
	}

	// 6. 更新材质的贴图类型列表
	if err := updateTextureTypesList(db); err != nil {
		logger.Log.Errorf("更新材质贴图类型列表失败: %v", err)
	}

	logger.Log.Info("所有升级任务执行完成")
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

	// 查找所有材质
	var textures []models.Texture
	if err := db.Find(&textures).Error; err != nil {
		return err
	}

	if len(textures) == 0 {
		logger.Log.Info("没有找到材质记录")
		return nil
	}

	logger.Log.Infof("找到 %d 个材质，开始检查文件状态...", len(textures))

	updatedCount := 0
	for _, texture := range textures {
		// 查询该材质的所有文件
		var files []models.File
		if err := db.Where("related_id = ? AND related_type = ?", texture.ID, "Texture").Find(&files).Error; err != nil {
			logger.Log.Warnf("查询材质文件失败 (texture_id=%d): %v", texture.ID, err)
			continue
		}

		// 检查是否有缩略图和贴图文件
		hasThumbnail := false
		hasTextures := false

		for _, file := range files {
			if file.FileType == "thumbnail" {
				hasThumbnail = true
			} else if file.FileType == "texture" {
				hasTextures = true
			}
		}

		// 根据文件情况更新状态
		newSyncStatus := texture.SyncStatus
		newDownloadCompleted := texture.DownloadCompleted

		if len(files) == 0 {
			// 没有文件，未同步
			newSyncStatus = 0
			newDownloadCompleted = false
		} else if hasThumbnail && hasTextures {
			// 有缩略图和贴图，已同步完成
			newSyncStatus = 2
			newDownloadCompleted = true
		} else if hasThumbnail || hasTextures {
			// 只有部分文件，同步中
			newSyncStatus = 1
			newDownloadCompleted = false
		} else {
			// 有文件但类型不明确，标记为同步中
			newSyncStatus = 1
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
				logger.Log.Debugf("更新材质 %s: sync_status=%d, download_completed=%v (文件数=%d, 缩略图=%v, 贴图=%v)",
					texture.AssetID, newSyncStatus, newDownloadCompleted, len(files), hasThumbnail, hasTextures)
			}
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
		
		// 如果提取到了类型且与当前不同，则更新
		if textureType != "" && textureType != file.TextureType {
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

	// 查询所有材质
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
	for _, texture := range textures {
		// 查询该材质的所有文件
		var files []models.File
		if err := db.Where("related_id = ? AND related_type = ? AND file_type = ?", 
			texture.ID, "Texture", "texture").Find(&files).Error; err != nil {
			logger.Log.Warnf("查询材质文件失败 (texture_id=%d): %v", texture.ID, err)
			continue
		}

		// 收集所有唯一的贴图类型
		typeSet := make(map[string]bool)
		for _, file := range files {
			if file.TextureType != "" {
				typeSet[file.TextureType] = true
			}
		}

		// 转换为逗号分隔的字符串
		var types []string
		for textureType := range typeSet {
			types = append(types, textureType)
		}

		if len(types) > 0 {
			textureTypes := joinStrings(types, ",")
			
			// 更新材质的 texture_types 字段
			if err := db.Model(&texture).Update("texture_types", textureTypes).Error; err != nil {
				logger.Log.Warnf("更新材质贴图类型列表失败 (id=%d): %v", texture.ID, err)
				continue
			}
			
			updatedCount++
			if updatedCount%50 == 0 {
				logger.Log.Infof("已处理 %d/%d 个材质...", updatedCount, len(textures))
			}
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
