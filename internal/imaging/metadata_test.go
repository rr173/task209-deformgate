package imaging

import (
	"testing"

	"task209-deformgate/internal/model"
)

func TestValidatePair(t *testing.T) {
	ok := &model.ImagePair{Name: "a", Dims: model.Dims{8, 8, 4}, Spacing: []float64{1, 1, 1}, Axes: "RAS"}
	if err := ValidatePair(ok); err != nil {
		t.Fatalf("合法影像对应通过，得到 %v", err)
	}

	if err := ValidatePair(&model.ImagePair{Name: "b", Dims: model.Dims{8}, Spacing: []float64{1}, Axes: "RAS"}); err == nil {
		t.Fatal("一维维度应拒绝")
	}
	if err := ValidatePair(&model.ImagePair{Name: "c", Dims: model.Dims{8, 8}, Spacing: []float64{1, 1}, Axes: ""}); err == nil {
		t.Fatal("未声明坐标方向应拒绝")
	}
	if err := ValidatePair(&model.ImagePair{Name: "d", Dims: model.Dims{8, 8}, Spacing: []float64{1}, Axes: "RAS"}); err == nil {
		t.Fatal("spacing 长度不一致应拒绝")
	}
}

func TestValidateSamplePosition(t *testing.T) {
	dims := model.Dims{8, 8, 4}
	if err := ValidateSamplePosition(model.Vec3{X: 3, Y: 3, Z: 2}, dims); err != nil {
		t.Fatalf("范围内采样点应通过，得到 %v", err)
	}
	if err := ValidateSamplePosition(model.Vec3{X: 9, Y: 3, Z: 2}, dims); err == nil {
		t.Fatal("越界采样点应拒绝")
	}
}

func TestBoundaryVoxel(t *testing.T) {
	dims := model.Dims{8, 8, 4}
	if !IsBoundaryVoxel(0, dims) {
		t.Fatal("(0,0,0) 应为边界体素")
	}
	if IsBoundaryVoxel(1*8*8 + 4*8 + 4, dims) {
		t.Fatal("(4,4,1) 不应为边界体素")
	}
}
