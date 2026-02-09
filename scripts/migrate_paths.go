package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ProjectVersion 项目版本表
type ProjectVersion struct {
	ID            uint   `gorm:"primaryKey"`
	FilePath      string `gorm:"size:512"`
	ExtractedPath string `gorm:"size:512"`
	HistoryPath   string `gorm:"size:512"`
	ThumbnailPath string `gorm:"size:512"`
}

// Project 项目表
type Project struct {
	ID            uint   `gorm:"primaryKey"`
	ThumbnailPath string `gorm:"size:512"`
}

func main() {
	// 连接数据库
	dbPath := "data/app.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	log.Printf("开始迁移数据库路径: %s", dbPath)

	// 迁移 project_version 表
	var versions []ProjectVersion
	if err := db.Find(&versions).Error; err != nil {
		log.Fatalf("查询版本记录失败: %v", err)
	}

	log.Printf("找到 %d 条版本记录", len(versions))

	for _, version := range versions {
		updated := false
		updates := make(map[string]interface{})

		// 转换 FilePath
		if version.FilePath != "" && needsConversion(version.FilePath) {
			newPath := extractRelativePath(version.FilePath)
			if newPath != version.FilePath {
				updates["file_path"] = newPath
				updated = true
				log.Printf("版本 %d FilePath: %s -> %s", version.ID, version.FilePath, newPath)
			}
		}

		// 转换 ExtractedPath
		if version.ExtractedPath != "" && needsConversion(version.ExtractedPath) {
			newPath := extractRelativePath(version.ExtractedPath)
			if newPath != version.ExtractedPath {
				updates["extracted_path"] = newPath
				updated = true
				log.Printf("版本 %d ExtractedPath: %s -> %s", version.ID, version.ExtractedPath, newPath)
			}
		}

		// 转换 HistoryPath
		if version.HistoryPath != "" && needsConversion(version.HistoryPath) {
			newPath := extractRelativePath(version.HistoryPath)
			if newPath != version.HistoryPath {
				updates["history_path"] = newPath
				updated = true
				log.Printf("版本 %d HistoryPath: %s -> %s", version.ID, version.HistoryPath, newPath)
			}
		}

		// 转换 ThumbnailPath
		if version.ThumbnailPath != "" && needsConversion(version.ThumbnailPath) {
			newPath := extractRelativePath(version.ThumbnailPath)
			if newPath != version.ThumbnailPath {
				updates["thumbnail_path"] = newPath
				updated = true
				log.Printf("版本 %d ThumbnailPath: %s -> %s", version.ID, version.ThumbnailPath, newPath)
			}
		}

		// 更新数据库
		if updated {
			if err := db.Model(&ProjectVersion{}).Where("id = ?", version.ID).Updates(updates).Error; err != nil {
				log.Printf("更新版本 %d 失败: %v", version.ID, err)
			}
		}
	}

	// 迁移 project 表
	var projects []Project
	if err := db.Find(&projects).Error; err != nil {
		log.Fatalf("查询项目记录失败: %v", err)
	}

	log.Printf("找到 %d 条项目记录", len(projects))

	for _, project := range projects {
		if project.ThumbnailPath != "" && needsConversion(project.ThumbnailPath) {
			newPath := extractRelativePath(project.ThumbnailPath)
			if newPath != project.ThumbnailPath {
				log.Printf("项目 %d ThumbnailPath: %s -> %s", project.ID, project.ThumbnailPath, newPath)
				if err := db.Model(&Project{}).Where("id = ?", project.ID).Update("thumbnail_path", newPath).Error; err != nil {
					log.Printf("更新项目 %d 失败: %v", project.ID, err)
				}
			}
		}
	}

	log.Println("迁移完成！")
}

// needsConversion 判断路径是否需要转换（是否为绝对路径）
func needsConversion(path string) bool {
	// 如果包含 /static/ 或 \static\，说明是绝对路径
	return strings.Contains(path, "/static/") || strings.Contains(path, "\\static\\")
}

// extractRelativePath 从绝对路径提取相对路径（相对于 static 目录）
// 例如: /vol1/1003/project/editor_v2/static/projects/项目名 -> projects/项目名
//      \\192.168.3.10\project\editor_v2\static\projects\项目名 -> projects/项目名
func extractRelativePath(absolutePath string) string {
	// 统一路径分隔符
	path := strings.ReplaceAll(absolutePath, "\\", "/")

	// 查找 static/ 的位置
	staticIdx := strings.LastIndex(path, "/static/")
	if staticIdx == -1 {
		staticIdx = strings.LastIndex(path, "static/")
		if staticIdx != -1 {
			// 返回 static/ 之后的部分
			return path[staticIdx+7:] // 跳过 "static/"
		}
	} else {
		// 返回 static/ 之后的部分
		return path[staticIdx+8:] // 跳过 "/static/"
	}

	// 如果找不到 static/，返回原路径
	return path
}
