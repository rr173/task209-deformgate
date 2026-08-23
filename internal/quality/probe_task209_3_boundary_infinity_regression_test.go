package quality

import (
	"math"
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug03BoundaryInfinityRegression(t *testing.T) {
	disp := make([]model.Vec3, 4)
	disp[0].X = math.Inf(1)
	stats := ComputeBoundaryStats(disp, model.Dims{2, 2})
	if stats.Missing != 1 || stats.Covered != 3 {
		t.Fatalf("infinite boundary vector not rejected: %+v", stats)
	}
}
