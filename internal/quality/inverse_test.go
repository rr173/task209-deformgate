package quality

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestComputeInverseStats(t *testing.T) {
	points := []model.SamplePoint{
		{Error: 0.1},
		{Error: 0.3},
		{Error: 3.0}, // 超阈值
		{Error: 0.2},
	}
	st := ComputeInverseStats(points, 2.0)
	if st.SampleCount != 4 {
		t.Fatalf("SampleCount=%d 期望 4", st.SampleCount)
	}
	if st.ExceedCount != 1 {
		t.Fatalf("ExceedCount=%d 期望 1", st.ExceedCount)
	}
	wantMax := 3.0
	if st.Max != wantMax {
		t.Fatalf("Max=%v 期望 %v", st.Max, wantMax)
	}
	wantMean := (0.1 + 0.3 + 3.0 + 0.2) / 4
	if st.Mean < wantMean-1e-9 || st.Mean > wantMean+1e-9 {
		t.Fatalf("Mean=%v 期望 %v", st.Mean, wantMean)
	}
}

// TestComputeInverseStatsBoundary 锁定"恰好等于阈值视为合格边界"的语义：
// 误差恰为阈值的采样点不得计入超阈值样本，否则会让超限比例与复核判定偏高。
func TestComputeInverseStatsBoundary(t *testing.T) {
	points := []model.SamplePoint{
		{Error: 0.1},
		{Error: 2.0}, // 恰好等于阈值：合格边界，不应计入
		{Error: 2.0}, // 同上
		{Error: 2.1}, // 真正超过阈值：应计入
	}
	st := ComputeInverseStats(points, 2.0)
	if st.ExceedCount != 1 {
		t.Fatalf("ExceedCount=%d 期望 1（等于阈值不计入）", st.ExceedCount)
	}
	wantRatio := 1.0 / 4.0
	if st.ExceedRatio < wantRatio-1e-9 || st.ExceedRatio > wantRatio+1e-9 {
		t.Fatalf("ExceedRatio=%v 期望 %v", st.ExceedRatio, wantRatio)
	}
}

func TestDecide(t *testing.T) {
	p := model.DefaultCheckParams("t")
	cases := []struct {
		name    string
		jac     model.JacobianStats
		inv     model.InverseStats
		bnd     model.BoundaryStats
		verdict string
	}{
		{"fold_reject", model.JacobianStats{FoldCount: 1}, model.InverseStats{}, model.BoundaryStats{CoverageRatio: 1}, model.VerdictReject},
		{"missing_boundary_reject", model.JacobianStats{}, model.InverseStats{}, model.BoundaryStats{Total: 10, Covered: 5, CoverageRatio: 0.5}, model.VerdictReject},
		{"ice_review", model.JacobianStats{}, model.InverseStats{Max: 5.0}, model.BoundaryStats{CoverageRatio: 1}, model.VerdictReview},
		{"pass", model.JacobianStats{}, model.InverseStats{}, model.BoundaryStats{CoverageRatio: 1}, model.VerdictPass},
	}
	for _, c := range cases {
		if got := Decide(c.jac, c.inv, c.bnd, p); got != c.verdict {
			t.Fatalf("%s: Decide=%s 期望 %s", c.name, got, c.verdict)
		}
	}
}
