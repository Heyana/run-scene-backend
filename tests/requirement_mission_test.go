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

// setupMissionRouter 设置任务路由
func setupMissionRouter() *gin.Engine {
	router := gin.New()
	api.SetupRequirementRoutes(router, TestJWT)
	return router
}

// TestMissionCreate 测试创建任务
func TestMissionCreate(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "Sprint 1", "sprint")
	
	t.Run("成功创建任务", func(t *testing.T) {
		req := requirementControllers.CreateMissionRequestBody{
			MissionListID: list.ID,
			Title:         "实现用户登录功能",
			Description:   "需要实现用户登录和注册功能",
			Type:          "feature",
			Priority:      "P1",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", req, token)
		mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
		
		assert.Equal(t, "实现用户登录功能", mission.Title)
		assert.Equal(t, "feature", mission.Type)
		assert.Equal(t, "P1", mission.Priority)
		assert.NotEmpty(t, mission.MissionKey)
	})
	
	t.Run("缺少必填字段", func(t *testing.T) {
		req := map[string]uint{
			"mission_list_id": list.ID,
			"project_id":      project.ID,
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", req, token)
		AssertError(t, w, http.StatusBadRequest, 400)
	})
}

// TestMissionList 测试获取任务列表
func TestMissionList(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser2", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "Sprint 1", "sprint")
	
	// 创建多个任务
	missions := []requirementControllers.CreateMissionRequestBody{
		{
			MissionListID: list.ID,
			Title:         "任务A",
			Type:          "feature",
			Priority:      "P0",
		},
		{
			MissionListID: list.ID,
			Title:         "任务B",
			Type:          "bug",
			Priority:      "P1",
		},
		{
			MissionListID: list.ID,
			Title:         "任务C",
			Type:          "enhancement",
			Priority:      "P2",
		},
	}
	
	for _, req := range missions {
		MakeRequestWithBody(t, "POST", "/api/requirement/missions", req, token)
	}
	
	t.Run("获取任务列表", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions?project_id=%d", project.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按任务列表筛�?, func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions?mission_list_id=%d", list.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按类型筛�?, func(t *testing.T) {
		w := MakeRequestWithBody(t, "GET", "/api/requirement/missions?type=bug", nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按优先级筛�?, func(t *testing.T) {
		w := MakeRequestWithBody(t, "GET", "/api/requirement/missions?priority=P0", nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestMissionUpdate 测试更新任务
func TestMissionUpdate(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser3", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "Sprint 1", "sprint")
	
	// 创建任务
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "原始任务",
		Type:          "feature",
		Priority:      "P2",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("成功更新任务", func(t *testing.T) {
		req := requirementControllers.UpdateMissionRequestBody{
			Title:       "更新后的任务",
			Description: "更新后的描述",
			Priority:    "P0",
		}
		
		url := fmt.Sprintf("/api/requirement/missions/%d", mission.ID)
		w := MakeRequestWithBody(t, "PUT", url, req, token)
		updated := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
		
		assert.Equal(t, "更新后的任务", updated.Title)
		assert.Equal(t, "P0", updated.Priority)
	})
}

// TestMissionStatusUpdate 测试更新任务状�?
func TestMissionStatusUpdate(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser4", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "Sprint 1", "sprint")
	
	// 创建任务
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "测试任务",
		Type:          "feature",
		Priority:      "P1",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("更新任务状�?, func(t *testing.T) {
		req := map[string]string{
			"status": "in_progress",
		}
		
		url := fmt.Sprintf("/api/requirement/missions/%d/status", mission.ID)
		w := MakeRequestWithBody(t, "PATCH", url, req, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestMissionComments 测试任务评论
func TestMissionComments(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser5", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "Sprint 1", "sprint")
	
	// 创建任务
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "测试任务",
		Type:          "feature",
		Priority:      "P1",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("添加评论", func(t *testing.T) {
		req := requirementControllers.AddCommentRequest{
			Content: "这是一条测试评�?,
		}
		
		url := fmt.Sprintf("/api/requirement/missions/%d/comments", mission.ID)
		w := MakeRequestWithBody(t, "POST", url, req, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestMissionDelete 测试删除任务
func TestMissionDelete(t *testing.T) {`n`tRecordTestResult(t)`n`tCleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser6", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "测试公司")
	project := CreateTestProject(t, token, company.ID, "测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "Sprint 1", "sprint")
	
	// 创建任务
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "待删除任�?,
		Type:          "feature",
		Priority:      "P3",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("成功删除任务", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions/%d", mission.ID)
		w := MakeRequestWithBody(t, "DELETE", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}
