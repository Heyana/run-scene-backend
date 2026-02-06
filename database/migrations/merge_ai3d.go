package migrations

import (
	"encoding/json"
	"fmt"
	"time"

	"go_wails_project_manager/models/ai3d"
	hunyuanModel "go_wails_project_manager/models/hunyuan"
	meshyModel "go_wails_project_manager/models/meshy"

	"gorm.io/gorm"
)

// MergeAI3DTasks 合并AI3D任务表
func MergeAI3DTasks(db *gorm.DB) error {
	fmt.Println("开始合并AI3D任务表...")

	// 1. 创建新表
	fmt.Println("1. 创建 ai3d_tasks 表...")
	if err := db.AutoMigrate(&ai3d.Task{}); err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	// 2. 检查是否已经迁移过
	var count int64
	db.Model(&ai3d.Task{}).Count(&count)
	if count > 0 {
		fmt.Printf("ai3d_tasks 表已有 %d 条数据，跳过迁移\n", count)
		return nil
	}

	// 3. 检查旧表是否存在
	hasHunyuan := db.Migrator().HasTable("hunyuan_tasks")
	hasMeshy := db.Migrator().HasTable("meshy_tasks")

	if !hasHunyuan && !hasMeshy {
		fmt.Println("未发现旧表，这是全新安装，跳过数据迁移")
		return nil
	}

	// 4. 迁移混元任务
	if hasHunyuan {
		fmt.Println("2. 迁移混元任务...")
		var hunyuanTasks []hunyuanModel.HunyuanTask
		if err := db.Table("hunyuan_tasks").Find(&hunyuanTasks).Error; err != nil {
			fmt.Printf("查询混元任务失败: %v\n", err)
		} else {
			for _, ht := range hunyuanTasks {
				task := convertHunyuanTask(&ht)
				if err := db.Create(task).Error; err != nil {
					fmt.Printf("迁移混元任务 %d 失败: %v\n", ht.ID, err)
				}
			}
			fmt.Printf("已迁移 %d 个混元任务\n", len(hunyuanTasks))
		}
	}

	// 5. 迁移Meshy任务
	if hasMeshy {
		fmt.Println("3. 迁移Meshy任务...")
		var meshyTasks []meshyModel.MeshyTask
		if err := db.Table("meshy_tasks").Find(&meshyTasks).Error; err != nil {
			fmt.Printf("查询Meshy任务失败: %v\n", err)
		} else {
			for _, mt := range meshyTasks {
				task := convertMeshyTask(&mt)
				if err := db.Create(task).Error; err != nil {
					fmt.Printf("迁移Meshy任务 %d 失败: %v\n", mt.ID, err)
				}
			}
			fmt.Printf("已迁移 %d 个Meshy任务\n", len(meshyTasks))
		}
	}

	// 6. 备份旧表（重命名）
	fmt.Println("4. 备份旧表...")
	timestamp := time.Now().Format("20060102_150405")
	
	if hasHunyuan {
		backupName := fmt.Sprintf("hunyuan_tasks_backup_%s", timestamp)
		if err := db.Migrator().RenameTable("hunyuan_tasks", backupName); err != nil {
			fmt.Printf("备份混元表失败: %v\n", err)
		} else {
			fmt.Printf("混元表已备份为: %s\n", backupName)
		}
	}
	
	if hasMeshy {
		backupName := fmt.Sprintf("meshy_tasks_backup_%s", timestamp)
		if err := db.Migrator().RenameTable("meshy_tasks", backupName); err != nil {
			fmt.Printf("备份Meshy表失败: %v\n", err)
		} else {
			fmt.Printf("Meshy表已备份为: %s\n", backupName)
		}
	}

	fmt.Println("AI3D任务表合并完成！")
	return nil
}

// convertHunyuanTask 转换混元任务
func convertHunyuanTask(ht *hunyuanModel.HunyuanTask) *ai3d.Task {
	task := &ai3d.Task{
		CreatedAt:      ht.CreatedAt,
		UpdatedAt:      ht.UpdatedAt,
		Provider:       "hunyuan",
		ProviderTaskID: ht.JobID,
		Status:         ht.Status,
		Progress:       100, // 混元任务没有进度字段，默认100
		InputType:      ht.InputType,
		Prompt:         ht.Prompt,
		// ImageURL 字段已删除，不再保存
		Name:          ht.Name,
		Description:   ht.Description,
		Category:      ht.Category,
		Tags:          ht.Tags,
		CreatedBy:     ht.CreatedBy,
		CreatedIP:     ht.CreatedIP,
		LocalPath:     ht.LocalPath,
		NASPath:       ht.NASPath,
		ThumbnailPath: ht.ThumbnailPath,
		FileSize:      ht.FileSize,
		FileHash:      ht.FileHash,
		ErrorCode:     ht.ErrorCode,
		ErrorMessage:  ht.ErrorMessage,
	}

	// 构建生成参数
	params := ai3d.GenerationParams{
		"model":        ht.Model,
		"faceCount":    ht.FaceCount,
		"generateType": ht.GenerateType,
		"enablePbr":    ht.EnablePBR,
		"resultFormat": ht.ResultFormat,
	}
	task.GenerationParams = params

	return task
}

// convertMeshyTask 转换Meshy任务
func convertMeshyTask(mt *meshyModel.MeshyTask) *ai3d.Task {
	// 转换Meshy状态为统一状态
	status := mt.Status
	switch mt.Status {
	case "PENDING":
		status = "WAIT"
	case "IN_PROGRESS":
		status = "RUN"
	case "SUCCEEDED":
		status = "DONE"
	case "FAILED", "CANCELED":
		status = "FAIL"
	}

	task := &ai3d.Task{
		CreatedAt:      mt.CreatedAt,
		UpdatedAt:      mt.UpdatedAt,
		Provider:       "meshy",
		ProviderTaskID: mt.TaskID,
		Status:         status,
		Progress:       mt.Progress,
		InputType:      "image",
		Name:           mt.Name,
		Description:    mt.Description,
		Category:       mt.Category,
		CreatedBy:      mt.CreatedBy,
		CreatedIP:      mt.CreatedIP,
	}

	// ImageURL 字段已删除，不再保存

	// 设置文件路径
	if mt.LocalPath != "" {
		task.LocalPath = &mt.LocalPath
	}
	if mt.NASPath != "" {
		task.NASPath = &mt.NASPath
	}
	if mt.ThumbnailPath != "" {
		task.ThumbnailPath = &mt.ThumbnailPath
	}
	if mt.FileSize > 0 {
		task.FileSize = &mt.FileSize
	}
	if mt.FileHash != "" {
		task.FileHash = &mt.FileHash
	}
	if mt.ModelURL != "" {
		task.ModelURL = &mt.ModelURL
	}
	if mt.ThumbnailURL != "" {
		task.ThumbnailURL = &mt.ThumbnailURL
	}
	if mt.ErrorMessage != "" {
		task.ErrorMessage = &mt.ErrorMessage
	}

	// 构建生成参数
	params := ai3d.GenerationParams{
		"enablePbr":       mt.EnablePBR,
		"shouldRemesh":    mt.ShouldRemesh,
		"shouldTexture":   mt.ShouldTexture,
		"savePreRemeshed": mt.SavePreRemeshed,
	}
	
	// 解析Tags
	if mt.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(mt.Tags), &tags); err == nil {
			task.Tags = &mt.Tags
		}
	}
	
	task.GenerationParams = params

	return task
}
