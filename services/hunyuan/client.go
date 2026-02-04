package hunyuan

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// HunyuanClient 混元3D API客户端
type HunyuanClient struct {
	secretID  string
	secretKey string
	region    string
	endpoint  string
}

// NewHunyuanClient 创建客户端
func NewHunyuanClient(secretID, secretKey, region string) *HunyuanClient {
	return &HunyuanClient{
		secretID:  secretID,
		secretKey: secretKey,
		region:    region,
		endpoint:  "ai3d.tencentcloudapi.com",
	}
}

// GenerateParams 生成参数
type GenerateParams struct {
	Model           string      `json:"Model,omitempty"`
	Prompt          *string     `json:"Prompt,omitempty"`
	ImageBase64     *string     `json:"ImageBase64,omitempty"`
	ImageURL        *string     `json:"ImageUrl,omitempty"`
	MultiViewImages []ViewImage `json:"MultiViewImages,omitempty"`
	EnablePBR       *bool       `json:"EnablePBR,omitempty"`
	FaceCount       *int        `json:"FaceCount,omitempty"`
	GenerateType    string      `json:"GenerateType,omitempty"`
	PolygonType     *string     `json:"PolygonType,omitempty"`
	ResultFormat    *string     `json:"ResultFormat,omitempty"`
}

// ViewImage 多视角图片
type ViewImage struct {
	View        string `json:"View"`
	ImageBase64 string `json:"ImageBase64"`
}

// SubmitJobResponse 提交任务响应
type SubmitJobResponse struct {
	Response struct {
		JobID     string `json:"JobId"`
		RequestID string `json:"RequestId"`
		Error     *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"Response"`
}

// QueryJobResponse 查询任务响应
type QueryJobResponse struct {
	Response struct {
		Status       string  `json:"Status"`
		ErrorCode    string  `json:"ErrorCode"`
		ErrorMessage string  `json:"ErrorMessage"`
		ResultFiles  []File3D `json:"ResultFile3Ds"`
		RequestID    string  `json:"RequestId"`
		Error        *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"Response"`
}

// File3D 3D文件
type File3D struct {
	Type            string `json:"Type"`
	URL             string `json:"Url"`
	PreviewImageURL string `json:"PreviewImageUrl"`
}

// SubmitJob 提交任务
func (c *HunyuanClient) SubmitJob(params *GenerateParams) (string, error) {
	action := "SubmitHunyuanTo3DProJob"
	version := "2025-05-13"

	// 构建请求体
	payload, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("序列化参数失败: %w", err)
	}

	// 发送请求
	respBody, err := c.doRequest(action, version, payload)
	if err != nil {
		return "", err
	}

	// 解析响应
	var resp SubmitJobResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if resp.Response.Error != nil {
		return "", fmt.Errorf("API错误: %s - %s", resp.Response.Error.Code, resp.Response.Error.Message)
	}

	return resp.Response.JobID, nil
}

// QueryJob 查询任务
func (c *HunyuanClient) QueryJob(jobID string) (*QueryJobResponse, error) {
	action := "QueryHunyuanTo3DProJob"
	version := "2025-05-13"

	// 构建请求体
	payload, err := json.Marshal(map[string]string{
		"JobId": jobID,
	})
	if err != nil {
		return nil, fmt.Errorf("序列化参数失败: %w", err)
	}

	// 发送请求
	respBody, err := c.doRequest(action, version, payload)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var resp QueryJobResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if resp.Response.Error != nil {
		return nil, fmt.Errorf("API错误: %s - %s", resp.Response.Error.Code, resp.Response.Error.Message)
	}

	return &resp, nil
}

// doRequest 执行HTTP请求
func (c *HunyuanClient) doRequest(action, version string, payload []byte) ([]byte, error) {
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")

	// 构建请求
	url := fmt.Sprintf("https://%s", c.endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", c.endpoint)
	req.Header.Set("X-TC-Action", action)
	req.Header.Set("X-TC-Version", version)
	req.Header.Set("X-TC-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-TC-Region", c.region)

	// 生成签名
	authorization := c.sign(action, payload, timestamp, date)
	req.Header.Set("Authorization", authorization)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// sign 生成腾讯云API签名v3
func (c *HunyuanClient) sign(action string, payload []byte, timestamp int64, date string) string {
	// 1. 拼接规范请求串
	httpRequestMethod := "POST"
	canonicalURI := "/"
	canonicalQueryString := ""
	canonicalHeaders := fmt.Sprintf("content-type:application/json\nhost:%s\n", c.endpoint)
	signedHeaders := "content-type;host"
	hashedRequestPayload := sha256Hex(payload)

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		httpRequestMethod,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		hashedRequestPayload)

	// 2. 拼接待签名字符串
	algorithm := "TC3-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/ai3d/tc3_request", date)
	hashedCanonicalRequest := sha256Hex([]byte(canonicalRequest))

	stringToSign := fmt.Sprintf("%s\n%d\n%s\n%s",
		algorithm,
		timestamp,
		credentialScope,
		hashedCanonicalRequest)

	// 3. 计算签名
	secretDate := hmacSHA256([]byte("TC3"+c.secretKey), []byte(date))
	secretService := hmacSHA256(secretDate, []byte("ai3d"))
	secretSigning := hmacSHA256(secretService, []byte("tc3_request"))
	signature := hex.EncodeToString(hmacSHA256(secretSigning, []byte(stringToSign)))

	// 4. 拼接 Authorization
	authorization := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		c.secretID,
		credentialScope,
		signedHeaders,
		signature)

	return authorization
}

// sha256Hex 计算SHA256哈希并返回十六进制字符串
func sha256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hmacSHA256 计算HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
