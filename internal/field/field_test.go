package field

import (
	"math"
	"testing"

	"task209-deformgate/internal/model"
)

func TestHashFieldDeterministicAndContentSensitive(t *testing.T) {
	dims := model.Dims{2, 2}
	base := []model.Vec3{{X: 1}, {Y: 2}, {Z: 3}, {X: 4}}
	if got, want := HashField(dims, base), HashField(dims, base); got != want {
		t.Fatalf("same field hash changed: %s != %s", got, want)
	}
	changed := append([]model.Vec3(nil), base...)
	changed[0].X++
	if HashField(dims, base) == HashField(dims, changed) {
		t.Fatal("different displacement content must have different hashes")
	}
}

func TestValidateDisplacementsRejectsWrongVolumeAndNonFiniteValues(t *testing.T) {
	dims := model.Dims{2, 2}
	valid := make([]model.Vec3, dims.Volume())
	if err := ValidateDisplacements(dims, valid); err != nil {
		t.Fatalf("valid displacement field rejected: %v", err)
	}
	if err := ValidateDisplacements(dims, valid[:len(valid)-1]); err == nil {
		t.Fatal("short displacement field must be rejected")
	}
	invalid := append([]model.Vec3(nil), valid...)
	invalid[1].Y = math.NaN()
	if err := ValidateDisplacements(dims, invalid); err == nil {
		t.Fatal("NaN displacement must be rejected")
	}
	invalid[1].Y = math.Inf(1)
	if err := ValidateDisplacements(dims, invalid); err == nil {
		t.Fatal("infinite displacement must be rejected")
	}
}

func TestInverseConsistencyErrorUsesVectorNorm(t *testing.T) {
	got := InverseConsistencyError(model.Vec3{X: 3}, model.Vec3{Y: 4})
	if math.Abs(got-5) > 1e-9 {
		t.Fatalf("inverse consistency error=%v, want 5", got)
	}
}
