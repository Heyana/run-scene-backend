package texture

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/HugoSmits86/nativewebp"
	"github.com/jlaffaye/ftp"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
	"gorm.io/gorm"
)

// DownloadService 下载服务
type DownloadService struct {
	db                  *gorm.DB
	localStorageEnabled bool
	storageDir          string
	logger              *logrus.Logger
	httpClient          *http.Client
	nasEnabled          bool
	nasPath             string
	webdavClient        *gowebdav.Client
	webdavEnabled       bool
	ftpEnabled          bool
	ftpHost             string
	ftpPort             int
	ftpUsername         string
	ftpPassword         string
	ftpBasePath         string
}

// NewDownloadService 创建下载服务
func NewDownloadService(db *gorm.DB, logger *logrus.Logger) *DownloadService {
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
			logger.Infof("下载服务已启用代理: %s", config.AppConfig.Texture.ProxyURL)
		} else {
			logger.Warnf("代理 URL 解析失败: %v", err)
		}
	}

	// 本地存储配置
	localStorageEnabled := config.AppConfig.Texture.LocalStorageEnabled
	if localStorageEnabled {
		logger.Infof("本地存储已启用: %s", config.AppConfig.Texture.StorageDir)
	} else {
		logger.Infof("本地存储已禁用")
	}

	// NAS SMB 配置
	nasEnabled := false
	nasPath := ""
	if config.AppConfig.Texture.NASEnabled && config.AppConfig.Texture.NASPath != "" {
		nasEnabled = true
		nasPath = config.AppConfig.Texture.NASPath
		logger.Infof("NAS SMB 存储已启用: %s", nasPath)
	}

	// 创建 WebDAV 客户端
	var webdavClient *gowebdav.Client
	webdavEnabled := false
	if config.AppConfig.Texture.WebDAVEnabled && config.AppConfig.Texture.WebDAVURL != "" {
		webdavClient = gowebdav.NewClient(
			config.AppConfig.Texture.WebDAVURL,
			config.AppConfig.Texture.WebDAVUsername,
			config.AppConfig.Texture.WebDAVPassword,
		)
		webdavEnabled = true
		logger.Infof("WebDAV 客户端已初始化: %s", config.AppConfig.Texture.WebDAVURL)
	}

	// FTP 配置
	ftpEnabled := false
	if config.AppConfig.Texture.FTPEnabled && config.AppConfig.Texture.FTPHost != "" {
		ftpEnabled = true
		logger.Infof("FTP 客户端已配置: %s:%d", config.AppConfig.Texture.FTPHost, config.AppConfig.Texture.FTPPort)
	}

	return &DownloadService{
		db:                  db,
		localStorageEnabled: localStorageEnabled,
		storageDir:          config.AppConfig.Texture.StorageDir,
		logger:              logger,
		httpClient:          client,
		nasEnabled:          nasEnabled,
		nasPath:             nasPath,
		webdavClient:        webdavClient,
		webdavEnabled:       webdavEnabled,
		ftpEnabled:          ftpEnabled,
		ftpHost:             config.AppConfig.Texture.FTPHost,
		ftpPort:             config.AppConfig.Texture.FTPPort,
		ftpUsername:         config.AppConfig.Texture.FTPUsername,
		ftpPassword:         config.AppConfig.Texture.FTPPassword,
		ftpBasePath:         config.AppConfig.Texture.FTPBasePath,
	}
}

// DownloadThumbnail 下载缩略图
func (s *DownloadService) DownloadThumbnail(textureID uint, assetID string, thumbnailURL string) (*models.File, error) {
	startTime := time.Now()
	s.logger.Infof("开始下载缩略图: %s", assetID)

	// 下载图片
	downloadStart := time.Now()
	imageData, err := s.downloadImage(thumbnailURL)
	if err != nil {
		return nil, fmt.Errorf("下载缩略图失败: %w", err)
	}
	downloadDuration := time.Since(downloadStart)
	s.logger.Infof("下载完成: %.2f KB, 耗时: %v", float64(len(imageData))/1024, downloadDuration)

	// 直接保存原图，不转码
	fileName := "thumbnail.png"
	file, err := s.saveFile(textureID, "Texture", "thumbnail", imageData, fileName, assetID)
	if err != nil {
		return nil, fmt.Errorf("保存缩略图失败: %w", err)
	}

	totalDuration := time.Since(startTime)
	s.logger.Infof("缩略图保存成功: %s, 总耗时: %v", file.LocalPath, totalDuration)
	return file, nil
}

// DownloadAndConvert 下载并转码贴图
func (s *DownloadService) DownloadAndConvert(textureID uint, assetID string, files map[string]interface{}) error {
	s.logger.Infof("开始下载贴图文件: %s", assetID)

	// 遍历贴图类型（Diffuse, Normal, Roughness等）
	for mapType, resolutions := range files {
		if mapType == "blend" || mapType == "gltf" {
			continue // 跳过blend和gltf文件
		}

		resMap, ok := resolutions.(map[string]interface{})
		if !ok {
			continue
		}

		// 选择最优分辨率
		url, resolution := s.selectBestResolution(resMap)
		if url == "" {
			s.logger.Warnf("未找到合适的分辨率: %s/%s", assetID, mapType)
			continue
		}

		textureStart := time.Now()
		s.logger.Infof("下载贴图: %s/%s [%s]", assetID, mapType, resolution)

		// 下载图片
		downloadStart := time.Now()
		imageData, err := s.downloadImage(url)
		if err != nil {
			s.logger.Errorf("下载失败: %s/%s - %v", assetID, mapType, err)
			continue
		}
		downloadDuration := time.Since(downloadStart)
		s.logger.Infof("下载完成: %.2f MB, 耗时: %v", float64(len(imageData))/1024/1024, downloadDuration)

		// 直接保存原图，不转码
		// 根据 URL 判断文件格式
		fileExt := "png"
		if strings.Contains(url, ".jpg") {
			fileExt = "jpg"
		} else if strings.Contains(url, ".exr") {
			fileExt = "exr"
		}

		// 保存文件
		fileName := fmt.Sprintf("%s_%s.%s", mapType, resolution, fileExt)
		file, err := s.saveFile(textureID, "Texture", "texture", imageData, fileName, assetID)
		if err != nil {
			s.logger.Errorf("保存失败: %s/%s - %v", assetID, mapType, err)
			continue
		}

		totalDuration := time.Since(textureStart)
		s.logger.Infof("贴图保存成功: %s, 总耗时: %v", file.LocalPath, totalDuration)

		// 创建 TextureFile 关联
		textureFile := models.TextureFile{
			TextureID:  textureID,
			FileID:     file.ID,
			MapType:    mapType,
			Resolution: "1k",
		}
		if err := s.db.Create(&textureFile).Error; err != nil {
			s.logger.Errorf("创建关联失败: %v", err)
		}

		s.logger.Infof("贴图保存成功: %s", file.LocalPath)
	}

	return nil
}

// selectBestResolution 选择最优分辨率（>1K的最小）
func (s *DownloadService) selectBestResolution(resMap map[string]interface{}) (string, string) {
	// 优先级：2k > 4k > 8k > 1k
	priorities := []string{"2k", "4k", "8k", "16k", "1k"}

	for _, res := range priorities {
		if resData, ok := resMap[res].(map[string]interface{}); ok {
			// 优先选择 jpg 格式
			if jpgData, ok := resData["jpg"].(map[string]interface{}); ok {
				if url, ok := jpgData["url"].(string); ok {
					return url, res
				}
			}
			// 其次选择 png
			if pngData, ok := resData["png"].(map[string]interface{}); ok {
				if url, ok := pngData["url"].(string); ok {
					return url, res
				}
			}
		}
	}

	return "", ""
}

// downloadImage 下载图片
func (s *DownloadService) downloadImage(url string) ([]byte, error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// convertToWebP 转码为 WebP
func (s *DownloadService) convertToWebP(imageData []byte, targetSize int) ([]byte, error) {
	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, err
	}

	// 调整大小
	resized := resize.Resize(uint(targetSize), uint(targetSize), img, resize.Lanczos3)

	// 编码为 WebP (无损)
	var buf bytes.Buffer
	options := &nativewebp.Options{
		UseExtendedFormat: false, // 不需要元数据支持
	}
	if err := nativewebp.Encode(&buf, resized, options); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// saveFile 保存文件到本地并创建 File 记录
func (s *DownloadService) saveFile(relatedID uint, relatedType string, fileType string, data []byte, fileName string, assetID string) (*models.File, error) {
	// 检查文件是否已存在（根据 related_id, file_name 和 file_type）
	var existingFile models.File
	err := s.db.Where("related_id = ? AND related_type = ? AND file_name = ? AND file_type = ?", 
		relatedID, relatedType, fileName, fileType).First(&existingFile).Error
	
	if err == nil {
		// 文件已存在，跳过
		s.logger.Infof("文件已存在，跳过: %s/%s", assetID, fileName)
		return &existingFile, nil
	}
	
	// 本地存储路径（用于数据库记录）
	var filePath string
	
	// 如果启用本地存储，保存到本地
	if s.localStorageEnabled {
		// 创建目录
		dirPath := filepath.Join(s.storageDir, assetID)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, err
		}

		// 保存文件
		filePath = filepath.Join(dirPath, fileName)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return nil, err
		}
		s.logger.Debugf("本地保存成功: %s", filePath)
	} else {
		// 不保存本地，但仍需要路径用于数据库记录
		filePath = filepath.Join(s.storageDir, assetID, fileName)
		s.logger.Debugf("跳过本地保存（已禁用）")
	}

	// 如果启用 NAS SMB，直接保存到网络共享
	if s.nasEnabled && s.nasPath != "" {
		nasFilePath := filepath.Join(s.nasPath, assetID, fileName)
		nasDirPath := filepath.Join(s.nasPath, assetID)
		
		// 创建目录
		if err := os.MkdirAll(nasDirPath, 0755); err != nil {
			s.logger.Warnf("NAS 创建目录失败: %s - %v", nasDirPath, err)
		} else {
			// 保存文件
			if err := os.WriteFile(nasFilePath, data, 0644); err != nil {
				s.logger.Warnf("NAS 保存失败: %s - %v", nasFilePath, err)
			} else {
				s.logger.Infof("NAS 保存成功: %s", nasFilePath)
			}
		}
	}

	// 如果启用 FTP，上传到 NAS（备选方案）
	if s.ftpEnabled {
		if err := s.uploadToFTP(assetID, fileName, data); err != nil {
			s.logger.Warnf("FTP 上传失败: %s/%s - %v", assetID, fileName, err)
		} else {
			s.logger.Infof("FTP 上传成功: %s/%s", assetID, fileName)
		}
	}

	// 如果启用 WebDAV，上传到 NAS（备选方案）
	if s.webdavEnabled && s.webdavClient != nil {
		webdavPath := filepath.Join(assetID, fileName)
		// 使用正斜杠作为 WebDAV 路径分隔符
		webdavPath = strings.ReplaceAll(webdavPath, "\\", "/")
		
		// 创建目录
		webdavDir := assetID
		if err := s.webdavClient.MkdirAll(webdavDir, 0755); err != nil {
			s.logger.Warnf("WebDAV 创建目录失败: %v", err)
		}
		
		// 上传文件
		if err := s.webdavClient.Write(webdavPath, data, 0644); err != nil {
			s.logger.Warnf("WebDAV 上传失败: %s - %v", webdavPath, err)
		} else {
			s.logger.Infof("WebDAV 上传成功: %s", webdavPath)
		}
	}

	// 计算 MD5
	hash := md5.Sum(data)
	md5Str := hex.EncodeToString(hash[:])

	// 获取图片尺寸
	width, height := 0, 0
	format := "unknown"
	if img, imgFormat, err := image.Decode(bytes.NewReader(data)); err == nil {
		bounds := img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
		format = imgFormat
	}

	// 生成相对路径（只保存 assetID/fileName，不包含 base_url）
	relativePath := filepath.Join(assetID, fileName)
	// 使用正斜杠作为路径分隔符
	cdnPath := strings.ReplaceAll(relativePath, "\\", "/")

	// 创建 File 记录
	file := models.File{
		FileType:    fileType,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		LocalPath:   filePath,
		CDNPath:     cdnPath, // 只保存相对路径
		FileName:    fileName,
		FileSize:    int64(len(data)),
		Width:       width,
		Height:      height,
		Format:      format,
		MD5:         md5Str,
		Status:      1, // 已下载
	}

	if err := s.db.Create(&file).Error; err != nil {
		return nil, err
	}

	return &file, nil
}

// logDownloadProgress 记录下载进度
func (s *DownloadService) logDownloadProgress(assetID string, step string, details string) {
	s.logger.Infof("[%s] %s: %s", assetID, step, details)
}

// uploadToFTP 上传文件到 FTP 服务器
func (s *DownloadService) uploadToFTP(assetID string, fileName string, data []byte) error {
	// 连接 FTP 服务器
	addr := fmt.Sprintf("%s:%d", s.ftpHost, s.ftpPort)
	conn, err := ftp.Dial(addr, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return fmt.Errorf("连接 FTP 服务器失败: %w", err)
	}
	defer conn.Quit()

	// 登录
	if err := conn.Login(s.ftpUsername, s.ftpPassword); err != nil {
		return fmt.Errorf("FTP 登录失败: %w", err)
	}

	// 切换到基础路径
	if s.ftpBasePath != "" {
		if err := conn.ChangeDir(s.ftpBasePath); err != nil {
			return fmt.Errorf("切换到基础路径失败: %w", err)
		}
	}

	// 创建资源目录
	if err := conn.MakeDir(assetID); err != nil {
		// 目录可能已存在，忽略错误
		s.logger.Debugf("FTP 创建目录: %s (可能已存在)", assetID)
	}

	// 切换到资源目录
	if err := conn.ChangeDir(assetID); err != nil {
		return fmt.Errorf("切换到资源目录失败: %w", err)
	}

	// 上传文件
	reader := bytes.NewReader(data)
	if err := conn.Stor(fileName, reader); err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}
