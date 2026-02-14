package api

import (
	"go_wails_project_manager/config"
	"go_wails_project_manager/controllers/requirement"
	"go_wails_project_manager/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRequirementRoutes 设置需求管理路由
func SetupRequirementRoutes(router *gin.Engine, jwtAuth *middleware.JWTAuth, db *gorm.DB) {
	// 检查功能是否启用
	if !config.IsRequirementEnabled() {
		return
	}

	// 创建控制器
	companyController := requirement.NewCompanyController()
	projectController := requirement.NewProjectController()
	missionListController := requirement.NewMissionListController()
	missionController := requirement.NewMissionController()

	// 需求管理路由组
	requirementGroup := router.Group("/api/requirement")
	requirementGroup.Use(jwtAuth.AuthMiddleware()) // 需要认证
	{
		// 公司管理
		companies := requirementGroup.Group("/companies")
		{
			companies.GET("", companyController.List)
			companies.POST("", companyController.Create)
			companies.GET("/:id", companyController.GetDetail)
			companies.PUT("/:id", companyController.Update)
			companies.POST("/:id/members", companyController.AddMember)
			companies.DELETE("/:id/members/:user_id", companyController.RemoveMember)
			companies.GET("/:id/members", companyController.GetMembers)
		}

		// 项目管理
		projects := requirementGroup.Group("/projects")
		{
			projects.GET("", projectController.List)
			projects.POST("", projectController.Create)
			projects.GET("/:id", projectController.GetDetail)
			projects.PUT("/:id", projectController.Update)
			projects.POST("/:id/members", projectController.AddMember)
			projects.DELETE("/:id/members/:user_id", projectController.RemoveMember)
			projects.GET("/:id/members", projectController.GetMembers)
			projects.GET("/:id/statistics", projectController.GetStatistics)
		}

		// 任务列表管理（看板列）
		missionLists := requirementGroup.Group("/mission-lists")
		{
			missionLists.GET("", missionListController.List)
			missionLists.POST("", missionListController.Create)
			missionLists.GET("/:id", missionListController.GetDetail)
			missionLists.PUT("/:id", missionListController.Update)
			missionLists.DELETE("/:id", missionListController.Delete)
		}

		// 任务管理
		missions := requirementGroup.Group("/missions")
		{
			missions.GET("", missionController.List)
			missions.POST("", missionController.Create)
			missions.GET("/:id", missionController.GetDetail)
			missions.PUT("/:id", missionController.Update)
			missions.DELETE("/:id", missionController.Delete)
			missions.PATCH("/:id/status", missionController.UpdateStatus)
			missions.POST("/:id/comments", missionController.AddComment)
		}
	}
}
