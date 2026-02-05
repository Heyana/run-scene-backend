package blueprint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"go_wails_project_manager/services/ai"
)

// BlueprintService 蓝图生成服务
type BlueprintService struct {
	db       *gorm.DB
	aiClient *ai.UnifiedAIClient
}

// NewBlueprintService 创建蓝图服务
func NewBlueprintService(db *gorm.DB) *BlueprintService {
	return &BlueprintService{
		db:       db,
		aiClient: ai.NewAIClient(),
	}
}

// Generate 生成蓝图
func (s *BlueprintService) Generate(ctx context.Context, req *models.GenerateRequest, userIP string) (*models.GenerateResponse, error) {
	startTime := time.Now()

	// 1. 构建Prompt
	prompt := ai.BuildPrompt(req.UserRequest, req.Metadata)

	// 2. 调用AI
	content, tokensUsed, err := s.aiClient.GenerateWithRetry(ctx, prompt)
	if err != nil {
		// 保存失败记录
		s.saveHistory(req.UserRequest, "", false, err.Error(), userIP, 0, 0)
		return nil, fmt.Errorf("AI生成失败: %w", err)
	}

	// 3. 提取JSON
	graphJSON, err := extractJSON(content)
	if err != nil {
		s.saveHistory(req.UserRequest, content, false, err.Error(), userIP, tokensUsed, 0)
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	// 4. 计算耗时
	duration := time.Since(startTime).Milliseconds()

	// 5. 保存成功记录
	jsonStr, _ := json.Marshal(graphJSON)
	s.saveHistory(req.UserRequest, string(jsonStr), true, "", userIP, tokensUsed, duration)

	// 6. 构建响应
	response := &models.GenerateResponse{
		GraphJSON: *graphJSON,
		Metadata: models.GenerateMetadata{
			Provider:    config.AppAIConfig.Provider,
			ModelName:   config.AppAIConfig.GetModel(),
			TokensUsed:  tokensUsed,
			Duration:    duration,
			GeneratedAt: time.Now(),
		},
	}

	return response, nil
}

// extractJSON 从AI响应中提取JSON
func extractJSON(text string) (*models.GraphJSON, error) {
	// 尝试提取 ```json...``` 代码块
	if strings.Contains(text, "```json") {
		start := strings.Index(text, "```json") + 7
		end := strings.Index(text[start:], "```")
		if end != -1 {
			text = text[start : start+end]
		}
	} else if strings.Contains(text, "```") {
		// 尝试提取 ```...``` 代码块
		start := strings.Index(text, "```") + 3
		end := strings.Index(text[start:], "```")
		if end != -1 {
			text = text[start : start+end]
		}
	}

	// 查找第一个 { 和最后一个 }
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("未找到有效的JSON")
	}

	jsonStr := strings.TrimSpace(text[start : end+1])

	// 解析JSON
	var graphJSON models.GraphJSON
	if err := json.Unmarshal([]byte(jsonStr), &graphJSON); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	return &graphJSON, nil
}

// saveHistory 保存生成历史
func (s *BlueprintService) saveHistory(
	userRequest string,
	graphJSON string,
	success bool,
	errorMsg string,
	userIP string,
	tokensUsed int,
	duration int64,
) {
	history := &models.BlueprintHistory{
		UserRequest: userRequest,
		GraphJSON:   graphJSON,
		Provider:    config.AppAIConfig.Provider,
		ModelName:   config.AppAIConfig.GetModel(),
		TokensUsed:  tokensUsed,
		Duration:    duration,
		Success:     success,
		ErrorMsg:    errorMsg,
		UserIP:      userIP,
	}

	s.db.Create(history)
}

// GetHistory 获取生成历史
func (s *BlueprintService) GetHistory(page, pageSize int) ([]models.BlueprintHistory, int64, error) {
	var histories []models.BlueprintHistory
	var total int64

	offset := (page - 1) * pageSize

	if err := s.db.Model(&models.BlueprintHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}
