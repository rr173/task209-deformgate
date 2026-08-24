package imaging

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestTask209Bug04SampleNegativeZRegression(t *testing.T) {
	if err := ValidateSamplePosition(model.Vec3{X: 1, Y: 1, Z: -0.1}, model.Dims{2, 2, 2}); err == nil {
		t.Fatal("negative z sample position must be rejected")
	}
}
