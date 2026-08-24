package quality

import (
	"math"
	"testing"

	"task209-deformgate/internal/model"
)

// TestComputeBoundaryStatsInfCountsAsMissing 回归：边界体素位移为 ±Inf 时，
// 必须与 NaN 一样计入缺失（missing），不得视为有效覆盖。
// 否则 CoverageRatio 被高估，质量判定会过于乐观地放过缺边界场。
func TestComputeBoundaryStatsInfCountsAsMissing(t *testing.T) {
	// 3x3x3 体素：所有面体素都是边界体素，唯一内部体素是 (1,1,1)。
	dims := model.Dims{3, 3, 3}
	disp := make([]model.Vec3, dims.Volume())
	for i := range disp {
		disp[i] = model.Vec3{X: 0, Y: 0, Z: 0}
	}

	// 选取一个边界体素序号置为 +Inf，另一个置为 -Inf，第三个置为 NaN。
	// (0,0,0)=0、(2,0,0)=2、(0,2,0)=6 均为边界体素。
	disp[0] = model.Vec3{X: math.Inf(1)}
	disp[2] = model.Vec3{Y: math.Inf(-1)}
	disp[6] = model.Vec3{Z: math.NaN()}

	st := ComputeBoundaryStats(disp, dims)

	// 3x3x3 边界体素总数 = 27 - 1（内部）= 26。
	if st.Total != 26 {
		t.Fatalf("Total=%d 期望 26", st.Total)
	}
	if st.Missing != 3 {
		t.Fatalf("Missing=%d 期望 3（±Inf 与 NaN 各计一处）", st.Missing)
	}
	if st.Covered != 23 {
		t.Fatalf("Covered=%d 期望 23", st.Covered)
	}
	wantRatio := float64(23) / float64(26)
	if st.CoverageRatio < wantRatio-1e-9 || st.CoverageRatio > wantRatio+1e-9 {
		t.Fatalf("CoverageRatio=%v 期望 %v", st.CoverageRatio, wantRatio)
	}
}

// TestComputeBoundaryStatsFiniteCovered 有限位移的边界体素全部计入有效覆盖。
func TestComputeBoundaryStatsFiniteCovered(t *testing.T) {
	dims := model.Dims{3, 3, 3}
	disp := make([]model.Vec3, dims.Volume())
	for i := range disp {
		disp[i] = model.Vec3{X: 1.0, Y: -2.0, Z: 0.5}
	}
	st := ComputeBoundaryStats(disp, dims)
	if st.Total != 26 {
		t.Fatalf("Total=%d 期望 26", st.Total)
	}
	if st.Missing != 0 {
		t.Fatalf("Missing=%d 期望 0", st.Missing)
	}
	if st.Covered != 26 {
		t.Fatalf("Covered=%d 期望 26", st.Covered)
	}
	if st.CoverageRatio != 1.0 {
		t.Fatalf("CoverageRatio=%v 期望 1.0", st.CoverageRatio)
	}
}
