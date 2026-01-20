package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TextureMappingConfig 贴图映射配置
type TextureMappingConfig struct {
	ThreeJS      map[string][]string `yaml:"threejs"`
	DisplayNames map[string]string   `yaml:"display_names"`
}

// TextureMapping 全局贴图映射配置
var TextureMapping *TextureMappingConfig

// LoadTextureMappingConfig 加载贴图映射配置
func LoadTextureMappingConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var config TextureMappingConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	TextureMapping = &config
	return nil
}

// GetThreeJSTypes 获取 Three.js 的所有贴图类型
func (c *TextureMappingConfig) GetThreeJSTypes() []string {
	if c == nil || c.ThreeJS == nil {
		return []string{}
	}

	types := make([]string, 0, len(c.ThreeJS))
	for mapType := range c.ThreeJS {
		types = append(types, mapType)
	}
	return types
}

// MapToThreeJS 将原始类型映射到 Three.js 类型
func (c *TextureMappingConfig) MapToThreeJS(originalType string) string {
	if c == nil || c.ThreeJS == nil {
		return ""
	}

	for threeJSType, originalTypes := range c.ThreeJS {
		for _, t := range originalTypes {
			if t == originalType {
				return threeJSType
			}
		}
	}
	return ""
}

// GetDisplayName 获取显示名称
func (c *TextureMappingConfig) GetDisplayName(threeJSType string) string {
	if c == nil || c.DisplayNames == nil {
		return threeJSType
	}

	if name, ok := c.DisplayNames[threeJSType]; ok {
		return name
	}
	return threeJSType
}

// GetThreeJSTypeInfo 获取 Three.js 类型的详细信息
func (c *TextureMappingConfig) GetThreeJSTypeInfo() []map[string]interface{} {
	if c == nil || c.ThreeJS == nil {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, 0, len(c.ThreeJS))
	for threeJSType, originalTypes := range c.ThreeJS {
		result = append(result, map[string]interface{}{
			"type":          threeJSType,
			"display_name":  c.GetDisplayName(threeJSType),
			"original_types": originalTypes,
			"count":         len(originalTypes),
		})
	}
	return result
}
