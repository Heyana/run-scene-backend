package requirement_service

import (
	"go_wails_project_manager/models/requirement"
	"gorm.io/gorm"
)

// CreateMissionColumn 创建任务列
func CreateMissionColumn(db *gorm.DB, column *requirement.MissionColumn) error {
	return db.Create(column).Error
}

// GetMissionColumnList 获取任务列列表
func GetMissionColumnList(db *gorm.DB, missionListID uint) ([]requirement.MissionColumn, error) {
	var columns []requirement.MissionColumn
	err := db.Where("mission_list_id = ?", missionListID).
		Order("sort_order ASC, id ASC").
		Find(&columns).Error
	return columns, err
}

// GetMissionColumnByID 根据ID获取任务列
func GetMissionColumnByID(db *gorm.DB, id uint) (*requirement.MissionColumn, error) {
	var column requirement.MissionColumn
	err := db.First(&column, id).Error
	if err != nil {
		return nil, err
	}
	return &column, nil
}

// UpdateMissionColumn 更新任务列
func UpdateMissionColumn(db *gorm.DB, column *requirement.MissionColumn) error {
	return db.Save(column).Error
}

// DeleteMissionColumn 删除任务列
func DeleteMissionColumn(db *gorm.DB, id uint) error {
	// 先将该列下的任务的 mission_column_id 设置为 NULL
	db.Model(&requirement.Mission{}).
		Where("mission_column_id = ?", id).
		Update("mission_column_id", nil)
	
	// 删除列
	return db.Delete(&requirement.MissionColumn{}, id).Error
}

// UpdateMissionColumnOrder 更新任务列排序
func UpdateMissionColumnOrder(db *gorm.DB, columnID uint, sortOrder int) error {
	return db.Model(&requirement.MissionColumn{}).
		Where("id = ?", columnID).
		Update("sort_order", sortOrder).Error
}
