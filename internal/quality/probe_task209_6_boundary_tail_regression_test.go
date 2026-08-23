package quality

import (
	"math"
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug06BoundaryTailRegression(t *testing.T) {
	disp := make([]model.Vec3, 4)
	disp[3].Y = math.NaN()
	stats := ComputeBoundaryStats(disp, model.Dims{2, 2})
	if stats.Missing != 1 {
		t.Fatalf("last boundary voxel was not counted: %+v", stats)
	}
}
