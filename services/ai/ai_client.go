package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go_wails_project_manager/config"
)

// AIClient AI客户端接口
type AIClient interface {
	Generate(ctx context.Context, prompt string) (string, int, error)
}

// UnifiedAIClient 统一的AI客户端（支持DeepSeek和OpenAI）
type UnifiedAIClient struct {
	apiKey      string
	model       string
	baseURL     string
	temperature float64
	maxTokens   int
	timeout     time.Duration
}

// NewAIClient 创建AI客户端
func NewAIClient() *UnifiedAIClient {
	cfg := config.AppAIConfig
	
	return &UnifiedAIClient{
		apiKey:      cfg.GetAPIKey(),
		model:       cfg.GetModel(),
		baseURL:     cfg.GetBaseURL(),
		temperature: cfg.GetTemperature(),
		maxTokens:   cfg.GetMaxTokens(),
		timeout:     time.Duration(cfg.Timeout) * time.Second,
	}
}

// ChatCompletionRequest OpenAI兼容的请求格式
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse OpenAI兼容的响应格式
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选择结构
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage token使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Generate 生成文本
func (c *UnifiedAIClient) Generate(ctx context.Context, prompt string) (string, int, error) {
	// 构建请求
	reqBody := ChatCompletionRequest{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: c.temperature,
		MaxTokens:   c.maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("API返回错误 [%d]: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("解析响应失败: %w", err)
	}

	// 提取内容
	if len(result.Choices) == 0 {
		return "", 0, fmt.Errorf("响应中没有内容")
	}

	content := result.Choices[0].Message.Content
	tokensUsed := result.Usage.TotalTokens

	return content, tokensUsed, nil
}

// GenerateWithRetry 带重试的生成
func (c *UnifiedAIClient) GenerateWithRetry(ctx context.Context, prompt string) (string, int, error) {
	cfg := config.AppAIConfig
	maxRetries := cfg.MaxRetries
	retryInterval := time.Duration(cfg.RetryInterval) * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		content, tokens, err := c.Generate(ctx, prompt)
		if err == nil {
			return content, tokens, nil
		}

		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return "", 0, fmt.Errorf("重试%d次后仍然失败: %w", maxRetries, lastErr)
}
