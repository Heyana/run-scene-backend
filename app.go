package main

import (
	"context"
	"fmt"
	"go_wails_project_manager/bootstrap"

	"github.com/wailsapp/wails/v2/pkg/runtime" // Wails的runtime包
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	// 在goroutine中启动服务器，防止阻塞Wails GUI初始化
	// 并确保任何panic不会影响主应用
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("服务器启动过程中发生panic: %v\n", r)
			}
		}()
		bootstrap.RunServer()
	}()
	
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenURL 使用系统默认浏览器打开URL
func (a *App) OpenURL(url string) {
	// 使用Wails内置的runtime包打开浏览器
	runtime.BrowserOpenURL(a.ctx, url)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
