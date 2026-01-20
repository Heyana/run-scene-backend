# AmbientCG API 测试结果

## ✅ 测试时间

2026-01-20

## ✅ 测试结论

**所有 API 端点测试通过，可以正常爬取资源！**

---

## 1️⃣ 获取材质列表 ✅

### API 端点

```
GET https://ambientcg.com/api/v2/full_json?type=Material&sort=Popular&limit=3&offset=0
```

### 测试命令

```powershell
curl "https://ambientcg.com/api/v2/full_json?type=Material&sort=Popular&limit=3&offset=0" -o "test_list.json"
```

### 返回数据结构

```json
{
  "numberOfResults": 1957,  // 总共 1957 个材质
  "nextPageHttp": "...",     // 下一页链接
  "foundAssets": [
    {
      "assetId": "Ground103",
      "displayName": "Ground 103",
      "displayCategory": "Ground",
      "downloadCount": 1215,
      "releaseDate": "2026-01-12 17:00:00",
      "tags": ["brown", "dirt", "earth", "ground"],
      "maps": ["color", "displacement", "normal", "roughness", "ambient-occlusion"],
      "previewImage": {
        "256-PNG": "https://...",
        "512-PNG": "https://...",
        ...
      },
      "downloadFolders": null  // 列表接口不返回下载链接
    }
  ]
}
```

### 关键信息

- ✅ 总材质数：**1957 个**
- ✅ 支持分页：limit + offset
- ✅ 支持排序：Popular, Latest, Alphabet, Downloads
- ✅ 包含预览图：多种尺寸（64px - 2048px）
- ✅ 包含标签和分类信息

---

## 2️⃣ 获取单个材质详情（含下载链接）✅

### API 端点

```
GET https://ambientcg.com/api/v2/full_json?id=Ground103&include=downloadData
```

### 测试命令

```powershell
curl "https://ambientcg.com/api/v2/full_json?id=Ground103&include=downloadData" -o "test_detail.json"
```

### 下载信息结构

```json
{
  "downloadFolders": {
    "default": {
      "downloadFiletypeCategories": {
        "zip": {
          "downloads": [
            {
              "downloadLink": "https://ambientcg.com/get?file=Ground103_1K-JPG.zip",
              "fileName": "Ground103_1K-JPG.zip",
              "size": 9596711, // 9.15 MB
              "attribute": "1K-JPG"
            },
            {
              "downloadLink": "https://ambientcg.com/get?file=Ground103_2K-JPG.zip",
              "size": 33796068, // 32.2 MB
              "attribute": "2K-JPG"
            },
            {
              "downloadLink": "https://ambientcg.com/get?file=Ground103_4K-JPG.zip",
              "size": 125852705, // 120 MB
              "attribute": "4K-JPG"
            },
            {
              "downloadLink": "https://ambientcg.com/get?file=Ground103_8K-JPG.zip",
              "size": 482445282, // 460 MB
              "attribute": "8K-JPG"
            }
          ]
        }
      }
    }
  }
}
```

### 关键信息

- ✅ 提供多种分辨率：1K, 2K, 4K, 8K
- ✅ 提供多种格式：JPG, PNG
- ✅ 包含文件大小信息
- ✅ 直接下载链接可用

---

## 3️⃣ 下载材质包 ✅

### 测试命令

```powershell
Invoke-WebRequest -Uri "https://ambientcg.com/get?file=Ground103_1K-JPG.zip" -OutFile "test_download.zip"
```

### 下载结果

- ✅ 文件大小：9,596,711 字节（9.15 MB）
- ✅ 下载速度：正常
- ✅ 文件完整性：正常

---

## 4️⃣ 解压材质包 ✅

### 解压命令

```powershell
Expand-Archive -Path "test_download.zip" -DestinationPath "test_extract"
```

### 解压后文件列表

```
Ground103_1K-JPG_AmbientOcclusion.jpg  // AO 贴图
Ground103_1K-JPG_Color.jpg             // 颜色贴图（Diffuse）
Ground103_1K-JPG_Displacement.jpg      // 位移贴图
Ground103_1K-JPG_NormalDX.jpg          // 法线贴图（DirectX）
Ground103_1K-JPG_NormalGL.jpg          // 法线贴图（OpenGL）
Ground103_1K-JPG_Roughness.jpg         // 粗糙度贴图
Ground103_1K-JPG.blend                 // Blender 材质文件
Ground103_1K-JPG.mtlx                  // MaterialX 文件
Ground103_1K-JPG.tres                  // Godot 材质文件
Ground103_1K-JPG.usdc                  // USD 文件
Ground103.png                          // 预览图
```

### 文件命名规则

```
{AssetID}_{Resolution}-{Format}_{MapType}.jpg
```

例如：`Ground103_1K-JPG_Color.jpg`

---

## 📊 数据映射到现有模型

### Texture 表字段映射

| 现有字段      | AmbientCG 字段        | 映射方式                   |
| ------------- | --------------------- | -------------------------- |
| AssetID       | assetId               | 直接映射                   |
| Name          | displayName           | 直接映射                   |
| Description   | description           | 直接映射（通常为空）       |
| Type          | displayCategory       | 需要建立映射表             |
| Authors       | -                     | 固定为 "AmbientCG"         |
| MaxResolution | downloads[].attribute | 提取最大值（8K）           |
| FilesHash     | assetId               | 使用 assetId 作为唯一标识  |
| DatePublished | releaseDate           | 时间格式转换               |
| DownloadCount | downloadCount         | 直接映射                   |
| TextureTypes  | maps[]                | 数组转逗号分隔字符串       |
| Source        | -                     | 新增字段，值为 "ambientcg" |

### TextureFile 表字段映射

从 ZIP 包中的文件名解析：

- `TextureID`: 关联的 Texture ID
- `MapType`: 从文件名提取（Color, Normal, Roughness, etc.）
- `Resolution`: 从文件名提取（1K, 2K, 4K, 8K）
- `FileID`: 关联到 File 表

---

## 🎯 爬取策略建议

### 方案 1：全量爬取（推荐用于初始化）

```
1. 获取材质列表（分页，每页 100 个）
   GET /api/v2/full_json?type=Material&limit=100&offset=0

2. 遍历每个材质
   - 检查本地是否已存在
   - 如果不存在，获取详细信息

3. 获取下载信息
   GET /api/v2/full_json?id={assetId}&include=downloadData

4. 选择合适的分辨率下载（建议 2K-JPG）
   GET https://ambientcg.com/get?file={fileName}

5. 解压并保存到数据库
```

### 方案 2：增量更新

```
1. 按 Latest 排序获取最新材质
   GET /api/v2/full_json?type=Material&sort=Latest&limit=50

2. 检查 releaseDate，只处理新材质

3. 下载并保存
```

### 方案 3：按需下载

```
1. 用户搜索时，从 API 获取列表
2. 用户选择材质时，才下载
3. 下载后缓存到本地
```

---

## 📝 实现建议

### 1. 创建 AmbientCG 适配器

```go
type AmbientCGAdapter struct {
    baseURL string
    client  *http.Client
}

func (a *AmbientCGAdapter) GetMaterialList(limit, offset int) ([]Material, error)
func (a *AmbientCGAdapter) GetMaterialDetail(assetID string) (*MaterialDetail, error)
func (a *AmbientCGAdapter) DownloadMaterial(downloadLink, savePath string) error
```

### 2. 分类映射表

```go
var categoryMap = map[string]int{
    "Ground":       1,
    "Wood":         2,
    "Grass":        3,
    "PavingStones": 4,
    "Fabric":       5,
    "Concrete":     6,
    "Metal":        7,
    "Brick":        8,
    "Tiles":        9,
    "Rock":         10,
    "Marble":       11,
    "Leather":      12,
    "Plastic":      13,
}
```

### 3. 下载队列优先级

- 1K: 优先级 5（快速预览）
- 2K: 优先级 3（推荐使用）
- 4K: 优先级 7（高质量）
- 8K: 优先级 9（按需下载）

---

## ⚠️ 注意事项

1. **API 限流**：建议添加请求间隔（100-200ms）
2. **下载大小**：8K 材质包可达 1GB+，建议默认下载 2K
3. **存储空间**：1957 个材质 × 32MB ≈ 62GB（2K-JPG）
4. **并发控制**：建议同时下载不超过 3 个
5. **错误重试**：网络错误时自动重试 3 次
6. **断点续传**：大文件下载支持断点续传

---

## 🚀 下一步

1. ✅ API 测试完成
2. ⏭️ 修改 Texture 模型（添加 source 字段）
3. ⏭️ 实现 AmbientCG 适配器
4. ⏭️ 实现爬虫服务
5. ⏭️ 集成到现有同步系统

---

## 测试文件清单

- ✅ `test_list.json` - 材质列表（3 个示例）
- ✅ `test_detail.json` - 单个材质详情（Ground103）
- ✅ `test_download.zip` - 下载的材质包（9.15 MB）
- ✅ `test_extract/` - 解压后的文件（6 个贴图 + 5 个材质文件）
