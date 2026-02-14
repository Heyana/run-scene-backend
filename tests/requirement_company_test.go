package tests

import (
	"fmt"
	"go_wails_project_manager/api"
	requirementControllers "go_wails_project_manager/controllers/requirement"
	requirementModels "go_wails_project_manager/models/requirement"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupCompanyRouter 设置公司管理路由
func setupCompanyRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery()) // 添加恢复中间件
	api.SetupRequirementRoutes(router, TestJWT, TestDB)
	return router
}

// TestCompanyCreate 测试创建公司
func TestCompanyCreate(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupCompanyRouter()
	
	// 创建测试用户
	user, err := CreateTestUser(t, "testuser", "password123")
	assert.NoError(t, err)
	
	// 获取Token
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	// 测试创建公司
	t.Run("成功创建公司", func(t *testing.T) {
		req := requirementControllers.CreateCompanyRequest{
			Name:        "测试公司",
			Description: "这是一个测试公司",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/companies", req, token)
		company := AssertSuccessWithData[requirementModels.Company](t, w, http.StatusOK)
		
		assert.Equal(t, "测试公司", company.Name)
		assert.Equal(t, "这是一个测试公司", company.Description)
		assert.Equal(t, user.ID, company.OwnerID)
	})
	
	t.Run("缺少必填字段", func(t *testing.T) {
		req := map[string]interface{}{
			"description": "缺少名称",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/companies", req, token)
		AssertError(t, w, http.StatusBadRequest, 400)
	})
	
	t.Run("未认证访问", func(t *testing.T) {
		req := requirementControllers.CreateCompanyRequest{
			Name: "测试公司2",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/companies", req, "")
		AssertError(t, w, http.StatusOK, 401) // HTTP 200，业务状态码 401
	})
}

// TestCompanyList 测试获取公司列表
func TestCompanyList(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupCompanyRouter()
	
	// 创建测试用户
	user, err := CreateTestUser(t, "testuser2", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	// 创建多个公司
	CreateTestCompany(t, token, "公司A")
	CreateTestCompany(t, token, "公司B")
	CreateTestCompany(t, token, "公司C")
	
	t.Run("获取公司列表", func(t *testing.T) {
		w := MakeRequestWithBody(t, "GET", "/api/requirement/companies?page=1&page_size=10", nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("搜索公司", func(t *testing.T) {
		w := MakeRequestWithBody(t, "GET", "/api/requirement/companies?keyword=公司A", nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestCompanyUpdate 测试更新公司
func TestCompanyUpdate(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupCompanyRouter()
	
	user, err := CreateTestUser(t, "testuser3", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	// 创建公司
	company := CreateTestCompany(t, token, "原始公司")
	
	t.Run("成功更新公司", func(t *testing.T) {
		req := requirementControllers.UpdateCompanyRequest{
			Name:        "更新后的公司",
			Description: "更新后的描述",
		}
		
		url := fmt.Sprintf("/api/requirement/companies/%d", company.ID)
		w := MakeRequestWithBody(t, "PUT", url, req, token)
		updated := AssertSuccessWithData[requirementModels.Company](t, w, http.StatusOK)
		
		assert.Equal(t, "更新后的公司", updated.Name)
		assert.Equal(t, "更新后的描述", updated.Description)
	})
}

// TestCompanyMembers 测试公司成员管理
func TestCompanyMembers(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupCompanyRouter()
	
	// 创建两个用户
	owner, err := CreateTestUser(t, "owner", "password123")
	assert.NoError(t, err)
	
	member, err := CreateTestUser(t, "member", "password123")
	assert.NoError(t, err)
	
	ownerToken, err := GetTestToken(t, owner.ID, owner.Username, "admin")
	assert.NoError(t, err)
	
	// 创建公司
	company := CreateTestCompany(t, ownerToken, "团队公司")
	
	t.Run("添加成员", func(t *testing.T) {
		req := requirementControllers.AddMemberRequest{
			UserID: member.ID,
			Role:   "member",
		}
		
		url := fmt.Sprintf("/api/requirement/companies/%d/members", company.ID)
		w := MakeRequestWithBody(t, "POST", url, req, ownerToken)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("获取成员列表", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/companies/%d/members", company.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, ownerToken)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("移除成员", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/companies/%d/members/%d", company.ID, member.ID)
		w := MakeRequestWithBody(t, "DELETE", url, nil, ownerToken)
		AssertSuccess(t, w, http.StatusOK)
	})
}
