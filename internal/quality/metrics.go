package quality

import (
	"task209-deformgate/internal/model"
)

// Evaluate 是质量门的一站式评估入口：给定形变场位移、影像维度、
// 逆一致性采样点与检查参数，计算三组指标并给出判定。
func Evaluate(disp []model.Vec3, dims model.Dims, samples []model.SamplePoint, p model.CheckParams) model.QualityResult {
	r := model.QualityResult{
		ParamsID: p.ID,
	}

	jr := ComputeJacobian(disp, dims, p.JacMin, p.JacMax)
	r.Jacobian = jr.Stats
	r.Folds = jr.Folds
	r.Inverse = ComputeInverseStats(samples, p.ICEThreshold)
	r.Boundary = ComputeBoundaryStats(disp, dims)
	r.Verdict = Decide(r.Jacobian, r.Inverse, r.Boundary, p)

	return r
}
