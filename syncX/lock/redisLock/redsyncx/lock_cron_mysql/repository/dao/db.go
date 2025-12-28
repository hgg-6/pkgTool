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

// ExecutionStatus 任务执行状态
type ExecutionStatus string

const (
	// ExecutionStatusSuccess 执行成功
	ExecutionStatusSuccess ExecutionStatus = "success"
	// ExecutionStatusFailure 执行失败
	ExecutionStatusFailure ExecutionStatus = "failure"
	// ExecutionStatusRetrying 重试中
	ExecutionStatusRetrying ExecutionStatus = "retrying"
	// ExecutionStatusTimeout 执行超时
	ExecutionStatusTimeout ExecutionStatus = "timeout"
)

// JobHistory 任务执行历史记录
type JobHistory struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	// 任务ID
	CronId int64 `gorm:"column:cron_id;index;not null"`
	// 任务名称
	JobName string `gorm:"column:job_name;type:varchar(128);size:128"`
	// 执行状态
	Status ExecutionStatus `gorm:"column:status;type:varchar(32);size:32;index"`
	// 开始时间
	StartTime int64 `gorm:"column:start_time;not null;index"`
	// 结束时间
	EndTime int64 `gorm:"column:end_time"`
	// 执行时长(毫秒)
	Duration int64 `gorm:"column:duration"`
	// 重试次数
	RetryCount int `gorm:"column:retry_count;default:0"`
	// 错误信息
	ErrorMessage sql.NullString `gorm:"column:error_message;type:text"`
	// 执行结果详情
	Result sql.NullString `gorm:"column:result;type:text"`
	// 创建时间
	Ctime float64 `gorm:"column:ctime"`
}
