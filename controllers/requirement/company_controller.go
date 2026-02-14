package requirement

import (
	"go_wails_project_manager/response"
	"go_wails_project_manager/services/requirement_service"
	"strconv"
	"github.com/gin-gonic/gin"
)

// CompanyController 公司管理控制器
type CompanyController struct {
	service *requirement_service.CompanyService
}

// NewCompanyController 创建公司控制器
func NewCompanyController() *CompanyController {
	return &CompanyController{
		service: requirement_service.NewCompanyService(),
	}
}

// CreateCompanyRequest 创建公司请求
type CreateCompanyRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
}

// UpdateCompanyRequest 更新公司请求
type UpdateCompanyRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Logo        string `json:"logo"`
	Description string `json:"description"`
}

// AddMemberRequest 添加成员请求
type AddMemberRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=company_admin member viewer"`
}

// Create 创建公司
func (cc *CompanyController) Create(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	userID := c.GetUint("user_id")

	company, err := cc.service.CreateCompany(userID, req.Name, req.Logo, req.Description)
	if err != nil {
		response.InternalServerError(c, "创建公司失败")
		return
	}

	response.Success(c, company)
}

// List 获取用户的公司列表
func (cc *CompanyController) List(c *gin.Context) {
	userID := c.GetUint("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	companies, total, err := cc.service.ListUserCompanies(userID, page, pageSize, keyword)
	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, gin.H{
		"items":     companies,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetDetail 获取公司详情
func (cc *CompanyController) GetDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查访问权限
	if !cc.service.HasAccess(uint(id), userID) {
		response.Forbidden(c, "无权访问该公司")
		return
	}

	company, err := cc.service.GetCompanyDetail(uint(id))
	if err != nil {
		response.NotFound(c, "公司不存在")
		return
	}

	response.Success(c, company)
}

// Update 更新公司
func (cc *CompanyController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查是否是管理员
	if !cc.service.IsAdmin(uint(id), userID) {
		response.Forbidden(c, "只有管理员可以更新公司信息")
		return
	}

	company, err := cc.service.UpdateCompany(uint(id), req.Name, req.Logo, req.Description)
	if err != nil {
		response.InternalServerError(c, "更新失败")
		return
	}

	response.Success(c, company)
}

// AddMember 添加成员
func (cc *CompanyController) AddMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 检查是否是管理员
	if !cc.service.IsAdmin(uint(id), userID) {
		response.Forbidden(c, "只有管理员可以添加成员")
		return
	}

	member, err := cc.service.AddMember(uint(id), req.UserID, req.Role)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, member)
}

// RemoveMember 移除成员
func (cc *CompanyController) RemoveMember(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	memberUserID, _ := strconv.ParseUint(c.Param("user_id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查是否是管理员
	if !cc.service.IsAdmin(uint(id), userID) {
		response.Forbidden(c, "只有管理员可以移除成员")
		return
	}

	if err := cc.service.RemoveMember(uint(id), uint(memberUserID)); err != nil {
		response.InternalServerError(c, "移除成员失败")
		return
	}

	response.SuccessWithMsg(c, "移除成功", nil)
}

// GetMembers 获取公司成员列表
func (cc *CompanyController) GetMembers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.GetUint("user_id")

	// 检查访问权限
	if !cc.service.HasAccess(uint(id), userID) {
		response.Forbidden(c, "无权访问该公司")
		return
	}

	members, err := cc.service.GetMembers(uint(id))
	if err != nil {
		response.InternalServerError(c, "查询失败")
		return
	}

	response.Success(c, members)
}
