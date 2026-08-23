package quality

import (
	"testing"

	"task209-deformgate/internal/field"
	"task209-deformgate/internal/model"
)

func TestComputeJacobianIdentity(t *testing.T) {
	dims := model.Dims{4, 4, 4}
	disp := field.SynthesizeIdentity(dims)
	res := ComputeJacobian(disp, dims, 0.5, 2.0)
	if res.Stats.FoldCount != 0 {
		t.Fatalf("恒等场不应有折叠，得到 %d", res.Stats.FoldCount)
	}
	// 恒等场 det(J) 应恒为 1。
	if res.Stats.Min < 0.999 || res.Stats.Max > 1.001 {
		t.Fatalf("恒等场 det(J) 应≈1，得到 min=%v max=%v", res.Stats.Min, res.Stats.Max)
	}
}

func TestComputeJacobianLocalFold(t *testing.T) {
	dims := model.Dims{8, 8, 4}
	center := 4*8*4 + 4*8 + 4
	disp := field.SynthesizeLocalFold(dims, center)
	res := ComputeJacobian(disp, dims, 0.5, 2.0)
	if res.Stats.FoldCount == 0 {
		t.Fatal("局部反向位移应产生折叠体素")
	}
	// 折叠体素应集中在中心附近。
	for _, f := range res.Folds {
		dx := f.Pos.X - 4
		dy := f.Pos.Y - 4
		if dx*dx+dy*dy > 4 {
			t.Fatalf("折叠体素 %v 距中心过远", f.Pos)
		}
	}
}

func TestComputeJacobianAffine2D(t *testing.T) {
	dims := model.Dims{6, 6}
	disp := field.SynthesizeAffine(dims, 0.02, model.Vec3{})
	res := ComputeJacobian(disp, dims, 0.5, 2.0)
	if res.Stats.FoldCount != 0 {
		t.Fatalf("仿射场不应折叠，得到 %d", res.Stats.FoldCount)
	}
	// det(J) ≈ (1+0.02)^2 ≈ 1.0404。
	if res.Stats.Min < 1.03 || res.Stats.Max > 1.05 {
		t.Fatalf("仿射场 det(J) 应≈1.0404，得到 min=%v max=%v", res.Stats.Min, res.Stats.Max)
	}
}
