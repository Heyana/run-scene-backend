package models

import (
	"time"
)

// AuditLog 审计日志表
type AuditLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       *uint     `gorm:"index" json:"user_id,omitempty"`                    // 用户ID（可为空，匿名操作）
	Username     string    `gorm:"size:100;index" json:"username"`                    // 用户名
	UserIP       string    `gorm:"size:50;index" json:"user_ip"`                      // 客户端IP
	Action       string    `gorm:"size:50;index" json:"action"`                       // 操作类型
	Resource     string    `gorm:"size:100;index" json:"resource"`                    // 资源类型
	ResourceID   *uint     `gorm:"index" json:"resource_id,omitempty"`                // 资源ID
	Method       string    `gorm:"size:10" json:"method"`                             // HTTP方法
	Path         string    `gorm:"size:512" json:"path"`                              // 请求路径
	StatusCode   int       `gorm:"index" json:"status_code"`                          // 响应状态码
	Duration     int64     `json:"duration"`                                          // 请求耗时（毫秒）
	RequestBody  string    `gorm:"type:text" json:"request_body,omitempty"`           // 请求体（敏感信息已脱敏）
	ResponseBody string    `gorm:"type:text" json:"response_body,omitempty"`          // 响应体（可选）
	ErrorMsg     string    `gorm:"type:text" json:"error_msg,omitempty"`              // 错误信息
	UserAgent    string    `gorm:"size:512" json:"user_agent,omitempty"`              // 用户代理
	CreatedAt    time.Time `gorm:"index" json:"created_at"`                           // 创建时间
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// 操作类型常量
const (
	ActionLogin    = "login"     // 登录
	ActionLogout   = "logout"    // 登出
	ActionCreate   = "create"    // 创建
	ActionUpdate   = "update"    // 更新
	ActionDelete   = "delete"    // 删除
	ActionView     = "view"      // 查看
	ActionDownload = "download"  // 下载
	ActionUpload   = "upload"    // 上传
	ActionMove     = "move"      // 移动
	ActionRename   = "rename"    // 重命名
	ActionShare    = "share"     // 分享
	ActionExport   = "export"    // 导出
	ActionImport   = "import"    // 导入
)

// 资源类型常量
const (
	ResourceDocument = "document" // 文档
	ResourceFolder   = "folder"   // 文件夹
	ResourceUser     = "user"     // 用户
	ResourceProject  = "project"  // 项目
	ResourceModel    = "model"    // 模型
	ResourceTexture  = "texture"  // 材质
)
