package meshy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client Meshy API客户端
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient 创建Meshy客户端
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ImageTo3DRequest 图生3D请求
type ImageTo3DRequest struct {
	ImageURL             string `json:"image_url"`
	AIModel              string `json:"ai_model,omitempty"`              // meshy-5 | meshy-6 | latest
	EnablePBR            bool   `json:"enable_pbr"`
	Topology             string `json:"topology,omitempty"`              // quad | triangle (quad=低面数)
	TargetPolycount      int    `json:"target_polycount,omitempty"`      // 目标面数 (仅当topology=quad时有效，范围: 5000-30000)
	ShouldRemesh         bool   `json:"should_remesh"`
	ShouldTexture        bool   `json:"should_texture"`
	SavePreRemeshedModel bool   `json:"save_pre_remeshed_model"`
}

// SubmitResponse 提交任务响应
type SubmitResponse struct {
	Result string `json:"result"` // task_id
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID           string     `json:"id"`
	Status       string     `json:"status"` // PENDING, IN_PROGRESS, SUCCEEDED, FAILED, CANCELED
	Progress     int        `json:"progress"`
	ModelURLs    *ModelURLs `json:"model_urls,omitempty"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty"`
	TaskError    *TaskError `json:"task_error,omitempty"`
}

// ModelURLs 模型URL
type ModelURLs struct {
	GLB  string `json:"glb,omitempty"`
	FBX  string `json:"fbx,omitempty"`
	USDZ string `json:"usdz,omitempty"`
	OBJ  string `json:"obj,omitempty"`
}

// TaskError 任务错误
type TaskError struct {
	Message string `json:"message"`
}

// SubmitImageTo3D 提交图生3D任务
func (c *Client) SubmitImageTo3D(req *ImageTo3DRequest) (string, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/openapi/v1/image-to-3d", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 202 Accepted 和 200 OK 都是成功
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("API错误 (%d): %s", resp.StatusCode, string(body))
	}

	var submitResp SubmitResponse
	if err := json.Unmarshal(body, &submitResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	return submitResp.Result, nil
}

// GetTask 查询任务状态
func (c *Client) GetTask(taskID string) (*TaskResponse, error) {
	httpReq, err := http.NewRequest("GET", c.baseURL+"/openapi/v1/image-to-3d/"+taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API错误 (%d): %s", resp.StatusCode, string(body))
	}

	var taskResp TaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &taskResp, nil
}
