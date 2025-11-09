package dao

import "database/sql"

// TaskType 任务类型
type TaskType string

const (
	TaskTypeFunction TaskType = "function"
	TaskTypeHTTP     TaskType = "http"
	TaskTypeGRPC     TaskType = "grpc"
)

// JobStatus 任务状态
type JobStatus string

const (
	// JobStatusActive 启用
	JobStatusActive JobStatus = "active"
	// JobStatusRunning 运行中
	JobStatusRunning JobStatus = "running"
	// JobStatusPaused 暂停
	JobStatusPaused JobStatus = "paused"
	// JobStatusDeleted 删除
	JobStatusDeleted JobStatus = "deleted"
)

// CronJob 定时任务
type CronJob struct {
	ID     int64 `json:"id" gorm:"primaryKey, autoIncrement"`
	CronId int64 `json:"cron_id" gorm:"unique"`
	// 任务名
	Name string `gorm:"column:cron_name;type:varchar(128);size:128"`
	// 任务描述
	Description sql.NullString `gorm:"column:description;type=varchar(4096);size:4096"`
	// 任务执行表达式
	CronExpr string `gorm:"column:cron_expr"`
	// 任务类型
	TaskType TaskType `gorm:"column:task_type"`

	// 任务状态
	Status JobStatus `gorm:"column:status;type:varchar(128);size:128"`
	// 任务最大重试次数
	MaxRetry int `gorm:"column:max_retry"`
	// 任务超时时间(秒)
	Timeout int `gorm:"column:timeout"`

	Ctime float64
	Utime float64
}
