// Package database 提供贴图类型分析工具
package database

import (
	"go_wails_project_manager/logger"
	"go_wails_project_manager/models"
	"strings"

	"gorm.io/gorm"
)

// TextureTypeAnalysis 贴图类型分析结果
type TextureTypeAnalysis struct {
	OriginalType  string   `json:"original_type"`  // 原始类型名称
	Count         int      `json:"count"`          // 出现次数
	Source        string   `json:"source"`         // 数据源：polyhaven 或 ambientcg
	Examples      []string `json:"examples"`       // 示例文件名（最多5个）
	SuggestedType string   `json:"suggested_type"` // 建议的 Three.js 类型
}

// AnalyzeAllTextureTypes 分析所有贴图类型
func AnalyzeAllTextureTypes(db *gorm.DB) ([]TextureTypeAnalysis, error) {
	logger.Log.Info("开始分析所有贴图类型...")

	// 检查表是否存在
	if !db.Migrator().HasTable(&models.File{}) {
		logger.Log.Warn("File 表不存在，跳过分析")
		return []TextureTypeAnalysis{}, nil
	}
	if !db.Migrator().HasTable(&models.Texture{}) {
		logger.Log.Warn("Texture 表不存在，跳过分析")
		return []TextureTypeAnalysis{}, nil
	}

	// 查询所有贴图文件，按 texture_type 分组
	type TypeCount struct {
		TextureType string
		Count       int
		Source      string
	}

	var results []TypeCount
	
	// 使用 GORM 的方式查询，而不是原生 SQL
	// PolyHaven 数据源
	subQuery := db.Table("file").
		Select("file.texture_type, COUNT(*) as count, ? as source", "polyhaven").
		Joins("INNER JOIN texture ON file.related_id = texture.id AND file.related_type = ?", "Texture").
		Where("file.file_type = ? AND file.texture_type != ? AND texture.source = ?", "texture", "", "polyhaven").
		Group("file.texture_type")
	
	err := subQuery.Scan(&results).Error
	
	if err != nil {
		logger.Log.Warnf("查询 PolyHaven 贴图类型失败: %v", err)
		return nil, err
	}

	// AmbientCG 数据源
	var ambientResults []TypeCount
	subQuery = db.Table("file").
		Select("file.texture_type, COUNT(*) as count, ? as source", "ambientcg").
		Joins("INNER JOIN texture ON file.related_id = texture.id AND file.related_type = ?", "Texture").
		Where("file.file_type = ? AND file.texture_type != ? AND texture.source = ?", "texture", "", "ambientcg").
		Group("file.texture_type")
	
	err = subQuery.Scan(&ambientResults).Error
	
	if err != nil {
		logger.Log.Warnf("查询 AmbientCG 贴图类型失败: %v", err)
		return nil, err
	}

	// 合并结果
	results = append(results, ambientResults...)

	// 构建分析结果
	analysisMap := make(map[string]*TextureTypeAnalysis)
	
	for _, result := range results {
		key := result.Source + ":" + result.TextureType
		
		if _, exists := analysisMap[key]; !exists {
			analysisMap[key] = &TextureTypeAnalysis{
				OriginalType:  result.TextureType,
				Count:         result.Count,
				Source:        result.Source,
				Examples:      []string{},
				SuggestedType: suggestThreeJSType(result.TextureType),
			}
		}
		
		// 获取示例文件名（最多5个）
		var files []models.File
		db.Table("file").
			Select("file.file_name").
			Joins("INNER JOIN texture ON file.related_id = texture.id AND file.related_type = ?", "Texture").
			Where("file.file_type = ? AND file.texture_type = ? AND texture.source = ?", "texture", result.TextureType, result.Source).
			Limit(5).
			Scan(&files)
		
		for _, file := range files {
			analysisMap[key].Examples = append(analysisMap[key].Examples, file.FileName)
		}
	}

	// 转换为数组
	var analysis []TextureTypeAnalysis
	for _, item := range analysisMap {
		analysis = append(analysis, *item)
	}

	logger.Log.Infof("分析完成，共发现 %d 种贴图类型", len(analysis))
	return analysis, nil
}

// suggestThreeJSType 根据原始类型名称建议 Three.js 类型
func suggestThreeJSType(originalType string) string {
	// 转换为小写进行匹配
	lowerType := strings.ToLower(originalType)
	
	// 常见映射规则（与 texture_mapping.yaml 保持一致）
	mappings := map[string]string{
		// PolyHaven 常见类型
		"diffuse":      "map",
		"diff_png":     "map",
		"col":          "map",
		"col1":         "map",
		"col2":         "map",
		"coll1":        "map",
		"coll2":        "map",
		
		// AmbientCG 常见类型
		"color":        "map",
		
		// 法线贴图
		"nor_gl":       "normalMap",
		"nor_dx":       "normalMap",
		"normal_gl":    "normalMap",
		"normal_dx":    "normalMap",
		"normalgl":     "normalMap",
		"normaldx":     "normalMap",
		
		// 粗糙度
		"rough":        "roughnessMap",
		"rough_ao":     "roughnessMap",
		"rough_diff":   "roughnessMap",
		"roughness":    "roughnessMap",
		
		// 金属度
		"metal":        "metalnessMap",
		"me_arm":       "metalnessMap",
		"metalness":    "metalnessMap",
		
		// 环境光遮蔽
		"ao":           "aoMap",
		"arm":          "aoMap,roughnessMap,metalnessMap", // ARM 是组合贴图
		"ambientocclusion": "aoMap",
		
		// 位移
		"displacement": "displacementMap",
		"bump":         "displacementMap",
		
		// 高光
		"spec":         "specularMap",
		"spec_ior":     "specularMap",
		
		// 透明度
		"mask":         "alphaMap",
		"translucent":  "alphaMap",
		"opacity":      "alphaMap",
		
		// 各向异性
		"anisotropy_rotation": "anisotropyMap",
		"anisotropy_strength": "anisotropyMap",
		
		// 反射
		"ref":          "reflectionMap",
	}
	
	// 尝试精确匹配
	if suggested, ok := mappings[lowerType]; ok {
		return suggested
	}
	
	// 尝试部分匹配（检查是否包含关键词）
	for key, value := range mappings {
		if strings.Contains(lowerType, key) {
			return value
		}
	}
	
	// 默认返回未知
	return "unknown"
}
