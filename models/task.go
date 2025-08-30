package models

import "time"

// Task 抓取任务模型
type Task struct {
	ID         string    `json:"id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	PageNo     int       `json:"page_no"`
	Position   int       `json:"position"`
	Status     string    `json:"status"` // pending, running, completed, failed
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TaskStatus 任务状态常量
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// CrawlerStats 爬虫统计信息
type CrawlerStats struct {
	TotalTasks     int           `json:"total_tasks"`
	CompletedTasks int           `json:"completed_tasks"`
	FailedTasks    int           `json:"failed_tasks"`
	TotalArticles  int           `json:"total_articles"`
	StartTime      time.Time     `json:"start_time"`
	Duration       time.Duration `json:"duration"`
	ArticlesPerSec float64       `json:"articles_per_sec"`
}
