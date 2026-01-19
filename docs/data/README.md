# Apifox API测试导入指南

本目录包含了智慧工厂管理系统API的测试数据和导入工具，可用于快速导入至Apifox进行测试。

## 文件说明

1. **sample_data.json** - 产品和组件的示例数据
2. **api_collection.json** - Apifox可导入的API接口集合
3. **import_script.js** - 批量导入测试数据的Node.js脚本(可用bun运行)

## 导入步骤

### 导入API集合到Apifox

1. 打开Apifox客户端
2. 创建新项目或选择现有项目
3. 点击左侧菜单的"导入/导出" > "导入数据"
4. 选择"导入项目数据(Apifox格式)"
5. 选择 **api_collection.json** 文件
6. 点击"导入"完成操作

### 导入Swagger文档到Apifox

API文档也可以通过Swagger直接导入：

1. 在Apifox中，点击"导入" > "导入OpenAPI(Swagger)"
2. 输入URL: `http://localhost:8080/api/docs/swagger.json` 
3. 或者下载swagger.json后导入本地文件

### 导入测试数据到系统

使用Node.js/bun脚本自动导入测试数据:

1. 确保服务器已启动 (`air` 或 `go run -tags dev ./dev/dev.go ./dev/dev_server.go`)
2. 使用bun运行脚本:
   ```bash
   bun run import_script.js
   ```
   
   如果使用Node.js:
   ```bash
   node import_script.js
   ```

脚本会自动导入产品和组件的示例数据到系统中。

## 手动测试步骤

如果您想手动测试API，可按以下步骤操作：

1. 在Apifox中设置环境变量:
   - 添加变量 `baseUrl` = `http://localhost:8080`

2. 产品测试流程:
   - 创建产品 (POST `/api/products`)
   - 获取产品列表 (GET `/api/products`)
   - 获取产品详情 (GET `/api/products/{id}`)
   - 更新产品信息 (PUT `/api/products/{id}`)
   - 删除产品 (DELETE `/api/products/{id}`)

3. 组件测试流程:
   - 创建组件 (POST `/api/products/{id}/components`)
   - 获取组件列表 (GET `/api/products/{id}/components`)
   - 获取组件详情 (GET `/api/products/{id}/components/{componentId}`)
   - 更新组件 (PUT `/api/products/{id}/components/{componentId}`)
   - 删除组件 (DELETE `/api/products/{id}/components/{componentId}`)

## 直接访问API文档

系统内置了Swagger文档，可通过浏览器访问：

```
http://localhost:8080/api/docs
``` 