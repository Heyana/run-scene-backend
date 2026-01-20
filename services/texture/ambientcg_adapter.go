package texture

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AmbientCGAdapter AmbientCG API 适配器
type AmbientCGAdapter struct {
	baseURL    string
	httpClient *http.Client
}

// NewAmbientCGAdapter 创建 AmbientCG 适配器
func NewAmbientCGAdapter(baseURL string, timeout time.Duration) *AmbientCGAdapter {
	return &AmbientCGAdapter{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// AmbientCGMaterial AmbientCG 材质数据结构
type AmbientCGMaterial struct {
	AssetID          string                 `json:"assetId"`
	ReleaseDate      string                 `json:"releaseDate"`
	DataType         string                 `json:"dataType"`
	CreationMethod   string                 `json:"creationMethod"`
	DownloadCount    int                    `json:"downloadCount"`
	Tags             []string               `json:"tags"`
	DisplayName      string                 `json:"displayName"`
	Description      string                 `json:"description"`
	DisplayCategory  string                 `json:"displayCategory"`
	Maps             []string               `json:"maps"`
	PreviewImage     map[string]string      `json:"previewImage"`
	DownloadFolders  map[string]interface{} `json:"downloadFolders"`
}

// AmbientCGListResponse 列表响应
type AmbientCGListResponse struct {
	NumberOfResults int                 `json:"numberOfResults"`
	NextPageHTTP    string              `json:"nextPageHttp"`
	FoundAssets     []AmbientCGMaterial `json:"foundAssets"`
}

// AmbientCGDetailResponse 详情响应
type AmbientCGDetailResponse struct {
	FoundAssets []AmbientCGMaterial `json:"foundAssets"`
}

// AmbientCGDownload 下载信息
type AmbientCGDownload struct {
	DownloadLink string `json:"downloadLink"`
	FileName     string `json:"fileName"`
	Size         int64  `json:"size"`
	Attribute    string `json:"attribute"` // 1K-JPG, 2K-JPG, etc.
}

// GetMaterialList 获取材质列表
func (a *AmbientCGAdapter) GetMaterialList(limit, offset int) (*AmbientCGListResponse, error) {
	url := fmt.Sprintf("%s/api/v2/full_json?type=Material&sort=Popular&limit=%d&offset=%d",
		a.baseURL, limit, offset)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	var result AmbientCGListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return &result, nil
}

// GetMaterialDetail 获取材质详情（包含下载链接）
func (a *AmbientCGAdapter) GetMaterialDetail(assetID string) (*AmbientCGMaterial, error) {
	url := fmt.Sprintf("%s/api/v2/full_json?id=%s&include=downloadData", a.baseURL, assetID)

	resp, err := a.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态 %d: %s", resp.StatusCode, string(body))
	}

	var result AmbientCGDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	if len(result.FoundAssets) == 0 {
		return nil, fmt.Errorf("未找到材质: %s", assetID)
	}

	return &result.FoundAssets[0], nil
}

// GetDownloads 从材质详情中提取下载列表
func (a *AmbientCGAdapter) GetDownloads(material *AmbientCGMaterial) ([]AmbientCGDownload, error) {
	if material.DownloadFolders == nil {
		return nil, fmt.Errorf("无下载信息")
	}

	// 解析 downloadFolders 结构
	defaultFolder, ok := material.DownloadFolders["default"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的下载文件夹结构")
	}

	categories, ok := defaultFolder["downloadFiletypeCategories"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的文件类型分类")
	}

	zipCategory, ok := categories["zip"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("未找到 ZIP 下载选项")
	}

	downloadsRaw, ok := zipCategory["downloads"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的下载列表")
	}

	var downloads []AmbientCGDownload
	for _, dlRaw := range downloadsRaw {
		dlMap, ok := dlRaw.(map[string]interface{})
		if !ok {
			continue
		}

		download := AmbientCGDownload{
			DownloadLink: getString(dlMap, "downloadLink"),
			FileName:     getString(dlMap, "fileName"),
			Size:         getInt64(dlMap, "size"),
			Attribute:    getString(dlMap, "attribute"),
		}

		if download.DownloadLink != "" {
			downloads = append(downloads, download)
		}
	}

	return downloads, nil
}

// SelectBestDownload 选择最佳下载选项
func (a *AmbientCGAdapter) SelectBestDownload(downloads []AmbientCGDownload, resolution, format string) *AmbientCGDownload {
	if resolution == "" {
		resolution = "2K"
	}
	if format == "" {
		format = "JPG"
	}

	target := resolution + "-" + format

	// 查找匹配的
	for _, dl := range downloads {
		if dl.Attribute == target {
			return &dl
		}
	}

	// 如果没找到，返回第一个
	if len(downloads) > 0 {
		return &downloads[0]
	}

	return nil
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return 0
}
