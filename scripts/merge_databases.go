package main

import (
	"fmt"

	"go_wails_project_manager/config"
	"go_wails_project_manager/database/migrations"
	"go_wails_project_manager/models/ai3d"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("AI3D数据库合并工具")
	fmt.Println("========================================\n")

	// 初始化配置
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		return
	}

	// 打开服务器数据库
	serverDB, err := gorm.Open(sqlite.Open("./deploy/data/app.db"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		fmt.Printf("打开服务器数据库失败: %v\n", err)
		return
	}
	fmt.Println("✓ 已连接服务器数据库: deploy/data/app.db")

	// 打开本地数据库
	localDB, err := gorm.Open(sqlite.Open("./data/app.db"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		fmt.Printf("打开本地数据库失败: %v\n", err)
		return
	}
	fmt.Println("✓ 已连接本地数据库: data/app.db\n")

	// 1. 在服务器数据库执行迁移（创建ai3d_tasks表并迁移数据）
	fmt.Println("步骤1: 在服务器数据库执行迁移...")
	if err := migrations.MergeAI3DTasks(serverDB); err != nil {
		fmt.Printf("❌ 服务器数据库迁移失败: %v\n", err)
		return
	}
	fmt.Println("✓ 服务器数据库迁移完成\n")

	// 2. 统计服务器数据库的任务数
	var serverCount int64
	serverDB.Model(&ai3d.Task{}).Count(&serverCount)
	fmt.Printf("服务器数据库现有任务数: %d\n", serverCount)

	// 3. 从本地数据库读取所有ai3d任务
	var localTasks []ai3d.Task
	if err := localDB.Find(&localTasks).Error; err != nil {
		fmt.Printf("❌ 读取本地任务失败: %v\n", err)
		return
	}
	fmt.Printf("本地数据库任务数: %d\n\n", len(localTasks))

	// 4. 检查哪些任务需要合并（避免重复）
	fmt.Println("步骤2: 检查需要合并的任务...")
	var tasksToMerge []ai3d.Task
	
	for _, localTask := range localTasks {
		// 检查服务器数据库是否已存在相同的任务
		var existingTask ai3d.Task
		err := serverDB.Where("provider = ? AND provider_task_id = ?", 
			localTask.Provider, localTask.ProviderTaskID).First(&existingTask).Error
		
		if err == gorm.ErrRecordNotFound {
			// 不存在，需要合并
			tasksToMerge = append(tasksToMerge, localTask)
			fmt.Printf("  + 需要合并: ID=%d, Provider=%s, TaskID=%s\n", 
				localTask.ID, localTask.Provider, localTask.ProviderTaskID)
		} else if err != nil {
			fmt.Printf("  ⚠ 检查任务失败: %v\n", err)
		} else {
			fmt.Printf("  - 已存在，跳过: Provider=%s, TaskID=%s\n", 
				localTask.Provider, localTask.ProviderTaskID)
		}
	}

	if len(tasksToMerge) == 0 {
		fmt.Println("\n没有需要合并的任务")
		return
	}

	fmt.Printf("\n找到 %d 个需要合并的任务\n\n", len(tasksToMerge))

	// 5. 合并任务到服务器数据库
	fmt.Println("步骤3: 开始合并任务...")
	success := 0
	failed := 0

	for i, task := range tasksToMerge {
		fmt.Printf("[%d/%d] 合并任务: Provider=%s, TaskID=%s\n", 
			i+1, len(tasksToMerge), task.Provider, task.ProviderTaskID)
		
		// 清空ID，让数据库自动生成新ID
		task.ID = 0
		
		if err := serverDB.Create(&task).Error; err != nil {
			fmt.Printf("  ❌ 失败: %v\n", err)
			failed++
		} else {
			fmt.Printf("  ✓ 成功，新ID=%d\n", task.ID)
			success++
		}
	}

	// 6. 最终统计
	fmt.Println("\n========================================")
	fmt.Println("合并完成！")
	fmt.Println("========================================")
	
	var finalCount int64
	serverDB.Model(&ai3d.Task{}).Count(&finalCount)
	
	fmt.Printf("服务器数据库最终任务数: %d\n", finalCount)
	fmt.Printf("成功合并: %d 个\n", success)
	fmt.Printf("失败: %d 个\n", failed)
	fmt.Println("========================================")

	// 7. 显示合并后的任务列表
	fmt.Println("\n服务器数据库任务列表:")
	var allTasks []ai3d.Task
	serverDB.Order("created_at DESC").Find(&allTasks)
	
	for _, task := range allTasks {
		status := "✓"
		if task.Status != "DONE" {
			status = "⏳"
		}
		fmt.Printf("  %s ID=%d, Provider=%s, Status=%s, Name=%s\n", 
			status, task.ID, task.Provider, task.Status, task.Name)
	}
}
