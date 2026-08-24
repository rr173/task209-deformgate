// Package quality 实现形变场质量门的三类核心指标：雅可比行列式、
// 逆一致性与边界覆盖，并据此给出质量判定。
package quality

import (
	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// JacobianResult 是雅可比行列式评估的完整产物：统计摘要 + 折叠证据。
type JacobianResult struct {
	Stats model.JacobianStats
	Folds []model.FoldRegion
}

// ComputeJacobian 计算形变场每个体素的雅可比行列式 det(J)，
// 其中 J = I + ∇u。det(J) ≤ 0 表示局部折叠（配准不可逆）。
// 偏导数用体素单位中心差分近似（边界退化为单边差分）。
// jacMin/jacMax 用于统计异常收缩与膨胀体素数。
func ComputeJacobian(disp []model.Vec3, dims model.Dims, jacMin, jacMax float64) JacobianResult {
	total := len(disp)
	res := JacobianResult{
		Stats: model.JacobianStats{Total: total},
		Folds: []model.FoldRegion{},
	}
	if total == 0 {
		return res
	}

	rank := dims.Rank()
	var min, max, sum float64
	first := true
	for i := 0; i < total; i++ {
		det := jacobianAt(disp, dims, rank, i)
		if first {
			min, max = det, det
			first = false
		} else {
			if det < min {
				min = det
			}
			if det > max {
				max = det
			}
		}
		sum += det

		if det <= 0 {
			res.Stats.FoldCount++
			res.Folds = append(res.Folds, model.FoldRegion{
				Index: i,
				Pos:   imaging.VoxelPos(i, dims),
				Det:   det,
			})
		} else if det < jacMin {
			res.Stats.ShrinkCount++
		} else if det > jacMax {
			res.Stats.ExpandCount++
		}
	}

	res.Stats.Min = min
	res.Stats.Max = max
	res.Stats.Mean = sum / float64(total)
	return res
}

// jacobianAt 计算单个体素的雅可比行列式。
func jacobianAt(disp []model.Vec3, dims model.Dims, rank, index int) float64 {
	nx := dims[0]
	ny := dims[1]
	nz := 1
	if rank == 3 {
		nz = dims[2]
	}
	iz := index / (nx * ny)
	rem := index % (nx * ny)
	iy := rem / nx
	ix := rem % nx

	// 位移梯度（中心差分，边界 clamp）。
	duxdx := gradX(disp, dims, ix, iy, iz, nx, ny, nz, 0)
	duxdy := gradY(disp, dims, ix, iy, iz, nx, ny, nz, 0)
	duydx := gradX(disp, dims, ix, iy, iz, nx, ny, nz, 1)
	duydy := gradY(disp, dims, ix, iy, iz, nx, ny, nz, 1)

	if rank == 2 {
		// 2x2 雅可比 J = I + ∇u。
		j00 := 1 + duxdx
		j01 := duxdy
		j10 := duydx
		j11 := 1 + duydy
		return j00*j11 - j01*j10
	}

	// 3x3 雅可比。
	duxdz := gradZ(disp, dims, ix, iy, iz, nx, ny, nz, 0)
	duydz := gradZ(disp, dims, ix, iy, iz, nx, ny, nz, 1)
	duzdx := gradX(disp, dims, ix, iy, iz, nx, ny, nz, 2)
	duzdy := gradY(disp, dims, ix, iy, iz, nx, ny, nz, 2)
	duzdz := gradZ(disp, dims, ix, iy, iz, nx, ny, nz, 2)

	a := 1 + duxdx
	b := duxdy
	c := duxdz
	d := duydx
	e := 1 + duydy
	f := duydz
	g := duzdx
	h := duzdy
	i := 1 + duzdz

	return a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
}

// component 常量：位移向量的分量下标（X/Y/Z）。
const (
	cX = 0
	cY = 1
	cZ = 2
)

// gradX 计算位移向量第 c 分量对 x 的差分：内部中心差分（步长 2），
// 边界自适应单边差分（步长 1），保证边界体素梯度不被低估。
func gradX(disp []model.Vec3, dims model.Dims, ix, iy, iz, nx, ny, nz, c int) float64 {
	cur := component(dispAt(disp, dims, ix, iy, iz, nx, ny), c)
	if ix == 0 {
		return component(dispAt(disp, dims, ix+1, iy, iz, nx, ny), c) - cur
	}
	if ix == nx-1 {
		return cur - component(dispAt(disp, dims, ix-1, iy, iz, nx, ny), c)
	}
	return (component(dispAt(disp, dims, ix+1, iy, iz, nx, ny), c) -
		component(dispAt(disp, dims, ix-1, iy, iz, nx, ny), c)) / 2
}

// gradY 计算位移向量第 c 分量对 y 的差分（边界自适应单边）。
func gradY(disp []model.Vec3, dims model.Dims, ix, iy, iz, nx, ny, nz, c int) float64 {
	cur := component(dispAt(disp, dims, ix, iy, iz, nx, ny), c)
	if iy == 0 {
		return component(dispAt(disp, dims, ix, iy+1, iz, nx, ny), c) - cur
	}
	if iy == ny-1 {
		return cur - component(dispAt(disp, dims, ix, iy-1, iz, nx, ny), c)
	}
	return (component(dispAt(disp, dims, ix, iy+1, iz, nx, ny), c) -
		component(dispAt(disp, dims, ix, iy-1, iz, nx, ny), c)) / 2
}

// gradZ 计算位移向量第 c 分量对 z 的差分（边界自适应单边）。
func gradZ(disp []model.Vec3, dims model.Dims, ix, iy, iz, nx, ny, nz, c int) float64 {
	cur := component(dispAt(disp, dims, ix, iy, iz, nx, ny), c)
	if iz == 0 {
		return component(dispAt(disp, dims, ix, iy, iz+1, nx, ny), c) - cur
	}
	if iz == nz-1 {
		return cur - component(dispAt(disp, dims, ix, iy, iz-1, nx, ny), c)
	}
	return (component(dispAt(disp, dims, ix, iy, iz+1, nx, ny), c) -
		component(dispAt(disp, dims, ix, iy, iz-1, nx, ny), c)) / 2
}

// dispAt 按 (ix, iy, iz) 读取位移向量。
func dispAt(disp []model.Vec3, dims model.Dims, ix, iy, iz, nx, ny int) model.Vec3 {
	idx := (iz*ny+iy)*nx + ix
	if idx < 0 || idx >= len(disp) {
		return model.Vec3{}
	}
	return disp[idx]
}

// component 返回 Vec3 的第 c 分量。
func component(v model.Vec3, c int) float64 {
	switch c {
	case cX:
		return v.X
	case cY:
		return v.Y
	case cZ:
		return v.Z
	}
	return 0
}
