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

// setupMissionRouter 设置任务管理路由
func setupMissionRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	api.SetupRequirementRoutes(router, TestJWT)
	return router
}

// TestMissionCreate 测试创建任务
func TestMissionCreate(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "missionuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "任务测试公司")
	project := CreateTestProject(t, token, company.ID, "任务测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "待办事项", "sprint")
	
	t.Run("成功创建任务", func(t *testing.T) {
		req := requirementControllers.CreateMissionRequestBody{
			MissionListID: list.ID,
			Title:         "实现用户登录功能",
			Description:   "需要实现用户登录、注册和密码重置功能",
			Type:          "feature",
			Priority:      "P1",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", req, token)
		mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
		
		assert.Equal(t, "实现用户登录功能", mission.Title)
		assert.Equal(t, "feature", mission.Type)
		assert.Equal(t, "P1", mission.Priority)
		assert.Equal(t, "todo", mission.Status)
	})
	
	t.Run("缺少必填字段", func(t *testing.T) {
		req := map[string]interface{}{
			"mission_list_id": list.ID,
			"description":     "缺少标题",
		}
		
		w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", req, token)
		AssertError(t, w, http.StatusOK, 400)
	})
}

// TestMissionList 测试获取任务列表
func TestMissionList(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "listuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "列表测试公司")
	project := CreateTestProject(t, token, company.ID, "列表测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "任务列表", "sprint")
	
	// 创建多个任务
	missions := []struct {
		title    string
		mType    string
		priority string
	}{
		{"任务1", "feature", "P0"},
		{"任务2", "bug", "P1"},
		{"任务3", "enhancement", "P2"},
	}
	
	for _, m := range missions {
		req := requirementControllers.CreateMissionRequestBody{
			MissionListID: list.ID,
			Title:         m.title,
			Type:          m.mType,
			Priority:      m.priority,
		}
		MakeRequestWithBody(t, "POST", "/api/requirement/missions", req, token)
	}
	
	t.Run("获取任务列表", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions?mission_list_id=%d", list.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按任务列表筛选", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions?mission_list_id=%d&page=1&page_size=10", list.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按类型筛选", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions?mission_list_id=%d&type=bug", list.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
	
	t.Run("按优先级筛选", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions?mission_list_id=%d&priority=P0", list.ID)
		w := MakeRequestWithBody(t, "GET", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestMissionUpdate 测试更新任务
func TestMissionUpdate(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "updateuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "更新测试公司")
	project := CreateTestProject(t, token, company.ID, "更新测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "待更新", "sprint")
	
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "原始任务",
		Description:   "原始描述",
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
		assert.Equal(t, "更新后的描述", updated.Description)
		assert.Equal(t, "P0", updated.Priority)
	})
}

// TestMissionStatusUpdate 测试更新任务状态
func TestMissionStatusUpdate(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "statususer", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "状态测试公司")
	project := CreateTestProject(t, token, company.ID, "状态测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "状态列表", "sprint")
	
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "待更新状态的任务",
		Type:          "feature",
		Priority:      "P1",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("更新任务状态", func(t *testing.T) {
		req := map[string]string{
			"status": "in_progress",
		}
		
		url := fmt.Sprintf("/api/requirement/missions/%d/status", mission.ID)
		w := MakeRequestWithBody(t, "PATCH", url, req, token)
		AssertSuccess(t, w, http.StatusOK)
	})
}

// TestMissionComments 测试任务评论
func TestMissionComments(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "commentuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "评论测试公司")
	project := CreateTestProject(t, token, company.ID, "评论测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "评论列表", "sprint")
	
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "需要评论的任务",
		Type:          "feature",
		Priority:      "P1",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("添加评论", func(t *testing.T) {
		req := requirementControllers.AddCommentRequest{
			Content: "这是一条测试评论",
		}
		
		url := fmt.Sprintf("/api/requirement/missions/%d/comments", mission.ID)
		w := MakeRequestWithBody(t, "POST", url, req, token)
		comment := AssertSuccessWithData[requirementModels.MissionComment](t, w, http.StatusOK)
		
		assert.Equal(t, "这是一条测试评论", comment.Content)
		assert.Equal(t, mission.ID, comment.MissionID)
	})
}

// TestMissionDelete 测试删除任务
func TestMissionDelete(t *testing.T) {
	CleanupTestData(t)
	TestRouter = setupMissionRouter()
	
	user, err := CreateTestUser(t, "deleteuser", "password123")
	assert.NoError(t, err)
	
	token, err := GetTestToken(t, user.ID, user.Username, "admin")
	assert.NoError(t, err)
	
	company := CreateTestCompany(t, token, "删除测试公司")
	project := CreateTestProject(t, token, company.ID, "删除测试项目", "TEST")
	list := CreateTestMissionList(t, token, project.ID, "删除列表", "sprint")
	
	createReq := requirementControllers.CreateMissionRequestBody{
		MissionListID: list.ID,
		Title:         "待删除任务",
		Type:          "feature",
		Priority:      "P3",
	}
	w := MakeRequestWithBody(t, "POST", "/api/requirement/missions", createReq, token)
	mission := AssertSuccessWithData[requirementModels.Mission](t, w, http.StatusOK)
	
	t.Run("成功删除任务", func(t *testing.T) {
		url := fmt.Sprintf("/api/requirement/missions/%d", mission.ID)
		w := MakeRequestWithBody(t, "DELETE", url, nil, token)
		AssertSuccess(t, w, http.StatusOK)
		
		// 验证任务已被删除
		w = MakeRequestWithBody(t, "GET", url, nil, token)
		AssertError(t, w, http.StatusOK, 404)
	})
}
