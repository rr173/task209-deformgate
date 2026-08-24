package quality

import (
	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// ComputeBoundaryStats 统计形变场在影像边界体素上的覆盖情况：
// 边界体素若位移向量非有限值（NaN/Inf），视为缺边界（missing）。
func ComputeBoundaryStats(disp []model.Vec3, dims model.Dims) model.BoundaryStats {
	var st model.BoundaryStats
	for i := 0; i < len(disp); i++ {
		if !imaging.IsBoundaryVoxel(i, dims) {
			continue
		}
		st.Total++
		// 无穷位移（±Inf）与 NaN 同属无效数据：边界体素任一分量非有限值，
		// 即视为缺边界（missing），避免把无穷覆盖计入有效统计导致判定过于乐观。
		if disp[i].IsFinite() {
			st.Covered++
		} else {
			st.Missing++
		}
	}
	if st.Total > 0 {
		st.CoverageRatio = float64(st.Covered) / float64(st.Total)
	}
	return st
}
