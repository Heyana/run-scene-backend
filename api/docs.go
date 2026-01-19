// Package api 提供 REST API 控制器
package api

import (
	"net/http"
	"os"

	_ "go_wails_project_manager/docs" // 导入自动生成的swagger文档

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupAPIDocs 配置API文档路由
func SetupAPIDocs(router *gin.Engine) {
	// 检查模板目录是否存在
	templatesPath := "api/templates"
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		// 如果模板目录不存在，尝试创建
		if err := os.MkdirAll(templatesPath, 0755); err != nil {
			// 创建目录失败，禁用模板功能
			router.GET("/api/docs", func(c *gin.Context) {
				c.String(http.StatusOK, "API文档功能在打包模式下不可用")
			})
			return
		}

		// 创建基本的索引模板文件
		indexHTML := `<!DOCTYPE html>
<html>
<head>
    <title>API文档中心</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>API文档中心</h1>
    <p>在开发模式下提供完整功能。</p>
</body>
</html>`

		if err := os.WriteFile(templatesPath+"/docs_index.html", []byte(indexHTML), 0644); err != nil {
			// 写入模板文件失败，禁用模板功能
			router.GET("/api/docs", func(c *gin.Context) {
				c.String(http.StatusOK, "API文档功能在打包模式下不可用")
			})
			return
		}
	}

	// 加载HTML模板
	router.LoadHTMLGlob("api/templates/*.html")

	// 提供静态资源
	router.Static("/api/docs/assets", "./api/assets")

	// 检查并映射文档文件 - 优先使用YAML格式
	var specURL string

	// 尝试不同的文档文件格式
	if _, err := os.Stat("docs/swagger.yaml"); err == nil {
		specURL = "/api/docs/swagger.yaml"
	} else if _, err := os.Stat("docs/swagger.json"); err == nil {
		specURL = "/api/docs/swagger.json"
	} else if _, err := os.Stat("docs/doc.json"); err == nil {
		specURL = "/api/docs/doc.json"
	} else {
		// 如果没有找到任何文档文件，创建一个基本的
		specURL = "/api/docs/swagger.json"
	}

	// 将文档文件映射到统一路径
	router.StaticFile("/api/docs/swagger.yaml", "docs/swagger.yaml")
	router.StaticFile("/api/docs/swagger.json", "docs/swagger.json")
	router.StaticFile("/api/docs/doc.json", "docs/doc.json")

	// 1. 标准Swagger UI - 修复静态资源路径问题
	router.GET("/api/docs/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/api/docs/swagger/index.html")
	})

	// 使用正确的ginSwagger配置，确保静态资源能正确加载
	router.GET("/api/docs/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL(specURL),
		ginSwagger.DefaultModelsExpandDepth(-1),
		ginSwagger.DocExpansion("none"),
		ginSwagger.DeepLinking(true),
	))

	// 2. ReDoc - 现代简洁风格
	router.GET("/api/docs/redoc", func(c *gin.Context) {
		c.HTML(http.StatusOK, "redoc.html", gin.H{
			"title":   "API文档 - ReDoc",
			"specURL": specURL,
		})
	})

	// 3. RapiDoc - 灵活现代风格
	router.GET("/api/docs/rapidoc", func(c *gin.Context) {
		c.HTML(http.StatusOK, "rapidoc.html", gin.H{
			"title":   "API文档 - RapiDoc",
			"specURL": specURL,
		})
	})

	// 4. Stoplight Elements - 专业级风格
	router.GET("/api/docs/elements", func(c *gin.Context) {
		c.HTML(http.StatusOK, "elements.html", gin.H{
			"title":   "API文档 - Elements",
			"specURL": specURL,
		})
	})

	// 5. 文档主页 - 提供各种文档界面的入口
	router.GET("/api/docs", func(c *gin.Context) {
		c.HTML(http.StatusOK, "docs_index.html", gin.H{
			"title": "API文档中心",
		})
	})
}
