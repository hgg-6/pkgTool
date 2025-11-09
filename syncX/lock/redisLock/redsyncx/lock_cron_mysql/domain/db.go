package domain

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
	ID     int64 `json:"id"`
	CronId int64 `json:"cronId"`
	// 任务名
	Name string `json:"name"`
	// 任务描述
	Description string `json:"description"`
	// 任务执行表达式
	CronExpr string `json:"cronExpr"`
	// 任务类型
	TaskType TaskType `json:"taskType"`

	// 任务状态
	Status JobStatus `json:"status"`
	// 任务最大重试次数
	MaxRetry int `json:"maxRetry"`
	// 任务超时时间(秒)
	Timeout int `json:"timeout"`

	Ctime float64 `json:"ctime"`
	Utime float64 `json:"utime"`
}
