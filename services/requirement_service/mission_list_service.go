package requirement_service

import (
	"errors"
	"go_wails_project_manager/database"
	"go_wails_project_manager/models/requirement"
	"time"

	"gorm.io/gorm"
)

// MissionListService 任务列表服务
type MissionListService struct {
	db *gorm.DB
}

// NewMissionListService 创建任务列表服务
func NewMissionListService() *MissionListService {
	db, _ := database.GetDB()
	return &MissionListService{db: db}
}

// CreateMissionList 创建任务列表
func (mls *MissionListService) CreateMissionList(projectID uint, name, listType, description string, startDate, endDate *time.Time) (*requirement.MissionList, error) {
	// 获取当前最大排序号
	var maxOrder int
	mls.db.Model(&requirement.MissionList{}).
		Where("project_id = ?", projectID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder)

	missionList := requirement.MissionList{
		ProjectID:   projectID,
		Name:        name,
		Type:        listType,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      "planning",
		SortOrder:   maxOrder + 1,
	}

	if err := mls.db.Create(&missionList).Error; err != nil {
		return nil, err
	}

	return &missionList, nil
}

// ListMissionLists 获取任务列表
func (mls *MissionListService) ListMissionLists(projectID uint, status string) ([]requirement.MissionList, error) {
	var lists []requirement.MissionList

	query := mls.db.Where("project_id = ?", projectID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 不再 Preload Missions，改为单独查询
	err := query.Order("sort_order ASC").Find(&lists).Error

	// 统计每个列表的任务数量
	for i := range lists {
		var count int64
		mls.db.Model(&requirement.Mission{}).
			Where("mission_list_id = ?", lists[i].ID).
			Count(&count)
		lists[i].MissionCount = int(count)
	}

	return lists, err
}

// GetMissionListDetail 获取任务列表详情
func (mls *MissionListService) GetMissionListDetail(listID uint) (*requirement.MissionList, error) {
	var missionList requirement.MissionList
	err := mls.db.Preload("Project").
		Preload("Missions").
		First(&missionList, listID).Error
	return &missionList, err
}

// UpdateMissionList 更新任务列表
func (mls *MissionListService) UpdateMissionList(listID uint, name, description, status string, startDate, endDate *time.Time) (*requirement.MissionList, error) {
	var missionList requirement.MissionList
	if err := mls.db.First(&missionList, listID).Error; err != nil {
		return nil, err
	}

	if name != "" {
		missionList.Name = name
	}
	missionList.Description = description
	if status != "" {
		missionList.Status = status
	}
	missionList.StartDate = startDate
	missionList.EndDate = endDate

	if err := mls.db.Save(&missionList).Error; err != nil {
		return nil, err
	}

	return &missionList, nil
}

// DeleteMissionList 删除任务列表
func (mls *MissionListService) DeleteMissionList(listID uint) error {
	// 检查是否有任务
	var count int64
	mls.db.Model(&requirement.Mission{}).Where("mission_list_id = ?", listID).Count(&count)
	if count > 0 {
		return errors.New("任务列表中还有任务，无法删除")
	}

	return mls.db.Delete(&requirement.MissionList{}, listID).Error
}

// UpdateSortOrder 更新排序
func (mls *MissionListService) UpdateSortOrder(listID uint, sortOrder int) error {
	return mls.db.Model(&requirement.MissionList{}).
		Where("id = ?", listID).
		Update("sort_order", sortOrder).Error
}
