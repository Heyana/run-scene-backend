// Package services 提供企业名片系统的各种服务实现
package services

import (
	"context"
	"errors"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/logger"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CDNService 定义CDN服务接口
type CDNService interface {
	// Upload 上传文件到CDN
	Upload(ctx context.Context, data []byte, path string, mimeType string) (string, error)

	// Delete 从CDN删除文件
	Delete(ctx context.Context, path string) error

	// GetURL 获取文件的CDN URL
	GetURL(path string) string
}

// LocalCDNService 本地CDN实现(用于开发环境)
type LocalCDNService struct {
	basePath string // 本地存储路径
	baseURL  string // 访问基础URL
}

// NewLocalCDNService 创建本地CDN服务实例
func NewLocalCDNService() *LocalCDNService {
	// 确保static/cdn目录存在
	cdnDir := filepath.Join("static", "cdn")
	if err := os.MkdirAll(cdnDir, 0755); err != nil {
		logger.Log.Errorf("创建CDN目录失败: %v", err)
	}

	return &LocalCDNService{
		basePath: cdnDir,
		baseURL:  config.GetCDNBaseURL(),
	}
}

// Upload 将文件保存到本地目录
func (l *LocalCDNService) Upload(ctx context.Context, data []byte, path string, mimeType string) (string, error) {
	// 确保目录存在
	dir := filepath.Dir(filepath.Join(l.basePath, path))
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Log.Errorf("创建CDN目录失败: %v, 路径: %s", err, dir)
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 写入文件
	filePath := filepath.Join(l.basePath, path)
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		logger.Log.Errorf("写入CDN文件失败: %v, 路径: %s", err, filePath)
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 标准化路径（使用正斜杠）
	path = strings.ReplaceAll(path, "\\", "/")

	logger.Log.Infof("文件上传成功: 类型=%s, 大小=%d字节, 相对路径=%s", mimeType, len(data), path)

	// 返回相对路径（不包含base URL）
	return path, nil
}

// Delete 从本地删除文件
func (l *LocalCDNService) Delete(ctx context.Context, path string) error {
	// 检查是否是完整的URL路径（以http:或https:开头）
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// 从URL中提取相对路径部分
		baseURL := l.baseURL
		if !strings.HasSuffix(baseURL, "/") {
			baseURL = baseURL + "/"
		}

		// 从URL中移除baseURL前缀，获取相对路径
		if strings.HasPrefix(path, baseURL) {
			path = strings.TrimPrefix(path, baseURL)
		} else {
			// 如果不是当前CDN的URL，尝试提取路径部分
			urlParts := strings.SplitN(path, "/cdn/", 2)
			if len(urlParts) == 2 {
				path = urlParts[1]
			} else {
				return fmt.Errorf("无法从URL中提取有效路径: %s", path)
			}
		}
	}

	// 现在path应该是相对于CDN根目录的路径
	filePath := filepath.Join(l.basePath, path)
	logger.Log.Debugf("删除CDN文件: %s", filePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，视为删除成功
		logger.Log.Warnf("要删除的文件不存在: %s", filePath)
		return nil
	}

	return os.Remove(filePath)
}

// GetURL 获取文件的URL
func (l *LocalCDNService) GetURL(path string) string {
	// 确保路径使用正斜杠
	path = strings.ReplaceAll(path, "\\", "/")
	// 确保baseURL以/结尾
	baseURL := l.baseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL = baseURL + "/"
	}
	// 确保path不以/开头
	path = strings.TrimPrefix(path, "/")

	return baseURL + path
}

// CloudCDNService 云CDN实现(用于生产环境)
// 这里可以根据实际使用的云服务商实现(如阿里云OSS、腾讯云COS、七牛云等)
type CloudCDNService struct {
	// 根据实际使用的云服务添加必要的字段
}

// NewCDNService 创建CDN服务实例(工厂方法)
func NewCDNService() CDNService {
	// 根据环境配置选择实现
	if config.IsDev() {
		return NewLocalCDNService()
	}

	// TODO: 返回云CDN服务实例
	// 目前仍返回本地服务作为临时方案
	return NewLocalCDNService()
}

// GenerateFilePath 生成文件存储路径
func GenerateFilePath(originalFilename string, fileType string) string {
	// 按年月组织目录
	now := time.Now()
	yearMonth := now.Format("2006-01")

	// 生成唯一文件名(时间戳+随机字符)
	timestamp := now.UnixNano() / 1000000 // 毫秒时间戳
	randomStr := fmt.Sprintf("%d", timestamp)

	// 获取文件扩展名
	ext := filepath.Ext(originalFilename)
	if ext == "" && fileType != "" {
		// 如果原文件名没有扩展名，尝试根据fileType添加
		switch {
		case strings.Contains(fileType, "jpeg"), strings.Contains(fileType, "jpg"):
			ext = ".jpg"
		case strings.Contains(fileType, "png"):
			ext = ".png"
		case strings.Contains(fileType, "webp"):
			ext = ".webp"
		case strings.Contains(fileType, "gif"):
			ext = ".gif"
		}
	}

	// 构建路径: 年月/时间戳.扩展名
	return fmt.Sprintf("%s/%s%s", yearMonth, randomStr, ext)
}

// UploadFile 上传MultipartFile文件
func UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, enterpriseID uint, cdnService CDNService) (string, error) {
	// 读取文件数据
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// 检查文件大小
	if len(data) == 0 {
		return "", errors.New("文件内容为空")
	}

	// 生成存储路径
	path := GenerateFilePath(header.Filename, header.Header.Get("Content-Type"))

	// 上传到CDN
	return cdnService.Upload(ctx, data, path, header.Header.Get("Content-Type"))
}
