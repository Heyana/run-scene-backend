package audit

import (
	"go_wails_project_manager/config"
	"go_wails_project_manager/models"
	"time"

	"gorm.io/gorm"
)

// QueryService 查询服务
type QueryService struct {
	db     *gorm.DB
	config *config.AuditConfig
}

// NewQueryService 创建查询服务
func NewQueryService(db *gorm.DB, cfg *config.AuditConfig) *QueryService {
	return &QueryService{
		db:     db,
		config: cfg,
	}
}

// AuditFilter 审计日志过滤器
type AuditFilter struct {
	UserID     *uint
	Username   string
	UserIP     string
	Action     string
	Resource   string
	ResourceID *uint
	StartTime  *time.Time
	EndTime    *time.Time
	StatusCode *int
	Page       int
	PageSize   int
}

// AuditStatistics 审计统计信息
type AuditStatistics struct {
	TotalCount    int64              `json:"total_count"`
	ActionCount   map[string]int64   `json:"action_count"`
	ResourceCount map[string]int64   `json:"resource_count"`
	TopUsers      []UserActivity     `json:"top_users"`
	TopIPs        []IPActivity       `json:"top_ips"`
}

// UserActivity 用户活动统计
type UserActivity struct {
	Username string `json:"username"`
	Count    int64  `json:"count"`
}

// IPActivity IP活动统计
type IPActivity struct {
	IP    string `json:"ip"`
	Count int64  `json:"count"`
}

// List 查询审计日志列表
func (s *QueryService) List(filter AuditFilter) ([]models.AuditLog, int64, error) {
	query := s.db.Model(&models.AuditLog{})

	// 应用过滤条件
	query = s.applyFilters(query, filter)

	// 统计总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = s.config.DefaultPageSize
	}
	if filter.PageSize > s.config.MaxPageSize {
		filter.PageSize = s.config.MaxPageSize
	}

	offset := (filter.Page - 1) * filter.PageSize

	// 查询数据
	var logs []models.AuditLog
	if err := query.Order("created_at DESC").
		Offset(offset).
		Limit(filter.PageSize).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByID 根据ID查询审计日志
func (s *QueryService) GetByID(id uint) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := s.db.First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByUser 查询用户的审计日志
func (s *QueryService) GetByUser(userID uint, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	var logs []models.AuditLog
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetByResource 查询资源的审计日志
func (s *QueryService) GetByResource(resource string, resourceID uint, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	var logs []models.AuditLog
	if err := s.db.Where("resource = ? AND resource_id = ?", resource, resourceID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetByIP 查询IP的审计日志
func (s *QueryService) GetByIP(ip string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}

	var logs []models.AuditLog
	if err := s.db.Where("user_ip = ?", ip).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

// GetStatistics 获取统计信息
func (s *QueryService) GetStatistics(startTime, endTime time.Time) (*AuditStatistics, error) {
	stats := &AuditStatistics{
		ActionCount:   make(map[string]int64),
		ResourceCount: make(map[string]int64),
	}

	query := s.db.Model(&models.AuditLog{}).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime)

	// 总数
	if err := query.Count(&stats.TotalCount).Error; err != nil {
		return nil, err
	}

	// 按操作类型统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	if err := s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("action").
		Find(&actionStats).Error; err == nil {
		for _, stat := range actionStats {
			stats.ActionCount[stat.Action] = stat.Count
		}
	}

	// 按资源类型统计
	var resourceStats []struct {
		Resource string
		Count    int64
	}
	if err := s.db.Model(&models.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("resource").
		Find(&resourceStats).Error; err == nil {
		for _, stat := range resourceStats {
			stats.ResourceCount[stat.Resource] = stat.Count
		}
	}

	// Top 10 活跃用户
	if err := s.db.Model(&models.AuditLog{}).
		Select("username, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("username").
		Order("count DESC").
		Limit(10).
		Find(&stats.TopUsers).Error; err != nil {
		return nil, err
	}

	// Top 10 活跃IP
	if err := s.db.Model(&models.AuditLog{}).
		Select("user_ip as ip, COUNT(*) as count").
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Group("user_ip").
		Order("count DESC").
		Limit(10).
		Find(&stats.TopIPs).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// applyFilters 应用过滤条件
func (s *QueryService) applyFilters(query *gorm.DB, filter AuditFilter) *gorm.DB {
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.Username != "" {
		query = query.Where("username LIKE ?", "%"+filter.Username+"%")
	}

	if filter.UserIP != "" {
		query = query.Where("user_ip = ?", filter.UserIP)
	}

	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}

	if filter.ResourceID != nil {
		query = query.Where("resource_id = ?", *filter.ResourceID)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	if filter.StatusCode != nil {
		query = query.Where("status_code = ?", *filter.StatusCode)
	}

	return query
}
