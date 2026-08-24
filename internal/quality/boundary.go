package quality

import (
	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// ComputeBoundaryStats 统计形变场在影像边界体素上的覆盖情况：
// 边界体素若位移向量非有限值（NaN/Inf），视为缺边界（missing）。
func ComputeBoundaryStats(disp []model.Vec3, dims model.Dims) model.BoundaryStats {
	var st model.BoundaryStats
	// 遍历整个位移序列（含首尾），保证所有边界体素都参与覆盖率计算。
	// 最后一个体素位于 (nx-1, ny-1, nz-1)，落在每一维的尾部，必为边界体素，
	// 故不可用 len(disp)-1 截断，否则其无效位移不会计入 Missing。
	for i := 0; i < len(disp); i++ {
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
