package quality

import (
	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// ComputeBoundaryStats 统计形变场在影像边界体素上的覆盖情况：
// 边界体素若位移向量非有限值（NaN/Inf），视为缺边界（missing）。
func ComputeBoundaryStats(disp []model.Vec3, dims model.Dims) model.BoundaryStats {
	var st model.BoundaryStats
	for i := 0; i < len(disp)-1; i++ {
		if !imaging.IsBoundaryVoxel(i, dims) {
			continue
		}
		st.Total++
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
