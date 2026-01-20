# AmbientCG 材质库接入分析

## 网站信息

- 网站：https://ambientcg.com
- API 文档：https://docs.ambientcg.com/api/v2/full_json/
- 材质总数：约 1957 个（根据 JSON 数据）

## API 端点

### 1. 获取材质列表

```
GET https://ambientcg.com/api/v2/full_json?type=Material&sort=Popular&limit=100&offset=0&include=statisticsData,labelData,previewData,technicalData
```

### 2. 获取材质详细信息（包含下载链接）

```
GET https://ambientcg.com/api/v2/full_json?id={assetId}&include=downloadData
```

## 数据结构对比

### AmbientCG 数据结构

```json
{
  "assetId": "Ground103",
  "releaseDate": "2026-01-12 17:00:00",
  "dataType": "Material",
  "creationMethod": "PBRPhotogrammetry",
  "downloadCount": 1215,
  "tags": ["brown", "dirt", "earth", "ground"],
  "displayName": "Ground 103",
  "description": "",
  "displayCategory": "Ground",
  "maps": ["color", "displacement", "normal", "roughness", "ambient-occlusion"],
  "previewImage": {
    "256-PNG": "https://...",
    "512-PNG": "https://...",
    ...
  },
  "downloadFolders": null  // 需要单独请求获取
}
```

### 现有数据模型（texture.go）

```go
type Texture struct {
    AssetID           string     // ✅ 对应 assetId
    Name              string     // ✅ 对应 displayName
    Description       string     // ✅ 对应 description
    Type              int        // ✅ 可映射 displayCategory
    Authors           string     // ⚠️ AmbientCG 无此字段
    MaxResolution     string     // ⚠️ 需要从下载文件中提取
    FilesHash         string     // ✅ 可用 assetId 作为唯一标识
    DatePublished     int64      // ✅ 对应 releaseDate（需转换）
    DownloadCount     int        // ✅ 对应 downloadCount
    TextureTypes      string     // ✅ 对应 maps 数组（逗号分隔）
}
```

## 兼容性评估

### ✅ 完全兼容的字段

1. `AssetID` ← `assetId`
2. `Name` ← `displayName`
3. `Description` ← `description`
4. `DownloadCount` ← `downloadCount`
5. `DatePublished` ← `releaseDate`（需时间转换）
6. `TextureTypes` ← `maps`（数组转逗号分隔字符串）

### ⚠️ 需要适配的字段

1. `Type` - 需要建立分类映射表
   - Ground → 1
   - Wood → 2
   - Grass → 3
   - PavingStones → 4
   - Fabric → 5
   - Concrete → 6
   - Metal → 7
   - 等等...

2. `Authors` - AmbientCG 无作者信息，可设置为 "AmbientCG"

3. `MaxResolution` - 需要从下载文件列表中提取最大分辨率

4. `FilesHash` - 可以使用 `assetId` 或者基于文件列表生成 hash

### 📦 下载文件信息

需要第二次 API 调用获取：

```
GET https://ambientcg.com/api/v2/full_json?id=Ground103&include=downloadData
```

返回的 `downloadFolders` 结构（预期）：

```json
{
  "downloadFolders": {
    "default": {
      "zipFileSize": 123456,
      "downloadLink": "https://...",
      "files": [
        {
          "fileName": "Ground103_1K_Color.jpg",
          "fileSize": 12345,
          "resolution": "1K",
          "mapType": "Color"
        }
      ]
    }
  }
}
```

## 实现建议

### 方案 1：完全兼容（推荐）

**优点**：复用现有代码，改动最小
**实现**：

1. 创建 AmbientCG 适配器（adapter）
2. 将 AmbientCG 数据转换为现有 Texture 模型
3. 添加 `source` 字段区分数据来源（polyhaven/ambientcg）

```go
type Texture struct {
    // ... 现有字段
    Source string `gorm:"size:20;index" json:"source"` // 新增：数据来源
}
```

### 方案 2：扩展模型

**优点**：保留更多原始信息
**实现**：

1. 添加 JSON 字段存储原始数据
2. 保持核心字段兼容

```go
type Texture struct {
    // ... 现有字段
    Source     string `gorm:"size:20;index" json:"source"`
    RawData    string `gorm:"type:text" json:"raw_data"` // 存储原始 JSON
}
```

## 同步流程

### 1. 获取材质列表

```
GET /api/v2/full_json?type=Material&sort=Latest&limit=100&offset=0
```

### 2. 遍历每个材质

- 检查本地是否已存在（通过 assetId）
- 如果不存在或需要更新，获取详细信息

### 3. 获取下载信息

```
GET /api/v2/full_json?id={assetId}&include=downloadData
```

### 4. 解析并保存

- 转换数据格式
- 保存到 Texture 表
- 保存文件信息到 TextureFile 表
- 添加到下载队列

## 分类映射表建议

| AmbientCG Category | Type ID | 中文名称 |
| ------------------ | ------- | -------- |
| Ground             | 1       | 地面     |
| Wood               | 2       | 木材     |
| Grass              | 3       | 草地     |
| PavingStones       | 4       | 铺路石   |
| Fabric             | 5       | 织物     |
| Concrete           | 6       | 混凝土   |
| Metal              | 7       | 金属     |
| Brick              | 8       | 砖块     |
| Tiles              | 9       | 瓷砖     |
| Rock               | 10      | 岩石     |
| Marble             | 11      | 大理石   |
| Leather            | 12      | 皮革     |
| Plastic            | 13      | 塑料     |

## 结论

✅ **现有数据模型完全可以支持 AmbientCG**

只需要：

1. 添加 `source` 字段区分数据来源
2. 创建分类映射表
3. 实现数据转换适配器
4. 调整同步逻辑支持两个数据源

核心表结构无需大改，可以平滑接入！
