package quality

import (
	"math"
	"testing"

	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// TestComputeBoundaryStatsIncludesLastVoxel 回归：覆盖统计必须遍历到最后一个体素。
// 最后一个体素位于 (nx-1, ny-1, nz-1)，落在每一维尾部，必为边界体素。
// 旧实现用 len(disp)-1 截断循环，遗漏该体素，使其无效位移不计入 Missing。
func TestComputeBoundaryStatsIncludesLastVoxel(t *testing.T) {
	dims := model.Dims{3, 3, 3}
	last := dims.Volume() - 1 // 索引 26，坐标 (2,2,2)

	if !imaging.IsBoundaryVoxel(last, dims) {
		t.Fatalf("最后一个体素 %d 必须是边界体素", last)
	}

	// 全部边界体素先置为有效，再令最后一个体素位移为 NaN。
	disp := make([]model.Vec3, dims.Volume())
	for i := range disp {
		disp[i] = model.Vec3{X: 0, Y: 0, Z: 0}
	}
	disp[last] = model.Vec3{X: math.NaN(), Y: math.NaN(), Z: math.NaN()}

	st := ComputeBoundaryStats(disp, dims)

	if st.Missing != 1 {
		t.Fatalf("Missing=%d 期望 1（最后一个边界体素无效位移应计入）", st.Missing)
	}
	if st.Total != st.Covered+st.Missing {
		t.Fatalf("Total=%d 应等于 Covered+Missing=%d", st.Total, st.Covered+st.Missing)
	}
	if got := float64(st.Covered) / float64(st.Total); st.CoverageRatio < got-1e-9 || st.CoverageRatio > got+1e-9 {
		t.Fatalf("CoverageRatio=%v 期望 %v", st.CoverageRatio, got)
	}

	// 二维场同理：最后一个体素 (nx-1, ny-1) 也是边界体素。
	dims2 := model.Dims{4, 4}
	last2 := dims2.Volume() - 1
	disp2 := make([]model.Vec3, dims2.Volume())
	disp2[last2] = model.Vec3{X: math.NaN()}
	st2 := ComputeBoundaryStats(disp2, dims2)
	if st2.Missing != 1 {
		t.Fatalf("2D Missing=%d 期望 1（最后一个边界体素无效位移应计入）", st2.Missing)
	}
}
