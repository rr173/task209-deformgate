package quality

import "task209-deformgate/internal/model"

// Decide 依据雅可比、逆一致性与边界覆盖三组指标给出质量判定。
// 判定优先级：折叠/严重缺边界 → reject；指标轻度超限 → review；其余 → pass。
func Decide(jac model.JacobianStats, inv model.InverseStats, bnd model.BoundaryStats, p model.CheckParams) string {
	// 折叠：配准局部不可逆，直接拒绝。
	if jac.FoldCount > 0 {
		return model.VerdictReject
	}
	// 严重缺边界：边界体素覆盖不足，直接拒绝。
	if bnd.Total > 0 && bnd.CoverageRatio < p.CoverageThreshold {
		return model.VerdictReject
	}
	// 逆一致性超标：正逆变换不自洽，需复核。
	if inv.Max > p.ICEThreshold || inv.ExceedRatio > 0.05 {
		return model.VerdictReview
	}
	// 异常体积收缩/膨胀：需复核。
	if jac.ShrinkCount > 0 || jac.ExpandCount > 0 {
		return model.VerdictReview
	}
	return model.VerdictPass
}
