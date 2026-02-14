package requirement_service

import (
	"errors"
	"go_wails_project_manager/database"
	"go_wails_project_manager/models"
	"go_wails_project_manager/models/requirement"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ProjectService 项目服务
type ProjectService struct {
	db *gorm.DB
}

// NewProjectService 创建项目服务
func NewProjectService() *ProjectService {
	db, _ := database.GetDB()
	return &ProjectService{db: db}
}

// CreateProject 创建项目
func (ps *ProjectService) CreateProject(userID, companyID uint, name, key, description string, startDate, endDate *time.Time) (*requirement.Project, error) {
	// 检查用户是否是公司成员
	var member requirement.CompanyMember
	if err := ps.db.Where("company_id = ? AND user_id = ?", companyID, userID).
		First(&member).Error; err != nil {
		return nil, errors.New("您不是该公司成员")
	}

	// 检查项目Key是否已存在
	var existingProject requirement.Project
	if err := ps.db.Where("company_id = ? AND key = ?", companyID, strings.ToUpper(key)).
		First(&existingProject).Error; err == nil {
		return nil, errors.New("项目标识已存在")
	}

	project := requirement.Project{
		CompanyID:   companyID,
		Name:        name,
		Key:         strings.ToUpper(key),
		Description: description,
		OwnerID:     userID,
		Status:      "active",
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := ps.db.Create(&project).Error; err != nil {
		return nil, err
	}

	// 自动添加创建人为项目管理员
	projectMember := requirement.ProjectMember{
		ProjectID: project.ID,
		UserID:    userID,
		Role:      "project_admin",
		JoinedAt:  time.Now(),
	}
	ps.db.Create(&projectMember)

	return &project, nil
}

// ListUserProjects 获取用户的项目列表
func (ps *ProjectService) ListUserProjects(userID, companyID uint, page, pageSize int, status string) ([]requirement.Project, int64, error) {
	var projects []requirement.Project
	var total int64

	query := ps.db.Model(&requirement.Project{}).
		Joins("JOIN requirement_project_members ON requirement_projects.id = requirement_project_members.project_id").
		Where("requirement_project_members.user_id = ?", userID)

	if companyID > 0 {
		query = query.Where("requirement_projects.company_id = ?", companyID)
	}

	if status != "" {
		query = query.Where("requirement_projects.status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Preload("Company").
		Preload("Owner").
		Preload("Members.User").
		Find(&projects).Error

	return projects, total, err
}

// GetProjectDetail 获取项目详情
func (ps *ProjectService) GetProjectDetail(projectID uint) (*requirement.Project, error) {
	var project requirement.Project
	err := ps.db.Preload("Company").
		Preload("Owner").
		Preload("Members.User").
		Preload("MissionLists").
		First(&project, projectID).Error
	return &project, err
}

// UpdateProject 更新项目
func (ps *ProjectService) UpdateProject(projectID uint, name, description string, ownerID uint, status string, startDate, endDate *time.Time) (*requirement.Project, error) {
	var project requirement.Project
	if err := ps.db.First(&project, projectID).Error; err != nil {
		return nil, err
	}

	if name != "" {
		project.Name = name
	}
	project.Description = description
	if ownerID > 0 {
		project.OwnerID = ownerID
	}
	if status != "" {
		project.Status = status
	}
	project.StartDate = startDate
	project.EndDate = endDate

	if err := ps.db.Save(&project).Error; err != nil {
		return nil, err
	}

	return &project, nil
}

// AddMember 添加项目成员
func (ps *ProjectService) AddMember(projectID, userID uint, role string) (*requirement.ProjectMember, error) {
	// 检查用户是否存在
	var user models.User
	if err := ps.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户是否是公司成员
	var project requirement.Project
	ps.db.First(&project, projectID)

	var companyMember requirement.CompanyMember
	if err := ps.db.Where("company_id = ? AND user_id = ?", project.CompanyID, userID).
		First(&companyMember).Error; err != nil {
		return nil, errors.New("用户不是公司成员")
	}

	// 检查是否已经是项目成员
	var existingMember requirement.ProjectMember
	if err := ps.db.Where("project_id = ? AND user_id = ?", projectID, userID).
		First(&existingMember).Error; err == nil {
		return nil, errors.New("用户已经是项目成员")
	}

	member := requirement.ProjectMember{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
	}

	if err := ps.db.Create(&member).Error; err != nil {
		return nil, err
	}

	// 加载关联数据
	ps.db.Preload("User").First(&member, member.ID)

	return &member, nil
}

// RemoveMember 移除项目成员
func (ps *ProjectService) RemoveMember(projectID, userID uint) error {
	// 不能移除项目所有者
	var project requirement.Project
	if err := ps.db.First(&project, projectID).Error; err != nil {
		return err
	}

	if project.OwnerID == userID {
		return errors.New("不能移除项目所有者")
	}

	return ps.db.Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&requirement.ProjectMember{}).Error
}

// GetMembers 获取项目成员列表
func (ps *ProjectService) GetMembers(projectID uint) ([]requirement.ProjectMember, error) {
	var members []requirement.ProjectMember
	err := ps.db.Where("project_id = ?", projectID).
		Preload("User").
		Find(&members).Error
	return members, err
}

// HasAccess 检查用户是否有访问权限
func (ps *ProjectService) HasAccess(projectID, userID uint) bool {
	var count int64
	ps.db.Model(&requirement.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count)
	return count > 0
}

// IsAdmin 检查用户是否是项目管理员
func (ps *ProjectService) IsAdmin(projectID, userID uint) bool {
	var count int64
	ps.db.Model(&requirement.ProjectMember{}).
		Where("project_id = ? AND user_id = ? AND role = ?", projectID, userID, "project_admin").
		Count(&count)
	return count > 0
}

// GetStatistics 获取项目统计
func (ps *ProjectService) GetStatistics(projectID uint) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总任务数
	var totalMissions int64
	ps.db.Model(&requirement.Mission{}).Where("project_id = ?", projectID).Count(&totalMissions)
	stats["total_missions"] = totalMissions

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	ps.db.Model(&requirement.Mission{}).
		Select("status, count(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&statusStats)

	byStatus := make(map[string]int64)
	for _, stat := range statusStats {
		byStatus[stat.Status] = stat.Count
	}
	stats["by_status"] = byStatus

	// 按优先级统计
	var priorityStats []struct {
		Priority string
		Count    int64
	}
	ps.db.Model(&requirement.Mission{}).
		Select("priority, count(*) as count").
		Where("project_id = ?", projectID).
		Group("priority").
		Scan(&priorityStats)

	byPriority := make(map[string]int64)
	for _, stat := range priorityStats {
		byPriority[stat.Priority] = stat.Count
	}
	stats["by_priority"] = byPriority

	// 按类型统计
	var typeStats []struct {
		Type  string
		Count int64
	}
	ps.db.Model(&requirement.Mission{}).
		Select("type, count(*) as count").
		Where("project_id = ?", projectID).
		Group("type").
		Scan(&typeStats)

	byType := make(map[string]int64)
	for _, stat := range typeStats {
		byType[stat.Type] = stat.Count
	}
	stats["by_type"] = byType

	// 完成率
	completedCount := byStatus["done"] + byStatus["closed"]
	if totalMissions > 0 {
		stats["completion_rate"] = float64(completedCount) / float64(totalMissions) * 100
	} else {
		stats["completion_rate"] = 0
	}

	return stats, nil
}
