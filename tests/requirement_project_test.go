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

// setupProjectRouter 设置项目管理路由
func setupProjectRouter() *gin.Engine {
	router := gin.New()
	api.SetupRequirementRoutes(router, TestJWT)
	return router
}

// TestProjectCreate 测试创建项目
func TestProjectCreate(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupProjectRouter()
	
	user, err := CreateTestUser(t, "projectuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	
	t.Run("成功创建项目", func(t *testing.T) {
		req := requirementControllers.CreateProjectRequest{
			CompanyID:   company.ID,
			Name:        "测试项目",
			Key:         "TEST",
			Description: "这是一个测试项�?,
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/projects", req, token)
		project := AssertSuccessWithData[requirementModels.Project](t, w, http.StatusOK)
		
		assert.Equal(t, "测试项目", project.Name)
		assert.Equal(t, "TEST", project.Key)
		assert.Equal(t, user.ID, project.OwnerID)
	})
	
	t.Run("缺少必填字段", func(t *testing.T) {
		req := map[string]string{
			"key": "TEST2",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/projects", req, token)
		AssertError(t, w, http.StatusBadRequest, 400)
	})
}

// TestProjectList 测试获取项目列表
func TestProjectList(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupProjectRouter()
	
	user, err := CreateTestUser(t, "projectuser2", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	
	// 创建多个项目
	CreateTestProject(t, token, company.ID, "项目A", "PRJA")
	CreateTestProject(t, token, company.ID, "项目B", "PRJB")
	CreateTestProject(t, token, company.ID, "项目C", "PRJC")
	
	t.Run("获取项目列表", func(t *testing.T) {
		w := MakeRequestWithBody(t, "GET", "/api/requirement/projects?page=1&page_size=10", nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按公司筛选项�?, func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/projects?company_id=%d", company.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestProjectUpdate 测试更新项目
func TestProjectUpdate(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupProjectRouter()
	
	user, err := CreateTestUser(t, "projectuser3", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "原始项目", "ORIG")
	
	t.Run("成功更新项目", func(t *testing.T) {
		req := requirementControllers.UpdateProjectRequest{
			Name:        "更新后的项目",
			Description: "更新后的描述",
		}
		
		url := fmt.Sprintf("/api/requirement/projects/%d", project.ID)
		w := MakeRequestWithBody(t, "PUT", url, req, token)
		updated := AssertSuccessWithData[requirementModels.Project](t, w, http.StatusOK)
		
		assert.Equal(t, "更新后的项目", updated.Name)
	})
}

// TestProjectMembers 测试项目成员管理
func TestProjectMembers(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupProjectRouter()
	
	owner, err := CreateTestUser(t, "projectowner", "password123")
	assert.NoError(t, err)
	
	member, err := CreateTestUser(t, "projectmember", "password123")
	assert.NoError(t, err)
	
	ownerToken, err := GetTestToken(t, owner.ID, owner.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, ownerToken, "测试公司")
	project := CreateTestProject(t, ownerToken, company.ID, "团队项目", "TEAM")
	
	// 先将用户添加到公�?
	addCompanyMemberReq := requirementControllers.AddMemberRequest{
		UserID: member.ID,
		Role:   "member",
	}
	companyMemberURL := fmt.Sprintf("/api/requirement/companies/%d/members", company.ID)
	w := MakeRequestWithBody(t, "POST", companyMemberURL, addCompanyMemberReq, ownerToken)
	AssertSuccess(t, w, http.StatusOK)
	
	t.Run("添加成员", func(t *testing.T) {
		req := requirementControllers.AddProjectMemberRequest{
			UserID: member.ID,
			Role:   "developer",
		}
		
		url := fmt.Sprintf("/api/requirement/projects/%d/members", project.ID)
		w := MakeRequestWithBody(t, "POST", url, req, ownerToken)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("获取成员列表", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/projects/%d/members", project.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, ownerToken)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("移除成员", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/projects/%d/members/%d", project.ID, member.ID)
		w := MakeRequestWithBody(t, "DELETE", url, nil, ownerToken)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestProjectStatistics 测试项目统计
func TestProjectStatistics(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupProjectRouter()
	
	user, err := CreateTestUser(t, "projectuser4", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "统计项目", "STAT")
	
	t.Run("获取项目统计", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/projects/%d/statistics", project.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}
