package model

// DeformField 状态常量：形变场的生命周期。
const (
	// FieldUploading 上传中：位移向量尚未完整写入。
	FieldUploading = "uploading"
	// FieldValid 有效：位移向量完整、有限，且边界覆盖充分。
	FieldValid = "valid"
	// FieldFolded 折叠：存在雅可比行列式 ≤ 0 的折叠体素。
	FieldFolded = "folded"
	// FieldMissingBoundary 缺边界：边界体素存在无效/缺失位移。
	FieldMissingBoundary = "missing_boundary"
)

// DeformField 是配准产生的形变场：每个体素一个位移向量，把参考影像体素
// 映射到待配准影像空间。Displacements 按行优先（x 最内层）展平存储。
type DeformField struct {
	ID           int64   `json:"id"`
	ImagePairID  int64   `json:"image_pair_id"`
	Hash         string  `json:"hash"`         // 内容哈希（幂等键）
	Dims         Dims    `json:"dims"`         // 必须与关联影像对 Dims 一致
	Displacements []Vec3  `json:"displacements"` // 长度 = Dims.Volume()
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// FieldStates 返回所有合法形变场状态。
func FieldStates() []string {
	return []string{FieldUploading, FieldValid, FieldFolded, FieldMissingBoundary}
}

// SamplePoint 状态常量。
const (
	// SampleOpen 待校验：采样点已录入，尚未参与逆一致性评估。
	SampleOpen = "open"
	// SampleChecked 已校验：已计算逆一致性误差并纳入统计。
	SampleChecked = "checked"
)

// SamplePoint 是用于逆一致性检查的采样点：记录参考空间坐标 Pos、
// 正变换位移 Forward 与在映射点处的逆变换位移 Inverse。
// 逆一致性误差 = |Forward + Inverse|（理想情况互为相反向量）。
type SamplePoint struct {
	ID       int64  `json:"id"`
	FieldID  int64  `json:"field_id"`
	Pos      Vec3   `json:"pos"`
	Forward  Vec3   `json:"forward"`
	Inverse  Vec3   `json:"inverse"`
	Error    float64 `json:"error"` // 逆一致性误差（mm）
	Status   string `json:"status"`
}
