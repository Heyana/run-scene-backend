package requirement

import (
	"time"
	parentModels "go_wails_project_manager/models"
)

// Mission 任务
type Mission struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	MissionListID   uint       `gorm:"not null;index" json:"mission_list_id"`
	MissionColumnID *uint      `gorm:"index" json:"mission_column_id"` // 所属列ID（可选，用于看板视图）
	ProjectID       uint       `gorm:"not null;index" json:"project_id"`
	MissionKey      string     `gorm:"size:50;uniqueIndex" json:"mission_key"` // 如 PRJ-001
	Title           string     `gorm:"size:200;not null" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	Type            string     `gorm:"size:20;not null;index" json:"type"` // feature/enhancement/bug
	Priority        string     `gorm:"size:10;not null;index" json:"priority"` // P0/P1/P2/P3
	Status          string     `gorm:"size:20;not null;index" json:"status"` // todo/in_progress/done/closed
	AssigneeID      *uint      `gorm:"index" json:"assignee_id"`
	ReporterID      uint       `gorm:"not null;index" json:"reporter_id"`
	EstimatedHours  float64    `gorm:"default:0" json:"estimated_hours"`
	ActualHours     float64    `gorm:"default:0" json:"actual_hours"`
	StartDate       *time.Time `json:"start_date"`
	DueDate         *time.Time `gorm:"index" json:"due_date"`
	CompletedAt     *time.Time `json:"completed_at"`
	SortOrder       int        `gorm:"default:0" json:"sort_order"`
	CreatedAt       time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	MissionList   MissionList          `gorm:"foreignKey:MissionListID" json:"mission_list,omitempty"`
	MissionColumn *MissionColumn       `gorm:"foreignKey:MissionColumnID" json:"mission_column,omitempty"`
	Project       Project              `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Assignee      *parentModels.User   `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Reporter      parentModels.User    `gorm:"foreignKey:ReporterID" json:"reporter,omitempty"`
	Comments      []MissionComment     `gorm:"foreignKey:MissionID" json:"comments,omitempty"`
	Attachments   []MissionAttachment  `gorm:"foreignKey:MissionID" json:"attachments,omitempty"`
	Relations     []MissionRelation    `gorm:"foreignKey:SourceMissionID" json:"relations,omitempty"`
	Logs          []MissionLog         `gorm:"foreignKey:MissionID" json:"logs,omitempty"`
	Tags          []MissionTag         `gorm:"many2many:requirement_mission_tag_relations;" json:"tags,omitempty"`
}

// MissionComment 任务评论
type MissionComment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MissionID uint      `gorm:"not null;index:idx_comment_mission_created" json:"mission_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	ParentID  *uint     `gorm:"index" json:"parent_id"` // 父评论ID，支持回复
	CreatedAt time.Time `gorm:"index:idx_comment_mission_created" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Mission Mission           `gorm:"foreignKey:MissionID" json:"mission,omitempty"`
	User    parentModels.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent  *MissionComment   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies []MissionComment  `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// MissionAttachment 任务附件
type MissionAttachment struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	MissionID  uint      `gorm:"not null;index" json:"mission_id"`
	FileName   string    `gorm:"size:255;not null" json:"file_name"`
	FileURL    string    `gorm:"size:500;not null" json:"file_url"`
	FileSize   int64     `gorm:"not null" json:"file_size"`
	FileType   string    `gorm:"size:50" json:"file_type"`
	UploadedBy uint      `gorm:"not null;index" json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`

	Mission Mission           `gorm:"foreignKey:MissionID" json:"mission,omitempty"`
	User    parentModels.User `gorm:"foreignKey:UploadedBy" json:"user,omitempty"`
}

// MissionRelation 任务关联
type MissionRelation struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	SourceMissionID  uint      `gorm:"not null;index" json:"source_mission_id"`
	TargetMissionID  uint      `gorm:"not null;index" json:"target_mission_id"`
	RelationType     string    `gorm:"size:20;not null" json:"relation_type"` // depends_on/blocks/relates_to
	CreatedAt        time.Time `json:"created_at"`

	SourceMission Mission `gorm:"foreignKey:SourceMissionID" json:"source_mission,omitempty"`
	TargetMission Mission `gorm:"foreignKey:TargetMissionID" json:"target_mission,omitempty"`
}

// MissionLog 任务操作日志
type MissionLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	MissionID uint      `gorm:"not null;index:idx_log_mission_created" json:"mission_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Action    string    `gorm:"size:50;not null" json:"action"`
	Field     string    `gorm:"size:50" json:"field"`
	OldValue  string    `gorm:"type:text" json:"old_value"`
	NewValue  string    `gorm:"type:text" json:"new_value"`
	CreatedAt time.Time `gorm:"index:idx_log_mission_created" json:"created_at"`

	Mission Mission           `gorm:"foreignKey:MissionID" json:"mission,omitempty"`
	User    parentModels.User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// MissionTag 任务标签
type MissionTag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"not null;index" json:"project_id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Color     string    `gorm:"size:20" json:"color"`
	CreatedAt time.Time `json:"created_at"`

	Project Project   `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Missions []Mission `gorm:"many2many:requirement_mission_tag_relations;" json:"missions,omitempty"`
}

// TableName 指定表名
func (Mission) TableName() string {
	return "requirement_missions"
}

// TableName 指定表名
func (MissionComment) TableName() string {
	return "requirement_mission_comments"
}

// TableName 指定表名
func (MissionAttachment) TableName() string {
	return "requirement_mission_attachments"
}

// TableName 指定表名
func (MissionRelation) TableName() string {
	return "requirement_mission_relations"
}

// TableName 指定表名
func (MissionLog) TableName() string {
	return "requirement_mission_logs"
}

// TableName 指定表名
func (MissionTag) TableName() string {
	return "requirement_mission_tags"
}
