package model

// CheckParams 是质量门检查参数的版本化快照。
// 每个检查任务绑定一个参数版本；参数锁定后不可修改，只能新建版本再激活。
type CheckParams struct {
	ID                int64   `json:"id"`
	Name              string  `json:"name"`
	JacMin            float64 `json:"jac_min"` // 雅可比收缩阈值（如 0.5）
	JacMax            float64 `json:"jac_max"` // 雅可比膨胀阈值（如 2.0）
	ICEThreshold      float64 `json:"ice_threshold"` // 逆一致性误差阈值（mm）
	CoverageThreshold float64 `json:"coverage_threshold"` // 边界覆盖率阈值（0~1）
	Active            bool    `json:"active"`   // 是否为当前激活版本
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

// DefaultCheckParams 返回默认检查参数（与 REQ 描述的阈值对齐）。
func DefaultCheckParams(name string) CheckParams {
	return CheckParams{
		Name:              name,
		JacMin:            0.5,
		JacMax:            2.0,
		ICEThreshold:      2.0,
		CoverageThreshold: 0.95,
	}
}
