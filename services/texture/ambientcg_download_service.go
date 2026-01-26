package texture

import (
	"archive/zip"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AmbientCGDownloadService AmbientCG 按需下载服务
type AmbientCGDownloadService struct {
	db                  *gorm.DB
	logger              *logrus.Logger
	adapter             *AmbientCGAdapter
	httpClient          *http.Client
	downloadService     *DownloadService // 使用统一的下载服务
	localStorageEnabled bool
	storageDir          string
	nasEnabled          bool
	nasPath             string
}

// NewAmbientCGDownloadService 创建下载服务
func NewAmbientCGDownloadService(db *gorm.DB, logger *logrus.Logger) *AmbientCGDownloadService {
	adapter := NewAmbientCGAdapter("https://ambientcg.com", 120*time.Second) // 下载超时时间更长
	downloadService := NewDownloadService(db, logger)                        // 创建统一下载服务

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	// 如果启用代理，配置代理
	if config.AppConfig.Texture.ProxyEnabled && config.AppConfig.Texture.ProxyURL != "" {
		proxyURL, err := url.Parse(config.AppConfig.Texture.ProxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
			logger.Infof("[AmbientCG Download] 已启用代理: %s", config.AppConfig.Texture.ProxyURL)
		} else {
			logger.Warnf("[AmbientCG Download] 代理 URL 解析失败: %v", err)
		}
	}

	// 获取存储配置
	localStorageEnabled := config.AppConfig.Texture.LocalStorageEnabled
	storageDir := config.AppConfig.Texture.StorageDir
	nasEnabled := config.AppConfig.Texture.NASEnabled
	nasPath := config.AppConfig.Texture.NASPath

	return &AmbientCGDownloadService{
		db:                  db,
		logger:              logger,
		adapter:             adapter,
		httpClient:          client,
		downloadService:     downloadService,
		localStorageEnabled: localStorageEnabled,
		storageDir:          storageDir,
		nasEnabled:          nasEnabled,
		nasPath:             nasPath,
	}
}

// DownloadOptions 下载选项
type DownloadOptions struct {
	Resolution string // 1K, 2K, 4K, 8K
	Format     string // JPG, PNG
}

// DownloadTexture 下载材质包
func (s *AmbientCGDownloadService) DownloadTexture(assetID string, opts DownloadOptions) ([]models.File, error) {
	s.logInfo("开始下载材质: %s (分辨率: %s, 格式: %s)", assetID, opts.Resolution, opts.Format)

	// 1. 获取材质信息
	var texture models.Texture
	if err := s.db.Where("asset_id = ? AND source = ?", assetID, "ambientcg").First(&texture).Error; err != nil {
		return nil, fmt.Errorf("材质不存在: %w", err)
	}

	// 2. 检查是否已下载
	if texture.DownloadCompleted {
		s.logInfo("材质已下载，返回现有文件: %s", assetID)
		return s.getExistingFiles(texture.ID)
	}

	// 3. 获取下载链接
	detail, err := s.adapter.GetMaterialDetail(assetID)
	if err != nil {
		return nil, fmt.Errorf("获取材质详情失败: %w", err)
	}

	downloads, err := s.adapter.GetDownloads(detail)
	if err != nil {
		return nil, fmt.Errorf("获取下载列表失败: %w", err)
	}

	// 4. 选择合适的下载包
	selectedDownload := s.adapter.SelectBestDownload(downloads, opts.Resolution, opts.Format)
	if selectedDownload == nil {
		return nil, fmt.Errorf("未找到合适的下载包")
	}

	s.logInfo("选择下载包: %s (大小: %.2f MB)", selectedDownload.FileName, float64(selectedDownload.Size)/1024/1024)

	// 5. 下载 ZIP 包
	tempDir := filepath.Join("temp", "ambientcg")
	os.MkdirAll(tempDir, 0755)
	zipPath := filepath.Join(tempDir, selectedDownload.FileName)

	if err := s.downloadFile(selectedDownload.DownloadLink, zipPath); err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}
	defer os.Remove(zipPath) // 下载完成后删除临时文件

	s.logInfo("下载完成: %s", zipPath)

	// 6. 解压到目标目录（根据配置选择存储位置）
	var extractPath string
	if s.nasEnabled && s.nasPath != "" {
		// 使用 NAS 路径
		extractPath = filepath.Join(s.nasPath, assetID)
	} else {
		// 使用本地路径
		extractPath = filepath.Join(s.storageDir, assetID)
	}
	
	if err := s.extractZip(zipPath, extractPath); err != nil {
		return nil, fmt.Errorf("解压失败: %w", err)
	}

	s.logInfo("解压完成: %s", extractPath)

	// 7. 解析文件并保存到数据库
	files, err := s.parseAndSaveFiles(texture.ID, assetID, extractPath)
	if err != nil {
		return nil, fmt.Errorf("解析文件失败: %w", err)
	}

	s.logInfo("解析并保存 %d 个文件", len(files))

	// 8. 更新材质状态
	texture.DownloadCompleted = true
	texture.SyncStatus = 2 // 已同步
	if err := s.db.Save(&texture).Error; err != nil {
		return nil, fmt.Errorf("更新材质状态失败: %w", err)
	}

	s.logInfo("材质下载完成: %s", assetID)

	return files, nil
}

// downloadFile 下载文件
func (s *AmbientCGDownloadService) downloadFile(url, savePath string) error {
	s.logInfo("正在下载: %s", url)

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

	// 显示下载进度
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	s.logInfo("下载完成: %.2f MB", float64(written)/1024/1024)
	return nil
}

// extractZip 解压 ZIP 文件
func (s *AmbientCGDownloadService) extractZip(zipPath, destPath string) error {
	s.logInfo("正在解压: %s -> %s", zipPath, destPath)

	// 打开 ZIP 文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 创建目标目录
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}

	// 解压每个文件
	for _, file := range reader.File {
		path := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// 创建文件目录
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// 打开 ZIP 中的文件
		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		// 创建目标文件
		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			srcFile.Close()
			return err
		}

		// 复制内容
		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			return err
		}
	}

	s.logInfo("解压完成，共 %d 个文件", len(reader.File))
	return nil
}

// parseAndSaveFiles 解析并保存文件到数据库
func (s *AmbientCGDownloadService) parseAndSaveFiles(textureID uint, assetID, extractPath string) ([]models.File, error) {
	var files []models.File

	// 遍历目录
	err := filepath.Walk(extractPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 跳过非贴图文件
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			s.logDebug("跳过非贴图文件: %s", path)
			return nil
		}

		// 解析文件名
		fileName := info.Name()
		textureType := s.parseTextureType(fileName)

		// 生成 CDN 路径（只保存 assetID/fileName，不包含 textures/ 前缀）
		cdnPath := filepath.Join(assetID, fileName)
		cdnPath = strings.ReplaceAll(cdnPath, "\\", "/") // 使用正斜杠

		// 创建文件记录
		file := models.File{
			FileType:    "texture",
			RelatedID:   textureID,
			RelatedType: "Texture",
			LocalPath:   path,
			CDNPath:     cdnPath, // 只保存 assetID/fileName
			FileName:    fileName,
			FileSize:    info.Size(),
			Format:      strings.TrimPrefix(ext, "."),
			Status:      1, // 已下载
			TextureType: textureType,
		}

		if err := s.db.Create(&file).Error; err != nil {
			s.logError("保存文件记录失败: %v", err)
			return nil // 继续处理其他文件
		}

		files = append(files, file)
		s.logDebug("保存文件: %s (类型: %s)", fileName, textureType)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// parseTextureType 从文件名解析贴图类型
func (s *AmbientCGDownloadService) parseTextureType(fileName string) string {
	// 使用统一的贴图类型提取函数
	return models.ExtractTextureType(fileName)
}

// getExistingFiles 获取已存在的文件
func (s *AmbientCGDownloadService) getExistingFiles(textureID uint) ([]models.File, error) {
	var files []models.File
	err := s.db.Where("related_id = ? AND related_type = ? AND file_type = ?",
		textureID, "Texture", "texture").Find(&files).Error
	return files, err
}

// 日志方法
func (s *AmbientCGDownloadService) logInfo(format string, args ...interface{}) {
	s.logger.Infof("[AmbientCG Download] "+format, args...)
}

func (s *AmbientCGDownloadService) logWarn(format string, args ...interface{}) {
	s.logger.Warnf("[AmbientCG Download] "+format, args...)
}

func (s *AmbientCGDownloadService) logError(format string, err error, args ...interface{}) {
	allArgs := append([]interface{}{err}, args...)
	s.logger.Errorf("[AmbientCG Download] "+format, allArgs...)
}

func (s *AmbientCGDownloadService) logDebug(format string, args ...interface{}) {
	s.logger.Debugf("[AmbientCG Download] "+format, args...)
}
