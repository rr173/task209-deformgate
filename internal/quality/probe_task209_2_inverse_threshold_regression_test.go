package quality

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug02InverseThresholdRegression(t *testing.T) {
	stats := ComputeInverseStats([]model.SamplePoint{{Error: 2}}, 2)
	if stats.ExceedCount != 0 || stats.ExceedRatio != 0 {
		t.Fatalf("error at threshold counted as exceed: %+v", stats)
	}
}
