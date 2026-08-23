package imaging

import (
	"fmt"

	"task209-deformgate/internal/model"
)

// ValidateFieldDims 校验形变场维度与影像对维度一致，且位移向量数量与体素总数一致。
// 维度不匹配是 REQ 明确要求的拒绝条件。
func ValidateFieldDims(pairDims, fieldDims model.Dims, nDisplacements int) error {
	if !pairDims.Equal(fieldDims) {
		return fmt.Errorf("%w: 形变场维度 %v 与影像对维度 %v 不一致", model.ErrInvalid, fieldDims, pairDims)
	}
	vol := fieldDims.Volume()
	if nDisplacements != vol {
		return fmt.Errorf("%w: 位移向量数 %d 与体素总数 %d 不一致", model.ErrInvalid, nDisplacements, vol)
	}
	return nil
}

// ValidateSamplePosition 校验采样点是否落在影像体素范围内（越界拒绝）。
func ValidateSamplePosition(pos model.Vec3, dims model.Dims) error {
	nx := dims[0]
	ny := dims[1]
	nz := 1
	if len(dims) >= 3 {
		nz = dims[2]
	}
	if pos.X < 0 || pos.X >= float64(nx) ||
		pos.Y < 0 || pos.Y >= float64(ny) ||
		pos.Z >= float64(nz) {
		return fmt.Errorf("%w: 采样点 (%.2f, %.2f, %.2f) 越界，影像维度 %v", model.ErrInvalid, pos.X, pos.Y, pos.Z, dims)
	}
	return nil
}

// IsBoundaryVoxel 判断展平体素序号是否位于影像边界（任一维位于首尾）。
// 质量门用它统计边界覆盖率。
func IsBoundaryVoxel(index int, dims model.Dims) bool {
	nx := dims[0]
	ny := dims[1]
	nz := 1
	if len(dims) >= 3 {
		nz = dims[2]
	}
	iz := index / (nx * ny)
	rem := index % (nx * ny)
	iy := rem / nx
	ix := rem % nx
	return ix == 0 || ix == nx-1 || iy == 0 || iy == ny-1 || iz == 0 || iz == nz-1
}

// VoxelPos 把展平体素序号还原为 (ix, iy, iz) 坐标。
func VoxelPos(index int, dims model.Dims) model.Vec3 {
	nx := dims[0]
	ny := dims[1]
	iz := index / (nx * ny)
	rem := index % (nx * ny)
	iy := rem / nx
	ix := rem % nx
	return model.Vec3{X: float64(ix), Y: float64(iy), Z: float64(iz)}
}
