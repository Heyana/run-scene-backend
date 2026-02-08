package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// ScreenshotOptions 截图选项
type ScreenshotOptions struct {
	Width   int64         // 视口宽度
	Height  int64         // 视口高度
	Quality int           // 图片质量 (1-100)
	Delay   time.Duration // 等待页面加载时间
	Timeout time.Duration // 超时时间
}

// DefaultScreenshotOptions 默认截图选项
var DefaultScreenshotOptions = ScreenshotOptions{
	Width:   1920,
	Height:  1080,
	Quality: 90,
	Delay:   60 * time.Second, // 等待1分钟，确保页面完全加载
	Timeout: 90 * time.Second, // 超时时间90秒
}

// GenerateScreenshot 生成网页截图
func GenerateScreenshot(url string, outputPath string, opts *ScreenshotOptions) error {
	if opts == nil {
		opts = &DefaultScreenshotOptions
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 创建 context，设置超时
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// 创建 chromedp context，禁用 CSP 和其他安全限制
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-web-security", true),           // 禁用 web 安全
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"), // 禁用站点隔离
		chromedp.Flag("disable-blink-features", "AutomationControlled"),      // 隐藏自动化特征
	)
	
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// 截图任务
	var buf []byte
	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(opts.Width, opts.Height),
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery), // 等待body加载
		chromedp.Sleep(opts.Delay),                   // 等待页面完全加载和渲染
		chromedp.FullScreenshot(&buf, opts.Quality),
	}

	if err := chromedp.Run(browserCtx, tasks); err != nil {
		return fmt.Errorf("截图失败: %v", err)
	}

	// 保存文件
	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("保存截图失败: %v", err)
	}

	return nil
}

// GenerateThumbnail 生成缩略图（固定尺寸）
func GenerateThumbnail(url string, outputPath string, width, height int64) error {
	opts := &ScreenshotOptions{
		Width:   width,
		Height:  height,
		Quality: 85,
		Delay:   60 * time.Second, // 等待1分钟
		Timeout: 90 * time.Second, // 超时90秒
	}

	// 创建 context
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// 创建 chromedp context，禁用 CSP 和其他安全限制
	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("disable-web-security", true),           // 禁用 web 安全
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"), // 禁用站点隔离
		chromedp.Flag("disable-blink-features", "AutomationControlled"),      // 隐藏自动化特征
	)
	
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 截图任务（使用 CaptureScreenshot 而不是 FullScreenshot，只截取视口）
	var buf []byte
	tasks := chromedp.Tasks{
		chromedp.EmulateViewport(width, height),
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery), // 等待body加载
		chromedp.Sleep(opts.Delay),                   // 额外等待1分钟确保资源加载
		chromedp.CaptureScreenshot(&buf),
	}

	if err := chromedp.Run(browserCtx, tasks); err != nil {
		return fmt.Errorf("生成缩略图失败: %v", err)
	}

	// 保存文件
	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("保存缩略图失败: %v", err)
	}

	return nil
}

// GeneratePreviewScreenshots 生成多种尺寸的预览图
func GeneratePreviewScreenshots(url string, outputDir string) (map[string]string, error) {
	screenshots := make(map[string]string)

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 定义不同尺寸
	sizes := map[string]struct{ width, height int64 }{
		"thumbnail": {400, 300},   // 缩略图
		"medium":    {800, 600},   // 中等尺寸
		"large":     {1920, 1080}, // 大尺寸
	}

	// 生成各种尺寸的截图
	for name, size := range sizes {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("preview_%s.png", name))
		if err := GenerateThumbnail(url, outputPath, size.width, size.height); err != nil {
			// 记录错误但继续生成其他尺寸
			fmt.Printf("生成 %s 截图失败: %v\n", name, err)
			continue
		}
		screenshots[name] = outputPath
	}

	if len(screenshots) == 0 {
		return nil, fmt.Errorf("所有截图生成失败")
	}

	return screenshots, nil
}
