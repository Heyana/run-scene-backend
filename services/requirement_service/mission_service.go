package requirement_service

import (
	"errors"
	"fmt"
	"go_wails_project_manager/config"
	"go_wails_project_manager/database"
	"go_wails_project_manager/models/requirement"
	"time"

	"gorm.io/gorm"
)

// MissionService 任务服务
type MissionService struct {
	db *gorm.DB
}

// NewMissionService 创建任务服务
func NewMissionService() *MissionService {
	db, _ := database.GetDB()
	return &MissionService{db: db}
}

// CreateMissionRequest 创建任务请求
type CreateMissionRequest struct {
	MissionListID  uint
	Title          string
	Description    string
	Type           string
	Priority       string
	AssigneeID     *uint
	EstimatedHours float64
	StartDate      *time.Time
	DueDate        *time.Time
}

// CreateMission 创建任务
func (ms *MissionService) CreateMission(userID uint, req CreateMissionRequest) (*requirement.Mission, error) {
	// 获取任务列表信息
	var missionList requirement.MissionList
	if err := ms.db.Preload("Project").First(&missionList, req.MissionListID).Error; err != nil {
		return nil, errors.New("任务列表不存在")
	}

	// 生成任务编号
	missionKey, err := ms.generateMissionKey(missionList.ProjectID, missionList.Project.Key)
	if err != nil {
		return nil, err
	}

	// 获取默认值
	if req.Priority == "" {
		req.Priority = config.RequirementCfg.Requirement.Mission.DefaultPriority
	}

	// 获取最大排序号
	var maxOrder int
	ms.db.Model(&requirement.Mission{}).
		Where("mission_list_id = ?", req.MissionListID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder)

	mission := requirement.Mission{
		MissionListID:  req.MissionListID,
		ProjectID:      missionList.ProjectID,
		MissionKey:     missionKey,
		Title:          req.Title,
		Description:    req.Description,
		Type:           req.Type,
		Priority:       req.Priority,
		Status:         config.RequirementCfg.Requirement.Mission.DefaultStatus,
		AssigneeID:     req.AssigneeID,
		ReporterID:     userID,
		EstimatedHours: req.EstimatedHours,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		SortOrder:      maxOrder + 1,
	}

	if err := ms.db.Create(&mission).Error; err != nil {
		return nil, err
	}

	// 记录日志
	ms.logOperation(mission.ID, userID, "created", "", "", "")

	// 加载关联数据
	ms.db.Preload("Assignee").Preload("Reporter").First(&mission, mission.ID)

	return &mission, nil
}

// generateMissionKey 生成任务编号
func (ms *MissionService) generateMissionKey(projectID uint, projectKey string) (string, error) {
	var count int64
	ms.db.Model(&requirement.Mission{}).Where("project_id = ?", projectID).Count(&count)
	return fmt.Sprintf("%s-%d", projectKey, count+1), nil
}

// ListMissions 获取任务列表
func (ms *MissionService) ListMissions(missionListID uint, page, pageSize int, filters map[string]interface{}) ([]requirement.Mission, int64, error) {
	var missions []requirement.Mission
	var total int64

	query := ms.db.Model(&requirement.Mission{}).Where("mission_list_id = ?", missionListID)

	// 应用筛选条件
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if assigneeID, ok := filters["assignee_id"]; ok && assigneeID != nil {
		query = query.Where("assignee_id = ?", assigneeID)
	}
	if missionType, ok := filters["type"]; ok && missionType != "" {
		query = query.Where("type = ?", missionType)
	}
	if keyword, ok := filters["keyword"]; ok && keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("sort_order ASC").
		Preload("Assignee").
		Preload("Reporter").
		Preload("Tags").
		Find(&missions).Error

	return missions, total, err
}

// GetMissionDetail 获取任务详情
func (ms *MissionService) GetMissionDetail(missionID uint) (*requirement.Mission, error) {
	var mission requirement.Mission
	err := ms.db.Preload("MissionList").
		Preload("Project").
		Preload("Assignee").
		Preload("Reporter").
		Preload("Comments.User").
		Preload("Attachments.User").
		Preload("Relations.TargetMission").
		Preload("Tags").
		Preload("Logs.User").
		First(&mission, missionID).Error
	return &mission, err
}

// UpdateMission 更新任务
func (ms *MissionService) UpdateMission(missionID, userID uint, updates map[string]interface{}) (*requirement.Mission, error) {
	var mission requirement.Mission
	if err := ms.db.First(&mission, missionID).Error; err != nil {
		return nil, err
	}

	// 记录变更
	for field, newValue := range updates {
		var oldValue interface{}
		switch field {
		case "title":
			oldValue = mission.Title
		case "description":
			oldValue = mission.Description
		case "status":
			oldValue = mission.Status
		case "priority":
			oldValue = mission.Priority
		case "assignee_id":
			oldValue = mission.AssigneeID
		}
		ms.logOperation(missionID, userID, "updated", field, fmt.Sprint(oldValue), fmt.Sprint(newValue))
	}

	if err := ms.db.Model(&mission).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 如果状态变为完成，记录完成时间
	if status, ok := updates["status"]; ok && status == "done" {
		now := time.Now()
		ms.db.Model(&mission).Update("completed_at", now)
	}

	// 重新加载
	ms.db.Preload("Assignee").Preload("Reporter").First(&mission, missionID)

	return &mission, nil
}

// UpdateStatus 更新任务状态
func (ms *MissionService) UpdateStatus(missionID, userID uint, newStatus string) error {
	var mission requirement.Mission
	if err := ms.db.First(&mission, missionID).Error; err != nil {
		return err
	}

	// 检查状态流转是否合法
	if !ms.canTransitionTo(mission.Status, newStatus) {
		return errors.New("不允许的状态流转")
	}

	oldStatus := mission.Status
	mission.Status = newStatus

	if newStatus == "done" {
		now := time.Now()
		mission.CompletedAt = &now
	}

	if err := ms.db.Save(&mission).Error; err != nil {
		return err
	}

	ms.logOperation(missionID, userID, "status_changed", "status", oldStatus, newStatus)

	return nil
}

// canTransitionTo 检查状态流转是否合法
func (ms *MissionService) canTransitionTo(currentStatus, newStatus string) bool {
	transitions := map[string][]string{
		"todo":        {"in_progress", "closed"},
		"in_progress": {"done", "todo", "closed"},
		"done":        {"closed", "in_progress"},
		"closed":      {"todo"},
	}

	allowedStatuses, ok := transitions[currentStatus]
	if !ok {
		return false
	}

	for _, status := range allowedStatuses {
		if status == newStatus {
			return true
		}
	}
	return false
}

// DeleteMission 删除任务
func (ms *MissionService) DeleteMission(missionID, userID uint) error {
	ms.logOperation(missionID, userID, "deleted", "", "", "")
	return ms.db.Delete(&requirement.Mission{}, missionID).Error
}

// AddComment 添加评论
func (ms *MissionService) AddComment(missionID, userID uint, content string, parentID *uint) (*requirement.MissionComment, error) {
	comment := requirement.MissionComment{
		MissionID: missionID,
		UserID:    userID,
		Content:   content,
		ParentID:  parentID,
	}

	if err := ms.db.Create(&comment).Error; err != nil {
		return nil, err
	}

	ms.logOperation(missionID, userID, "commented", "", "", content)

	ms.db.Preload("User").First(&comment, comment.ID)

	return &comment, nil
}

// logOperation 记录操作日志
func (ms *MissionService) logOperation(missionID, userID uint, action, field, oldValue, newValue string) {
	log := requirement.MissionLog{
		MissionID: missionID,
		UserID:    userID,
		Action:    action,
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
	}
	ms.db.Create(&log)
}

// CanEditMission 检查是否可以编辑任务
func (ms *MissionService) CanEditMission(missionID, userID uint, projectRole string) bool {
	var mission requirement.Mission
	if err := ms.db.First(&mission, missionID).Error; err != nil {
		return false
	}

	// 项目管理员可以编辑所有任务
	if projectRole == "project_admin" {
		return true
	}

	// 任务负责人可以编辑
	if mission.AssigneeID != nil && *mission.AssigneeID == userID {
		return true
	}

	// 任务创建人可以编辑
	if mission.ReporterID == userID {
		return true
	}

	return false
}
