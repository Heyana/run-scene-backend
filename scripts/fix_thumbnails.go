package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go_wails_project_manager/config"
	"go_wails_project_manager/database"
	"go_wails_project_manager/models/ai3d"
)

func main() {
	// 初始化配置
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		return
	}

	db, err := database.GetDB()
	if err != nil {
		fmt.Printf("获取数据库连接失败: %v\n", err)
		return
	}

	// 查询所有有thumbnail_url但没有thumbnail_path的任务
	var tasks []ai3d.Task
	err = db.Where("thumbnail_url IS NOT NULL AND thumbnail_url != ?", "").
		Where("(thumbnail_path IS NULL OR thumbnail_path = ?)", "").
		Where("status = ?", "DONE").
		Find(&tasks).Error

	if err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("没有需要修复的任务")
		return
	}

	fmt.Printf("找到 %d 个需要下载缩略图的任务\n\n", len(tasks))

	count := 0
	success := 0

	for _, task := range tasks {
		count++
		fmt.Printf("[%d/%d] 处理任务 ID=%d, Provider=%s\n", count, len(tasks), task.ID, task.Provider)
		
		if task.ThumbnailURL == nil || *task.ThumbnailURL == "" {
			fmt.Println("    ⊘ 跳过：没有缩略图URL")
			continue
		}

		thumbnailURL := *task.ThumbnailURL
		fmt.Printf("    缩略图URL: %s\n", thumbnailURL)

		// 下载缩略图
		thumbData, err := downloadFile(thumbnailURL)
		if err != nil {
			fmt.Printf("    ❌ 下载失败: %v\n", err)
			continue
		}
		fmt.Printf("    ✓ 下载成功，大小: %d bytes\n", len(thumbData))

		// 生成文件名
		var thumbFilename string
		if task.FileHash != nil && *task.FileHash != "" {
			thumbFilename = fmt.Sprintf("%s_%s_thumb.png", task.ProviderTaskID, (*task.FileHash)[:8])
		} else {
			thumbFilename = fmt.Sprintf("%s_thumb.png", task.ProviderTaskID)
		}

		// 确定保存路径
		var savePath string
		if task.NASPath != nil && *task.NASPath != "" {
			// 从nas_path提取目录
			nasDir := filepath.Dir(*task.NASPath)
			savePath = filepath.Join(nasDir, thumbFilename)
		} else {
			// 使用配置中的NAS路径
			var nasBase string
			if task.Provider == "meshy" {
				nasBase = config.AppConfig.Meshy.NASPath
			} else if task.Provider == "hunyuan" {
				nasBase = config.AppConfig.Hunyuan.NASPath
			}
			savePath = filepath.Join(nasBase, thumbFilename)
		}

		// 确保目录存在
		dir := filepath.Dir(savePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("    ❌ 创建目录失败: %v\n", err)
			continue
		}

		// 保存文件
		if err := os.WriteFile(savePath, thumbData, 0644); err != nil {
			fmt.Printf("    ❌ 保存文件失败: %v\n", err)
			continue
		}
		fmt.Printf("    ✓ 已保存到: %s\n", savePath)

		// 更新数据库
		err = db.Model(&task).Update("thumbnail_path", savePath).Error
		if err != nil {
			fmt.Printf("    ❌ 更新数据库失败: %v\n", err)
			continue
		}
		fmt.Printf("    ✓ 数据库已更新\n\n")

		success++
	}

	fmt.Printf("========================================\n")
	fmt.Printf("处理完成！\n")
	fmt.Printf("总计: %d 个任务\n", count)
	fmt.Printf("成功: %d 个\n", success)
	fmt.Printf("失败: %d 个\n", count-success)
	fmt.Printf("========================================\n")
}

func downloadFile(url string) ([]byte, error) {
	// 跳过已经是本地路径的
	if strings.HasPrefix(url, "\\\\") || strings.HasPrefix(url, "/") || (strings.Contains(url, ":/") && !strings.HasPrefix(url, "http")) {
		return nil, fmt.Errorf("不是有效的HTTP URL")
	}

	// 创建带代理的HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(mustParseURL("http://127.0.0.1:7890")),
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

