package quality

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug05CoverageThresholdRegression(t *testing.T) {
	p := model.DefaultCheckParams("coverage")
	p.CoverageThreshold = 0.95
	got := Decide(model.JacobianStats{}, model.InverseStats{}, model.BoundaryStats{Total: 20, Covered: 19, Missing: 1, CoverageRatio: 0.95}, p)
	if got != model.VerdictPass {
		t.Fatalf("coverage exactly at threshold got %q, want pass", got)
	}
}
