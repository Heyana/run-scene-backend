package texture

import (
	"encoding/json"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var globalSyncService *SyncService

// SetGlobalSyncService 设置全局同步服务实例
func SetGlobalSyncService(service *SyncService) {
	globalSyncService = service
}

// GetGlobalSyncService 获取全局同步服务实例
func GetGlobalSyncService() *SyncService {
	return globalSyncService
}

// SyncService 同步服务
type SyncService struct {
	db              *gorm.DB
	downloadService *DownloadService
	tagService      *TagService
	logger          *logrus.Logger
	httpClient      *http.Client
	ticker          *time.Ticker
	stopChan        chan bool
}

// NewSyncService 创建同步服务
func NewSyncService(db *gorm.DB, logger *logrus.Logger) *SyncService {
	downloadService := NewDownloadService(db, logger)
	tagService := NewTagService(db)

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: time.Duration(config.AppConfig.Texture.APITimeout) * time.Second,
	}

	// 如果启用代理，配置代理
	if config.AppConfig.Texture.ProxyEnabled && config.AppConfig.Texture.ProxyURL != "" {
		proxyURL, err := url.Parse(config.AppConfig.Texture.ProxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			logger.Infof("贴图同步服务已启用代理: %s", config.AppConfig.Texture.ProxyURL)
		} else {
			logger.Warnf("代理 URL 解析失败: %v", err)
		}
	}

	return &SyncService{
		db:              db,
		downloadService: downloadService,
		tagService:      tagService,
		logger:          logger,
		httpClient:      client,
		stopChan:        make(chan bool),
	}
}

// FullSync 全量同步
func (s *SyncService) FullSync() error {
	s.logInfo("开始全量同步")

	// 创建同步日志
	syncLog := models.TextureSyncLog{
		SyncType:  "full",
		Status:    0, // 进行中
		StartTime: time.Now(),
	}
	if err := s.db.Create(&syncLog).Error; err != nil {
		return err
	}

	// 获取材质列表
	textureMap, err := s.fetchTextureList()
	if err != nil {
		s.updateSyncLogError(syncLog.ID, fmt.Sprintf("获取材质列表失败: %v", err))
		return err
	}

	totalCount := len(textureMap)
	syncLog.TotalCount = totalCount
	s.db.Save(&syncLog)

	s.logInfo("获取材质列表成功，共 %d 个材质", totalCount)

	// 并发处理每个材质
	successCount := 0
	failCount := 0
	skipCount := 0
	processedCount := 0

	// 创建工作队列，同时保存缩略图 URL
	type TextureJob struct {
		AssetID      string
		ThumbnailURL string
	}
	jobs := make([]TextureJob, 0, totalCount)
	debugCount := 0
	for assetID, data := range textureMap {
		job := TextureJob{AssetID: assetID}
		
		// 只对前3个材质打印详细调试信息
		if debugCount < 3 {
			s.logInfo("调试材质 %d: assetID=%s, dataType=%T", debugCount+1, assetID, data)
		}
		
		if dataMap, ok := data.(map[string]interface{}); ok {
			if thumbURL, ok := dataMap["thumbnail_url"].(string); ok {
				job.ThumbnailURL = thumbURL
				if debugCount < 3 {
					s.logInfo("✓ 成功提取缩略图 URL: %s", thumbURL)
				}
			} else {
				if debugCount < 3 {
					s.logInfo("✗ thumbnail_url 字段不存在或类型错误")
					// 打印所有可用的键
					keys := make([]string, 0, len(dataMap))
					for k := range dataMap {
						keys = append(keys, k)
					}
					s.logInfo("  可用字段: %v", keys)
				}
			}
		} else {
			if debugCount < 3 {
				s.logInfo("✗ 数据不是 map[string]interface{} 类型")
			}
		}
		
		jobs = append(jobs, job)
		debugCount++
	}
	
	s.logInfo("共提取 %d 个材质，其中 %d 个有缩略图 URL", len(jobs), func() int {
		count := 0
		for _, j := range jobs {
			if j.ThumbnailURL != "" {
				count++
			}
		}
		return count
	}())

	// 使用信号量控制并发数
	concurrency := config.AppConfig.Texture.DownloadConcurrency
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	s.logInfo("开始并发处理，并发数: %d", concurrency)

	for i, job := range jobs {
		wg.Add(1)
		semaphore <- struct{}{} // 获取信号量

		go func(index int, j TextureJob) {
			defer wg.Done()
			defer func() { <-semaphore }() // 释放信号量

			mu.Lock()
			processedCount++
			currentProcessed := processedCount
			mu.Unlock()

			s.logInfo("处理材质 [%d/%d]: %s", currentProcessed, totalCount, j.AssetID)

			// 更新进度
			s.updateProgress(syncLog.ID, currentProcessed, totalCount, j.AssetID)

			// 检查是否已完成下载
			var existingTexture models.Texture
			if err := s.db.Where("asset_id = ? AND download_completed = ?", j.AssetID, true).First(&existingTexture).Error; err == nil {
				s.logInfo("材质已完成下载，跳过: %s", j.AssetID)
				mu.Lock()
				skipCount++
				mu.Unlock()
				return
			}

			// 处理单个材质（传入缩略图 URL）
			if err := s.processTextureWithThumbnail(j.AssetID, j.ThumbnailURL); err != nil {
				s.logError("处理失败 %s: %v", err, j.AssetID)
				mu.Lock()
				failCount++
				mu.Unlock()
			} else {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i, job)
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

	s.logInfo("全量同步完成: 成功 %d, 失败 %d, 跳过 %d, 耗时 %v",
		successCount, failCount, skipCount, syncLog.EndTime.Sub(syncLog.StartTime))

	return nil
}

// IncrementalSync 增量同步
func (s *SyncService) IncrementalSync() error {
	s.logInfo("开始增量同步")

	// 创建同步日志
	syncLog := models.TextureSyncLog{
		SyncType:  "incremental",
		Status:    0,
		StartTime: time.Now(),
	}
	if err := s.db.Create(&syncLog).Error; err != nil {
		return err
	}

	// 获取材质列表
	textureList, err := s.fetchTextureList()
	if err != nil {
		s.updateSyncLogError(syncLog.ID, fmt.Sprintf("获取材质列表失败: %v", err))
		return err
	}

	// 检测新增和更新的材质
	type TextureJob struct {
		AssetID      string
		ThumbnailURL string
	}
	var needSyncJobs []TextureJob
	
	for assetID, data := range textureList {
		dataMap := data.(map[string]interface{})
		filesHash := dataMap["files_hash"].(string)

		var texture models.Texture
		err := s.db.Where("asset_id = ?", assetID).First(&texture).Error

		// 跳过已完成下载的材质
		if err == nil && texture.DownloadCompleted {
			continue
		}

		if err == gorm.ErrRecordNotFound || (err == nil && texture.FilesHash != filesHash) {
			// 新增材质或材质已更新
			job := TextureJob{AssetID: assetID}
			if thumbURL, ok := dataMap["thumbnail_url"].(string); ok {
				job.ThumbnailURL = thumbURL
			}
			needSyncJobs = append(needSyncJobs, job)
		}
	}

	totalCount := len(needSyncJobs)
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

	for i, job := range needSyncJobs {
		s.logInfo("处理材质 [%d/%d]: %s", i+1, totalCount, job.AssetID)
		s.updateProgress(syncLog.ID, i+1, totalCount, job.AssetID)

		if err := s.processTextureWithThumbnail(job.AssetID, job.ThumbnailURL); err != nil {
			s.logError("处理失败: %s - %v", err, job.AssetID)
			failCount++
		} else {
			successCount++
		}
	}

	// 更新同步日志
	syncLog.Status = 1
	syncLog.EndTime = time.Now()
	syncLog.ProcessedCount = totalCount
	syncLog.SuccessCount = successCount
	syncLog.FailCount = failCount
	syncLog.Progress = 100
	s.db.Save(&syncLog)

	s.logInfo("增量同步完成: 成功 %d, 失败 %d, 耗时 %v",
		successCount, failCount, syncLog.EndTime.Sub(syncLog.StartTime))

	return nil
}

// processTexture 处理单个材质
func (s *SyncService) processTexture(assetID string) error {
	return s.processTextureWithThumbnail(assetID, "")
}

// processTextureWithThumbnail 处理单个材质（带缩略图 URL）
func (s *SyncService) processTextureWithThumbnail(assetID string, thumbnailURL string) error {
	// 获取材质详情
	detail, err := s.fetchTextureDetail(assetID)
	if err != nil {
		// 标记为失败状态
		var texture models.Texture
		if dbErr := s.db.Where("asset_id = ?", assetID).First(&texture).Error; dbErr == nil {
			texture.SyncStatus = 3 // 失败
			s.db.Save(&texture)
		}
		return fmt.Errorf("获取详情失败: %w", err)
	}

	// 保存材质元数据
	texture, err := s.saveTexture(assetID, detail)
	if err != nil {
		// 标记为失败状态
		if texture != nil {
			texture.SyncStatus = 3 // 失败
			s.db.Save(texture)
		}
		return fmt.Errorf("保存元数据失败: %w", err)
	}

	// 处理标签
	if err := s.processTags(texture.ID, detail); err != nil {
		s.logError("处理标签失败: %v", err)
	}

	// 标记下载状态
	thumbnailDownloaded := false
	texturesDownloaded := false

	// 下载缩略图
	if config.AppConfig.Texture.DownloadThumbnail && thumbnailURL != "" {
		s.logInfo("下载缩略图: %s, URL: %s", assetID, thumbnailURL)
		if _, err := s.downloadService.DownloadThumbnail(texture.ID, assetID, thumbnailURL); err != nil {
			s.logError("下载缩略图失败: %v", err)
		} else {
			s.logInfo("缩略图下载成功: %s", assetID)
			thumbnailDownloaded = true
		}
	} else {
		s.logDebug("跳过缩略图下载: enabled=%v, url=%s", config.AppConfig.Texture.DownloadThumbnail, thumbnailURL)
		thumbnailDownloaded = true // 如果不需要下载，也标记为已完成
	}

	// 下载贴图文件
	if config.AppConfig.Texture.DownloadTextures {
		// 获取文件列表
		filesDetail, err := s.fetchTextureFiles(assetID)
		if err != nil {
			s.logError("获取文件列表失败: %v", err)
		} else {
			if err := s.downloadService.DownloadAndConvert(texture.ID, assetID, filesDetail); err != nil {
				s.logError("下载贴图失败: %v", err)
			} else {
				texturesDownloaded = true
			}
		}
	} else {
		texturesDownloaded = true // 如果不需要下载，也标记为已完成
	}

	// 更新同步状态
	// 如果缩略图和贴图都下载成功，标记为已完成
	if thumbnailDownloaded && texturesDownloaded {
		texture.SyncStatus = 2 // 已同步
		texture.DownloadCompleted = true
		s.logInfo("材质下载完成: %s", assetID)
	} else if !thumbnailDownloaded || !texturesDownloaded {
		// 如果有任何下载失败，标记为失败
		texture.SyncStatus = 3 // 失败
		s.logInfo("材质下载未完成: %s (缩略图: %v, 贴图: %v)", assetID, thumbnailDownloaded, texturesDownloaded)
	} else {
		texture.SyncStatus = 1 // 同步中（部分完成）
	}
	
	s.db.Save(texture)

	return nil
}

// fetchTextureList 获取材质列表（带重试）
func (s *SyncService) fetchTextureList() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/assets?type=textures", config.AppConfig.Texture.APIBaseURL)
	
	var lastErr error
	maxRetries := config.AppConfig.Texture.RetryTimes
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(i) * 2 * time.Second
			s.logInfo("重试获取材质列表 (%d/%d)，等待 %v...", i+1, maxRetries, waitTime)
			time.Sleep(waitTime)
		}

		resp, err := s.httpClient.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("请求失败: %w", err)
			s.logWarn("获取材质列表失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
			continue
		}

		// 读取响应体
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			lastErr = fmt.Errorf("读取响应失败: %w", err)
			s.logWarn("读取响应失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
			continue
		}

		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("API返回错误状态 %d: %s", resp.StatusCode, string(body))
			s.logWarn("API返回错误 (尝试 %d/%d): %v", i+1, maxRetries, lastErr)
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			lastErr = fmt.Errorf("解析JSON失败: %w", err)
			s.logWarn("解析JSON失败 (尝试 %d/%d): %v", i+1, maxRetries, err)
			continue
		}

		// 成功
		return result, nil
	}

	return nil, fmt.Errorf("获取材质列表失败，已重试 %d 次: %w", maxRetries, lastErr)
}

// fetchTextureDetail 获取材质详情
func (s *SyncService) fetchTextureDetail(assetID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/files/%s", config.AppConfig.Texture.APIBaseURL, assetID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return result, nil
}

// fetchTextureFiles 获取材质文件列表
func (s *SyncService) fetchTextureFiles(assetID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/files/%s", config.AppConfig.Texture.APIBaseURL, assetID)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// saveTexture 保存材质元数据
func (s *SyncService) saveTexture(assetID string, data map[string]interface{}) (*models.Texture, error) {
	var texture models.Texture
	err := s.db.Where("asset_id = ?", assetID).First(&texture).Error

	isNew := err == gorm.ErrRecordNotFound

	if isNew {
		texture.AssetID = assetID
	}

	// 更新字段
	if name, ok := data["name"].(string); ok {
		texture.Name = name
	}
	if desc, ok := data["description"].(string); ok {
		texture.Description = desc
	}
	if typeVal, ok := data["type"].(float64); ok {
		texture.Type = int(typeVal)
	}
	if filesHash, ok := data["files_hash"].(string); ok {
		texture.FilesHash = filesHash
	}
	if datePub, ok := data["date_published"].(float64); ok {
		texture.DatePublished = int64(datePub)
	}
	if downloadCount, ok := data["download_count"].(float64); ok {
		texture.DownloadCount = int(downloadCount)
	}

	// Authors 转 JSON
	if authors, ok := data["authors"].(map[string]interface{}); ok {
		if authorsJSON, err := json.Marshal(authors); err == nil {
			texture.Authors = string(authorsJSON)
		}
	}

	// MaxResolution
	if maxRes, ok := data["max_resolution"].([]interface{}); ok && len(maxRes) == 2 {
		texture.MaxResolution = fmt.Sprintf("%.0fx%.0f", maxRes[0], maxRes[1])
	}

	if isNew {
		if err := s.db.Create(&texture).Error; err != nil {
			return nil, err
		}
	} else {
		if err := s.db.Save(&texture).Error; err != nil {
			return nil, err
		}
	}

	return &texture, nil
}

// processTags 处理标签和分类
func (s *SyncService) processTags(textureID uint, data map[string]interface{}) error {
	var tagIDs []uint

	// 处理 tags
	if tags, ok := data["tags"].([]interface{}); ok {
		for _, tagName := range tags {
			if name, ok := tagName.(string); ok {
				tag, err := s.tagService.GetOrCreateTag(name, "tag")
				if err != nil {
					s.logError("创建标签失败: %v", err)
					continue
				}
				tagIDs = append(tagIDs, tag.ID)
				s.logDebug("创建/获取标签: %s (type=tag)", name)
			}
		}
	}

	// 处理 categories
	if categories, ok := data["categories"].([]interface{}); ok {
		for _, catName := range categories {
			if name, ok := catName.(string); ok {
				tag, err := s.tagService.GetOrCreateTag(name, "category")
				if err != nil {
					s.logError("创建分类失败: %v", err)
					continue
				}
				tagIDs = append(tagIDs, tag.ID)
				s.logDebug("创建/获取分类: %s (type=category)", name)
			}
		}
	}

	// 关联标签
	if len(tagIDs) > 0 {
		if err := s.tagService.AssociateTextureTags(textureID, tagIDs); err != nil {
			return err
		}
	}

	return nil
}

// updateProgress 更新同步进度
func (s *SyncService) updateProgress(logID uint, processed int, total int, currentAsset string) error {
	progress := float64(processed) / float64(total) * 100
	return s.db.Model(&models.TextureSyncLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
		"processed_count": processed,
		"current_asset":   currentAsset,
		"progress":        progress,
	}).Error
}

// updateSyncLogError 更新同步日志错误
func (s *SyncService) updateSyncLogError(logID uint, errorMsg string) {
	s.db.Model(&models.TextureSyncLog{}).Where("id = ?", logID).Updates(map[string]interface{}{
		"status":    2, // 失败
		"error_msg": errorMsg,
		"end_time":  time.Now(),
	})
}

// 日志方法
func (s *SyncService) logInfo(format string, args ...interface{}) {
	if config.AppConfig.Texture.LogEnabled {
		s.logger.Infof(format, args...)
	}
}

func (s *SyncService) logWarn(format string, args ...interface{}) {
	if config.AppConfig.Texture.LogEnabled {
		s.logger.Warnf(format, args...)
	}
}

func (s *SyncService) logError(format string, err error, args ...interface{}) {
	if config.AppConfig.Texture.LogEnabled {
		allArgs := append([]interface{}{err}, args...)
		s.logger.Errorf(format, allArgs...)
	}
}

func (s *SyncService) logDebug(format string, args ...interface{}) {
	if config.AppConfig.Texture.LogEnabled {
		s.logger.Debugf(format, args...)
	}
}

// StartScheduler 启动定时任务
func (s *SyncService) StartScheduler() {
	interval, err := time.ParseDuration(config.AppConfig.Texture.SyncInterval)
	if err != nil {
		s.logger.Errorf("解析同步间隔失败: %v", err)
		return
	}

	s.logInfo("启动定时同步任务，间隔: %v", interval)

	s.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.logInfo("定时任务触发，开始增量同步")
				if err := s.IncrementalSync(); err != nil {
					s.logError("定时同步失败: %v", err)
				}
			case <-s.stopChan:
				s.logInfo("定时同步任务已停止")
				return
			}
		}
	}()
}

// StopScheduler 停止定时任务
func (s *SyncService) StopScheduler() {
	if s.ticker != nil {
		s.ticker.Stop()
		close(s.stopChan)
		s.logInfo("定时同步任务停止信号已发送")
	}
}
