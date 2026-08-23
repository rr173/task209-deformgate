package model

// QualityResult 状态常量：质量结果的版本生命周期。
const (
	// ResultDraft 草稿：指标已计算但未发布。
	ResultDraft = "draft"
	// ResultPublished 发布：作为该影像对的权威质量结论。
	ResultPublished = "published"
	// ResultSuperseded 替代：已被更新的结果显式替代，保留历史。
	ResultSuperseded = "superseded"
)

// 质量门判定常量。
const (
	// VerdictPass 通过：所有指标在阈值内。
	VerdictPass = "pass"
	// VerdictReview 需复核：部分指标超限但无折叠/严重缺边界。
	VerdictReview = "review"
	// VerdictReject 拒绝：存在折叠或严重缺边界。
	VerdictReject = "reject"
)

// JacobianStats 汇总形变场雅可比行列式 det(J) 的统计与折叠证据。
type JacobianStats struct {
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Mean       float64 `json:"mean"`
	FoldCount  int     `json:"fold_count"`  // det(J) ≤ 0 的体素数
	ShrinkCount int    `json:"shrink_count"` // det(J) < 收缩阈值的体素数
	ExpandCount int    `json:"expand_count"` // det(J) > 膨胀阈值的体素数
	Total      int     `json:"total"`
}

// FoldRegion 描述一个折叠/异常体素的位置与行列式值。
type FoldRegion struct {
	Index int     `json:"index"` // 展平体素序号
	Pos   Vec3    `json:"pos"`   // 体素坐标（ix, iy, iz）
	Det   float64 `json:"det"`
}

// InverseStats 汇总逆一致性误差统计。
type InverseStats struct {
	Mean        float64 `json:"mean"`
	Max         float64 `json:"max"`
	SampleCount int     `json:"sample_count"`
	ExceedCount int     `json:"exceed_count"` // 误差超阈值的采样点数
	ExceedRatio float64 `json:"exceed_ratio"`
}

// BoundaryStats 汇总边界体素的覆盖情况。
type BoundaryStats struct {
	Total          int     `json:"total"`           // 边界体素总数
	Covered        int     `json:"covered"`         // 有效覆盖的边界体素数
	Missing        int     `json:"missing"`         // 缺失/无效的边界体素数
	CoverageRatio  float64 `json:"coverage_ratio"`  // Covered / Total
}

// QualityResult 是一次检查任务产出的、可版本化的质量结论。
// 它绑定检查参数版本与全部量化指标，发布后不可变，只能被新结果替代。
type QualityResult struct {
	ID          int64         `json:"id"`
	TaskID      int64         `json:"task_id"`
	ImagePairID int64         `json:"image_pair_id"`
	FieldID     int64         `json:"field_id"`
	ParamsID    int64         `json:"params_id"`
	Verdict     string        `json:"verdict"`
	Jacobian    JacobianStats `json:"jacobian"`
	Inverse     InverseStats  `json:"inverse"`
	Boundary    BoundaryStats `json:"boundary"`
	Folds       []FoldRegion  `json:"folds"`
	Status      string        `json:"status"`
	Supersedes  int64         `json:"supersedes"` // 被替代的结果 ID（0 表示无）
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

// ResultStates 返回所有合法质量结果状态。
func ResultStates() []string {
	return []string{ResultDraft, ResultPublished, ResultSuperseded}
}

// Verdicts 返回所有合法判定值。
func Verdicts() []string {
	return []string{VerdictPass, VerdictReview, VerdictReject}
}
