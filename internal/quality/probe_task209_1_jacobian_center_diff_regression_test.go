package quality

import (
	"testing"

	"task209-deformgate/internal/field"
	"task209-deformgate/internal/model"
)

func TestTask209Bug01JacobianCenterDifferenceRegression(t *testing.T) {
	dims := model.Dims{6, 6}
	result := ComputeJacobian(field.SynthesizeAffine(dims, 0.02, model.Vec3{}), dims, 0.5, 2.0)
	if result.Stats.Min < 1.03 || result.Stats.Max > 1.05 {
		t.Fatalf("affine det range=[%v,%v], want approximately [1.03,1.05]", result.Stats.Min, result.Stats.Max)
	}
}
