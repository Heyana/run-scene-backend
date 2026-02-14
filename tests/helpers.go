package tests

import (
	"bytes"
	"encoding/json"
	"go_wails_project_manager/models"
	requirementModels "go_wails_project_manager/models/requirement"
	"go_wails_project_manager/response"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, username, password string) (*models.User, error) {
	user := &models.User{
		Username: username,
		Password: password,
		Email:    username + "@test.com",
		Status:   "active",
	}
	
	if err := TestDB.Create(user).Error; err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetTestToken 获取测试用户的JWT Token
func GetTestToken(t *testing.T, userID uint, username, role string) (string, error) {
	return TestJWT.GenerateToken(userID, username, role)
}

// MakeRequestWithBody 发送HTTP请求（使用结构体）
func MakeRequestWithBody(t *testing.T, method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	
	return w
}

// ParseResponse 解析响应为指定类型
func ParseResponse[T any](t *testing.T, w *httptest.ResponseRecorder) *response.Response {
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	return &resp
}

// ParseDataAs 解析响应数据为指定类型
func ParseDataAs[T any](t *testing.T, resp *response.Response) *T {
	dataBytes, err := json.Marshal(resp.Data)
	assert.NoError(t, err)
	
	var result T
	err = json.Unmarshal(dataBytes, &result)
	assert.NoError(t, err)
	
	return &result
}

// AssertSuccessWithData 断言成功响应并返回数据
func AssertSuccessWithData[T any](t *testing.T, w *httptest.ResponseRecorder, expectedHTTPCode int) *T {
	assert.Equal(t, expectedHTTPCode, w.Code)
	
	resp := ParseResponse[T](t, w)
	assert.Equal(t, response.CodeSuccess, resp.Code)
	
	return ParseDataAs[T](t, resp)
}

// AssertSuccess 断言成功响应
func AssertSuccess(t *testing.T, w *httptest.ResponseRecorder, expectedHTTPCode int) {
	assert.Equal(t, expectedHTTPCode, w.Code)
	
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, response.CodeSuccess, resp.Code)
}

// AssertError 断言错误响应（检查业务状态码）
func AssertError(t *testing.T, w *httptest.ResponseRecorder, expectedHTTPCode int, expectedCode int) {
	// 注意：系统统一返回 HTTP 200，业务错误通过 response.code 区分
	assert.Equal(t, http.StatusOK, w.Code)
	
	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, expectedCode, int(resp.Code))
}

// CleanupTestData 清理测试数据
func CleanupTestData(t *testing.T) {
	// 清理需求管理相关表
	TestDB.Exec("DELETE FROM requirement_mission_logs")
	TestDB.Exec("DELETE FROM requirement_mission_attachments")
	TestDB.Exec("DELETE FROM requirement_mission_comments")
	TestDB.Exec("DELETE FROM requirement_mission_relations")
	TestDB.Exec("DELETE FROM requirement_mission_tag_relations")
	TestDB.Exec("DELETE FROM requirement_mission_tags")
	TestDB.Exec("DELETE FROM requirement_missions")
	TestDB.Exec("DELETE FROM requirement_mission_lists")
	TestDB.Exec("DELETE FROM requirement_project_members")
	TestDB.Exec("DELETE FROM requirement_projects")
	TestDB.Exec("DELETE FROM requirement_company_members")
	TestDB.Exec("DELETE FROM requirement_companies")
	
	// 清理用户相关表
	TestDB.Exec("DELETE FROM users WHERE id > 1") // 保留管理员用户
}

// CreateTestCompany 创建测试公司
func CreateTestCompany(t *testing.T, token string, name string) *requirementModels.Company {
	req := map[string]interface{}{
		"name": name,
	}
	
	w := MakeRequestWithBody(t, "POST", "/api/requirement/companies", req, token)
	company := AssertSuccessWithData[requirementModels.Company](t, w, http.StatusOK)
	return company
}

// CreateTestProject 创建测试项目
func CreateTestProject(t *testing.T, token string, companyID uint, name, key string) *requirementModels.Project {
	req := map[string]interface{}{
		"company_id": companyID,
		"name":       name,
		"key":        key,
	}
	
	w := MakeRequestWithBody(t, "POST", "/api/requirement/projects", req, token)
	project := AssertSuccessWithData[requirementModels.Project](t, w, http.StatusOK)
	return project
}

// CreateTestMissionList 创建测试任务列表
func CreateTestMissionList(t *testing.T, token string, projectID uint, name, listType string) *requirementModels.MissionList {
	req := map[string]interface{}{
		"project_id": projectID,
		"name":       name,
		"type":       listType,
	}
	
	w := MakeRequestWithBody(t, "POST", "/api/requirement/mission-lists", req, token)
	list := AssertSuccessWithData[requirementModels.MissionList](t, w, http.StatusOK)
	return list
}


// RecordTestResult 记录测试结果
func RecordTestResult(t *testing.T, name, status string, subTests []string) {
	if GlobalReporter != nil {
		errorMsg := ""
		if status == "FAIL" && t.Failed() {
			errorMsg = "测试失败，请查看详细日志"
		}
		GlobalReporter.AddResult(name, status, "0s", errorMsg, subTests)
	}
}
