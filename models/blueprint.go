package models

import (
	"time"

	"gorm.io/gorm"
)

// GenerateRequest 生成蓝图请求
type GenerateRequest struct {
	UserRequest string                    `json:"userRequest" binding:"required"`
	Metadata    map[string]NodeMetadata   `json:"metadata" binding:"required"`
}

// NodeMetadata 节点元数据
type NodeMetadata struct {
	Type       string         `json:"type"`
	Category   string         `json:"category"`
	Inputs     []PortInfo     `json:"inputs"`
	Outputs    []PortInfo     `json:"outputs"`
	Properties []PropertyInfo `json:"properties"`
}

// PortInfo 端口信息
type PortInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// PropertyInfo 属性信息
type PropertyInfo struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// GraphJSON 蓝图JSON结构
type GraphJSON struct {
	Nodes []NodeConfig `json:"nodes"`
	Links []LinkConfig `json:"links"`
}

// NodeConfig 节点配置
type NodeConfig struct {
	ID         int                    `json:"id"`
	Type       string                 `json:"type"`
	Pos        [2]float64             `json:"pos"`
	Properties map[string]interface{} `json:"properties"`
	Title      string                 `json:"title,omitempty"`
}

// LinkConfig 连接配置 [linkId, originId, originSlot, targetId, targetSlot, type]
type LinkConfig [6]interface{}

// GenerateResponse 生成响应
type GenerateResponse struct {
	GraphJSON GraphJSON        `json:"graphJson"`
	Metadata  GenerateMetadata `json:"metadata"`
}

// GenerateMetadata 生成元数据
type GenerateMetadata struct {
	Provider    string    `json:"provider"`
	ModelName   string    `json:"model"`
	TokensUsed  int       `json:"tokensUsed"`
	Duration    int64     `json:"duration"` // 毫秒
	GeneratedAt time.Time `json:"generatedAt"`
}

// BlueprintHistory 生成历史记录
type BlueprintHistory struct {
	gorm.Model
	UserRequest string `gorm:"type:text"`
	GraphJSON   string `gorm:"type:text"`
	Provider    string `gorm:"size:50"`
	ModelName   string `gorm:"size:100"`
	TokensUsed  int
	Duration    int64
	Success     bool
	ErrorMsg    string `gorm:"type:text"`
	UserIP      string `gorm:"size:50"`
}
