// Package task 任务队列
package task

import (
	"go_wails_project_manager/models"
	"sync"
)

// TaskQueue 任务队列（优先级队列）
type TaskQueue struct {
	items []*models.Task
	mu    sync.Mutex
}

// NewTaskQueue 创建任务队列
func NewTaskQueue() *TaskQueue {
	return &TaskQueue{
		items: make([]*models.Task, 0),
	}
}

// Push 添加任务
func (q *TaskQueue) Push(task *models.Task) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, task)
	q.sort()
}

// Pop 取出任务（按优先级）
func (q *TaskQueue) Pop() *models.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	task := q.items[0]
	q.items = q.items[1:]
	return task
}

// Peek 查看队首任务
func (q *TaskQueue) Peek() *models.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return nil
	}

	return q.items[0]
}

// Remove 移除任务
func (q *TaskQueue) Remove(taskID uint) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, task := range q.items {
		if task.ID == taskID {
			q.items = append(q.items[:i], q.items[i+1:]...)
			return true
		}
	}

	return false
}

// Len 队列长度
func (q *TaskQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}

// sort 按优先级排序（优先级高的在前）
func (q *TaskQueue) sort() {
	// 简单的冒泡排序
	for i := 0; i < len(q.items); i++ {
		for j := i + 1; j < len(q.items); j++ {
			if q.items[i].Priority < q.items[j].Priority {
				q.items[i], q.items[j] = q.items[j], q.items[i]
			}
		}
	}
}
