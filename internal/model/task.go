package model

// CheckTask 状态常量：检查任务的执行生命周期。
const (
	// TaskQueued 排队：已创建，等待执行。
	TaskQueued = "queued"
	// TaskRunning 运行中：正在计算质量指标。
	TaskRunning = "running"
	// TaskCompleted 完成：指标计算完成并已产出质量结果。
	TaskCompleted = "completed"
	// TaskFailed 失败：计算过程发生不可恢复错误。
	TaskFailed = "failed"
)

// CheckTask 表示一次对 (影像对, 形变场) 的质量门检查任务。
// 同一 (影像对, 形变场) 只允许一个活动（排队/运行中）任务，用部分唯一索引保证。
type CheckTask struct {
	ID          int64  `json:"id"`
	ImagePairID int64  `json:"image_pair_id"`
	FieldID     int64  `json:"field_id"`
	ParamsID    int64  `json:"params_id"` // 检查参数版本
	Status      string `json:"status"`
	Message     string `json:"message"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// TaskStates 返回所有合法检查任务状态。
func TaskStates() []string {
	return []string{TaskQueued, TaskRunning, TaskCompleted, TaskFailed}
}
