package texture

import (
	"fmt"
	"go_wails_project_manager/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AmbientCGSyncService AmbientCG 同步服务（按需下载模式）
type AmbientCGSyncService struct {
	db         *gorm.DB
	logger     *logrus.Logger
	adapter    *AmbientCGAdapter
	httpClient *http.Client
}

// NewAmbientCGSyncService 创建 AmbientCG 同步服务
func NewAmbientCGSyncService(db *gorm.DB, logger *logrus.Logger) *AmbientCGSyncService {
	adapter := NewAmbientCGAdapter("https://ambientcg.com", 30*time.Second)

	return &AmbientCGSyncService{
		db:      db,
		logger:  logger,
		adapter: adapter,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SyncMetadata 同步元数据（只下载预览图，不下载贴图）
func (s *AmbientCGSyncService) SyncMetadata() error {
	s.logInfo("开始 AmbientCG 元数据同步（按需下载模式）")

	// 创建同步日志
	syncLog := models.TextureSyncLog{
		SyncType:  "ambientcg_metadata",
		Status:    0, // 进行中
		StartTime: time.Now(),
	}
	if err := s.db.Create(&syncLog).Error; err != nil {
		return err
	}

	// 获取总数
	firstPage, err := s.adapter.GetMaterialList(1, 0)
	if err != nil {
		s.updateSyncLogError(syncLog.ID, fmt.Sprintf("获取材质列表失败: %v", err))
		return err
	}

	totalCount := firstPage.NumberOfResults
	syncLog.TotalCount = totalCount
	s.db.Save(&syncLog)

	s.logInfo("AmbientCG 共有 %d 个材质", totalCount)

	// 分页获取所有材质
	limit := 100
	offset := 0
	successCount := 0
	failCount := 0
	skipCount := 0
	processedCount := 0

	// 并发控制
	concurrency := 5 // AmbientCG 并发数
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for offset < totalCount {
		// 获取当前页
		page, err := s.adapter.GetMaterialList(limit, offset)
		if err != nil {
			s.logError("获取第 %d 页失败: %v", err, offset/limit+1)
			offset += limit
			continue
		}

		s.logInfo("获取第 %d 页，共 %d 个材质", offset/limit+1, len(page.FoundAssets))

		// 处理每个材质
		for _, material := range page.FoundAssets {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(mat AmbientCGMaterial) {
				defer wg.Done()
				defer func() { <-semaphore }()

				mu.Lock()
				processedCount++
				currentProcessed := processedCount
				mu.Unlock()

				s.logInfo("处理材质 [%d/%d]: %s", currentProcessed, totalCount, mat.AssetID)

				// 更新进度
				s.updateProgress(syncLog.ID, currentProcessed, totalCount, mat.AssetID)

				// 检查是否已存在
				var existing models.Texture
				err := s.db.Where("asset_id = ? AND source = ?", mat.AssetID, "ambientcg").First(&existing).Error
				if err == nil {
					// 已存在，跳过
					s.logDebug("材质已存在，跳过: %s", mat.AssetID)
					mu.Lock()
					skipCount++
					mu.Unlock()
					return
				}

				// 保存元数据
				if err := s.saveMetadata(&mat); err != nil {
					s.logError("保存元数据失败 %s: %v", err, mat.AssetID)
					mu.Lock()
					failCount++
					mu.Unlock()
					return
				}

				mu.Lock()
				successCount++
				mu.Unlock()
			}(material)
		}

		offset += limit
		time.Sleep(100 * time.Millisecond) // 避免请求过快
	}

	// 等待所有任务完成
	wg.Wait()
	s.logInfo("所有材质处理完成")

	// 更新同步日志
	syncLog.Status = 1 // 成功
	syncLog.EndTime = time.Now()
	syncLog.ProcessedCount = totalCount
	syncLog.SuccessCount = successCount
	syncLog.FailCount = failCount
	syncLog.SkipCount = skipCount
	syncLog.Progress = 100
	s.db.Save(&syncLog)

	s.logInfo("AmbientCG 元数据同步完成: 成功 %d, 失败 %d, 跳过 %d, 耗时 %v",
		successCount, failCount, skipCount, syncLog.EndTime.Sub(syncLog.StartTime))

	return nil
}

// saveMetadata 保存材质元数据和预览图
func (s *AmbientCGSyncService) saveMetadata(material *AmbientCGMaterial) error {
	// 1. 保存材质元数据
	texture := models.Texture{
		AssetID:           material.AssetID,
		Name:              material.DisplayName,
		Description:       material.Description,
		Source:            "ambientcg",
		DownloadCompleted: false, // ⚠️ 关键：标记未下载
		SyncStatus:        2,      // 已同步元数据
		DatePublished:     s.parseDate(material.ReleaseDate),
		DownloadCount:     material.DownloadCount,
		TextureTypes:      strings.Join(material.Maps, ","),
		Type:              s.mapCategory(material.DisplayCategory),
		Authors:           "AmbientCG",
		MaxResolution:     "8K", // AmbientCG 最高支持 8K
		FilesHash:         material.AssetID,
	}

	if err := s.db.Create(&texture).Error; err != nil {
		return fmt.Errorf("创建材质记录失败: %w", err)
	}

	s.logDebug("材质元数据已保存: %s (ID: %d)", texture.AssetID, texture.ID)

	// 2. 下载预览图
	previewURL := s.selectPreviewURL(material.PreviewImage)
	if previewURL != "" {
		if err := s.downloadPreview(texture.ID, texture.AssetID, previewURL); err != nil {
			s.logError("下载预览图失败 %s: %v", err, texture.AssetID)
			// 不影响主流程，继续
		} else {
			s.logDebug("预览图下载成功: %s", texture.AssetID)
		}
	}

	// 3. 处理标签
	if err := s.processTags(texture.ID, material.Tags, material.DisplayCategory); err != nil {
		s.logError("处理标签失败: %v", err)
		// 不影响主流程
	}

	return nil
}

// downloadPreview 下载预览图
func (s *AmbientCGSyncService) downloadPreview(textureID uint, assetID, previewURL string) error {
	// 创建目录
	dir := filepath.Join("static", "textures", assetID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 下载文件
	savePath := filepath.Join(dir, "thumbnail.png")
	if err := s.downloadFile(previewURL, savePath); err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(savePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 保存到数据库
	file := models.File{
		FileType:    "thumbnail",
		RelatedID:   textureID,
		RelatedType: "Texture",
		OriginalURL: previewURL,
		LocalPath:   savePath,
		CDNPath:     filepath.Join("textures", assetID, "thumbnail.png"),
		FileName:    "thumbnail.png",
		FileSize:    fileInfo.Size(),
		Format:      "png",
		Status:      1, // 已下载
	}

	if err := s.db.Create(&file).Error; err != nil {
		return fmt.Errorf("保存文件记录失败: %w", err)
	}

	return nil
}

// downloadFile 下载文件
func (s *AmbientCGSyncService) downloadFile(url, savePath string) error {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP 状态码: %d", resp.StatusCode)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// selectPreviewURL 选择合适的预览图 URL
func (s *AmbientCGSyncService) selectPreviewURL(previewImage map[string]string) string {
	// 优先选择 512-PNG，其次 256-PNG
	if url, ok := previewImage["512-PNG"]; ok && url != "" {
		return url
	}
	if url, ok := previewImage["256-PNG"]; ok && url != "" {
		return url
	}
	// 如果都没有，返回第一个可用的
	for _, url := range previewImage {
		if url != "" {
			return url
		}
	}
	return ""
}

// mapCategory 映射分类到类型 ID
func (s *AmbientCGSyncService) mapCategory(category string) int {
	categoryMap := map[string]int{
		"Ground":        1,
		"Wood":          2,
		"Grass":         3,
		"Paving Stones": 4,
		"Fabric":        5,
		"Concrete":      6,
		"Metal":         7,
		"Brick":         8,
		"Tiles":         9,
		"Rock":          10,
		"Marble":        11,
		"Leather":       12,
		"Plastic":       13,
	}

	if id, ok := categoryMap[category]; ok {
		return id
	}
	return 0 // 未知分类
}

// parseDate 解析日期字符串
func (s *AmbientCGSyncService) parseDate(dateStr string) int64 {
	// AmbientCG 日期格式: "2026-01-12 17:00:00"
	t, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		return 0
	}
	return t.Unix()
}

// processTags 处理标签
func (s *AmbientCGSyncService) processTags(textureID uint, tags []string, category string) error {
	tagService := NewTagService(s.db)
	var tagIDs []uint

	// 处理标签
	for _, tagName := range tags {
		tag, err := tagService.GetOrCreateTag(tagName, "tag")
		if err != nil {
			s.logError("创建标签失败: %v", err)
			continue
		}
		tagIDs = append(tagIDs, tag.ID)
	}

	// 处理分类
	if category != "" {
		tag, err := tagService.GetOrCreateTag(category, "category")
		if err == nil {
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	// 关联标签
	if len(tagIDs) > 0 {
		if err := tagService.AssociateTextureTags(textureID, tagIDs); err != nil {
			return err
		}
	}

	return nil
}

// updateProgress 更新同步进度
func (s *AmbientCGSyncService) updateProgress(logID uint, processed int, total int, currentAsset string) error {
	progress := float64(processed) / float64(total) * 100
	return s.db.Model(&models.TextureSyncLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
		"processed_count": processed,
		"current_asset":   currentAsset,
		"progress":        progress,
	}).Error
}

// updateSyncLogError 更新同步日志错误
func (s *AmbientCGSyncService) updateSyncLogError(logID uint, errorMsg string) {
	s.db.Model(&models.TextureSyncLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
		"status":    2, // 失败
		"error_msg": errorMsg,
		"end_time":  time.Now(),
	})
}

// 日志方法
func (s *AmbientCGSyncService) logInfo(format string, args ...interface{}) {
	s.logger.Infof("[AmbientCG] "+format, args...)
}

func (s *AmbientCGSyncService) logWarn(format string, args ...interface{}) {
	s.logger.Warnf("[AmbientCG] "+format, args...)
}

func (s *AmbientCGSyncService) logError(format string, err error, args ...interface{}) {
	allArgs := append([]interface{}{err}, args...)
	s.logger.Errorf("[AmbientCG] "+format, allArgs...)
}

func (s *AmbientCGSyncService) logDebug(format string, args ...interface{}) {
	s.logger.Debugf("[AmbientCG] "+format, args...)
}


// IncrementalSync 增量同步（检测新增和更新的材质）
func (s *AmbientCGSyncService) IncrementalSync() error {
	s.logInfo("开始 AmbientCG 增量同步")

	// 创建同步日志
	syncLog := models.TextureSyncLog{
		SyncType:  "ambientcg_incremental",
		Status:    0, // 进行中
		StartTime: time.Now(),
	}
	if err := s.db.Create(&syncLog).Error; err != nil {
		return err
	}

	// 获取最新的材质列表（按最新发布排序）
	limit := 100
	offset := 0
	needSyncMaterials := []AmbientCGMaterial{}

	// 获取本地最新的发布时间
	var latestTexture models.Texture
	err := s.db.Where("source = ?", "ambientcg").Order("date_published DESC").First(&latestTexture).Error
	var latestDate time.Time
	if err == nil {
		latestDate = time.Unix(latestTexture.DatePublished, 0)
		s.logInfo("本地最新材质发布时间: %s", latestDate.Format("2006-01-02"))
	}

	// 获取最新的材质（最多检查 500 个）
	maxCheck := 500
	for offset < maxCheck {
		page, err := s.adapter.GetMaterialList(limit, offset)
		if err != nil {
			s.logError("获取材质列表失败: %v", err)
			break
		}

		if len(page.FoundAssets) == 0 {
			break
		}

		// 检查每个材质
		for _, material := range page.FoundAssets {
			// 解析发布时间
			releaseDate := s.parseDate(material.ReleaseDate)
			materialDate := time.Unix(releaseDate, 0)

			// 如果材质发布时间早于本地最新时间，停止检查
			if !latestDate.IsZero() && materialDate.Before(latestDate) {
				s.logInfo("遇到旧材质，停止检查: %s (%s)", material.AssetID, materialDate.Format("2006-01-02"))
				offset = maxCheck // 跳出外层循环
				break
			}

			// 检查是否已存在
			var existing models.Texture
			err := s.db.Where("asset_id = ? AND source = ?", material.AssetID, "ambientcg").First(&existing).Error
			
			if err == gorm.ErrRecordNotFound {
				// 新材质
				needSyncMaterials = append(needSyncMaterials, material)
				s.logDebug("发现新材质: %s", material.AssetID)
			} else if err == nil {
				// 已存在，检查是否需要更新（比较发布时间）
				if releaseDate > existing.DatePublished {
					needSyncMaterials = append(needSyncMaterials, material)
					s.logDebug("发现更新材质: %s", material.AssetID)
				}
			}
		}

		offset += limit
		time.Sleep(100 * time.Millisecond)
	}

	totalCount := len(needSyncMaterials)
	syncLog.TotalCount = totalCount
	s.db.Save(&syncLog)

	s.logInfo("检测到 %d 个需要同步的材质", totalCount)

	if totalCount == 0 {
		syncLog.Status = 1
		syncLog.EndTime = time.Now()
		syncLog.Progress = 100
		s.db.Save(&syncLog)
		s.logInfo("无需同步")
		return nil
	}

	// 处理需要同步的材质
	successCount := 0
	failCount := 0
	processedCount := 0

	// 并发控制
	concurrency := 5
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, material := range needSyncMaterials {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(mat AmbientCGMaterial) {
			defer wg.Done()
			defer func() { <-semaphore }()

			mu.Lock()
			processedCount++
			currentProcessed := processedCount
			mu.Unlock()

			s.logInfo("处理材质 [%d/%d]: %s", currentProcessed, totalCount, mat.AssetID)
			s.updateProgress(syncLog.ID, currentProcessed, totalCount, mat.AssetID)

			// 保存或更新元数据
			if err := s.saveOrUpdateMetadata(&mat); err != nil {
				s.logError("保存元数据失败 %s: %v", err, mat.AssetID)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(material)
	}

	// 等待所有任务完成
	wg.Wait()

	// 更新同步日志
	syncLog.Status = 1
	syncLog.EndTime = time.Now()
	syncLog.ProcessedCount = totalCount
	syncLog.SuccessCount = successCount
	syncLog.FailCount = failCount
	syncLog.Progress = 100
	s.db.Save(&syncLog)

	s.logInfo("AmbientCG 增量同步完成: 成功 %d, 失败 %d, 耗时 %v",
		successCount, failCount, syncLog.EndTime.Sub(syncLog.StartTime))

	return nil
}

// saveOrUpdateMetadata 保存或更新材质元数据
func (s *AmbientCGSyncService) saveOrUpdateMetadata(material *AmbientCGMaterial) error {
	// 检查是否已存在
	var texture models.Texture
	err := s.db.Where("asset_id = ? AND source = ?", material.AssetID, "ambientcg").First(&texture).Error

	isNew := err == gorm.ErrRecordNotFound

	if isNew {
		// 新建
		texture = models.Texture{
			AssetID:           material.AssetID,
			Name:              material.DisplayName,
			Description:       material.Description,
			Source:            "ambientcg",
			DownloadCompleted: false,
			SyncStatus:        2, // 已同步元数据
			DatePublished:     s.parseDate(material.ReleaseDate),
			DownloadCount:     material.DownloadCount,
			TextureTypes:      strings.Join(material.Maps, ","),
			Type:              s.mapCategory(material.DisplayCategory),
			Authors:           "AmbientCG",
			MaxResolution:     "8K",
			FilesHash:         material.AssetID,
		}

		if err := s.db.Create(&texture).Error; err != nil {
			return fmt.Errorf("创建材质记录失败: %w", err)
		}

		s.logDebug("新材质已创建: %s (ID: %d)", texture.AssetID, texture.ID)
	} else {
		// 更新
		texture.Name = material.DisplayName
		texture.Description = material.Description
		texture.DatePublished = s.parseDate(material.ReleaseDate)
		texture.DownloadCount = material.DownloadCount
		texture.TextureTypes = strings.Join(material.Maps, ",")
		texture.Type = s.mapCategory(material.DisplayCategory)

		if err := s.db.Save(&texture).Error; err != nil {
			return fmt.Errorf("更新材质记录失败: %w", err)
		}

		s.logDebug("材质已更新: %s (ID: %d)", texture.AssetID, texture.ID)
	}

	// 下载预览图（如果还没有）
	var existingPreview models.File
	err = s.db.Where("related_id = ? AND related_type = ? AND file_type = ?",
		texture.ID, "Texture", "thumbnail").First(&existingPreview).Error

	if err == gorm.ErrRecordNotFound {
		previewURL := s.selectPreviewURL(material.PreviewImage)
		if previewURL != "" {
			if err := s.downloadPreview(texture.ID, texture.AssetID, previewURL); err != nil {
				s.logError("下载预览图失败 %s: %v", err, texture.AssetID)
			}
		}
	}

	// 处理标签
	if err := s.processTags(texture.ID, material.Tags, material.DisplayCategory); err != nil {
		s.logError("处理标签失败: %v", err)
	}

	return nil
}
