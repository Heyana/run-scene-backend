package requirement_service

import (
	"errors"
	"go_wails_project_manager/database"
	"go_wails_project_manager/models"
	"go_wails_project_manager/models/requirement"
	"time"

	"gorm.io/gorm"
)

// CompanyService 公司服务
type CompanyService struct {
	db *gorm.DB
}

// NewCompanyService 创建公司服务
func NewCompanyService() *CompanyService {
	db, _ := database.GetDB()
	return &CompanyService{db: db}
}

// CreateCompany 创建公司
func (cs *CompanyService) CreateCompany(userID uint, name, logo, description string) (*requirement.Company, error) {
	company := requirement.Company{
		Name:        name,
		Logo:        logo,
		Description: description,
		OwnerID:     userID,
		Status:      "active",
	}

	if err := cs.db.Create(&company).Error; err != nil {
		return nil, err
	}

	// 自动添加创建人为公司管理员
	member := requirement.CompanyMember{
		CompanyID: company.ID,
		UserID:    userID,
		Role:      "company_admin",
		JoinedAt:  time.Now(),
	}
	cs.db.Create(&member)

	return &company, nil
}

// ListUserCompanies 获取用户的公司列表
func (cs *CompanyService) ListUserCompanies(userID uint, page, pageSize int, keyword string) ([]requirement.Company, int64, error) {
	var companies []requirement.Company
	var total int64

	query := cs.db.Model(&requirement.Company{}).
		Joins("JOIN requirement_company_members ON requirement_companies.id = requirement_company_members.company_id").
		Where("requirement_company_members.user_id = ?", userID)

	if keyword != "" {
		query = query.Where("requirement_companies.name LIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Preload("Owner").
		Preload("Members.User").
		Find(&companies).Error

	return companies, total, err
}

// GetCompanyDetail 获取公司详情
func (cs *CompanyService) GetCompanyDetail(companyID uint) (*requirement.Company, error) {
	var company requirement.Company
	err := cs.db.Preload("Owner").
		Preload("Members.User").
		Preload("Projects").
		First(&company, companyID).Error
	return &company, err
}

// UpdateCompany 更新公司
func (cs *CompanyService) UpdateCompany(companyID uint, name, logo, description string) (*requirement.Company, error) {
	var company requirement.Company
	if err := cs.db.First(&company, companyID).Error; err != nil {
		return nil, err
	}

	if name != "" {
		company.Name = name
	}
	company.Logo = logo
	company.Description = description

	if err := cs.db.Save(&company).Error; err != nil {
		return nil, err
	}

	return &company, nil
}

// AddMember 添加成员
func (cs *CompanyService) AddMember(companyID, userID uint, role string) (*requirement.CompanyMember, error) {
	// 检查用户是否存在
	var user models.User
	if err := cs.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查是否已经是成员
	var existingMember requirement.CompanyMember
	if err := cs.db.Where("company_id = ? AND user_id = ?", companyID, userID).
		First(&existingMember).Error; err == nil {
		return nil, errors.New("用户已经是公司成员")
	}

	member := requirement.CompanyMember{
		CompanyID: companyID,
		UserID:    userID,
		Role:      role,
		JoinedAt:  time.Now(),
	}

	if err := cs.db.Create(&member).Error; err != nil {
		return nil, err
	}

	// 加载关联数据
	cs.db.Preload("User").First(&member, member.ID)

	return &member, nil
}

// RemoveMember 移除成员
func (cs *CompanyService) RemoveMember(companyID, userID uint) error {
	// 不能移除公司所有者
	var company requirement.Company
	if err := cs.db.First(&company, companyID).Error; err != nil {
		return err
	}

	if company.OwnerID == userID {
		return errors.New("不能移除公司所有者")
	}

	return cs.db.Where("company_id = ? AND user_id = ?", companyID, userID).
		Delete(&requirement.CompanyMember{}).Error
}

// GetMembers 获取公司成员列表
func (cs *CompanyService) GetMembers(companyID uint) ([]requirement.CompanyMember, error) {
	var members []requirement.CompanyMember
	err := cs.db.Where("company_id = ?", companyID).
		Preload("User").
		Find(&members).Error
	return members, err
}

// HasAccess 检查用户是否有访问权限
func (cs *CompanyService) HasAccess(companyID, userID uint) bool {
	var count int64
	cs.db.Model(&requirement.CompanyMember{}).
		Where("company_id = ? AND user_id = ?", companyID, userID).
		Count(&count)
	return count > 0
}

// IsAdmin 检查用户是否是公司管理员
func (cs *CompanyService) IsAdmin(companyID, userID uint) bool {
	var count int64
	cs.db.Model(&requirement.CompanyMember{}).
		Where("company_id = ? AND user_id = ? AND role = ?", companyID, userID, "company_admin").
		Count(&count)
	return count > 0
}
