package model

// ImagePair 状态常量：影像对的完整生命周期。
const (
	// PairRegistered 登记：已录入元数据，尚未关联可校验形变场。
	PairRegistered = "registered"
	// PairPending 待校验：已关联形变场，等待（或正在）执行质量检查。
	PairPending = "pending_validation"
	// PairPassed 已通过：质量门判定通过，下游可安全使用配准结果。
	PairPassed = "passed"
	// PairReview 需复核：指标部分超限，需人工复核形变场。
	PairReview = "needs_review"
	// PairRejected 拒绝：存在折叠或严重缺边界，配准结果不可信。
	PairRejected = "rejected"
	// PairArchived 封存：终态，不可再修改或追加场。
	PairArchived = "archived"
)

// ImagePair 是待校验配准结果的影像对（参考影像 fixed + 待配准影像 moving）。
// 它只保存几何元数据，不保存原始影像像素；形变场另见 DeformField。
type ImagePair struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Dims      Dims    `json:"dims"`    // 体素维度 [nx, ny] 或 [nx, ny, nz]
	Spacing   []float64 `json:"spacing"` // 体素间距（mm），长度与 Dims 一致
	Axes      string  `json:"axes"`    // 坐标方向约定，如 "RAS" / "LPS" / "ASL"
	Modality  string  `json:"modality"` // 模态，如 "MRI" / "CT" / "US"
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// StatusConstants 返回所有合法影像对状态（用于校验与文档）。
func ImagePairStates() []string {
	return []string{PairRegistered, PairPending, PairPassed, PairReview, PairRejected, PairArchived}
}
