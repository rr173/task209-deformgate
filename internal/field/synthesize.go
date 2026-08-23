package field

import (
	"task209-deformgate/internal/model"
)

// SynthesizeIdentity 生成一个恒等形变场（位移全零），用于测试与基线。
// 恒等场雅可比行列式恒为 1，边界覆盖完整，逆一致性误差为 0。
func SynthesizeIdentity(dims model.Dims) []model.Vec3 {
	n := dims.Volume()
	out := make([]model.Vec3, n)
	return out
}

// SynthesizeAffine 生成一个仿射形变场：位移 = scale * 坐标 + bias。
// 用于构造可预测的雅可比行列式（约等于 (1+scale)^rank）。
func SynthesizeAffine(dims model.Dims, scale float64, bias model.Vec3) []model.Vec3 {
	n := dims.Volume()
	out := make([]model.Vec3, n)
	for i := 0; i < n; i++ {
		pos := indexToPos(i, dims)
		out[i] = model.Vec3{
			X: scale*pos.X + bias.X,
			Y: scale*pos.Y + bias.Y,
			Z: scale*pos.Z + bias.Z,
		}
	}
	return out
}

// SynthesizeLocalFold 生成一个在指定体素位置制造折叠的形变场：
// 该体素邻域位移反向，产生 det(J) ≤ 0 的折叠证据，其余为恒等。
func SynthesizeLocalFold(dims model.Dims, center int) []model.Vec3 {
	n := dims.Volume()
	out := make([]model.Vec3, n)
	cx := indexToPos(center, dims)
	for i := 0; i < n; i++ {
		p := indexToPos(i, dims)
		dx := p.X - cx.X
		dy := p.Y - cx.Y
		dz := p.Z - cx.Z
		dist2 := dx*dx + dy*dy + dz*dz
		if dist2 <= 2.0 { // 中心体素及其近邻制造反向位移
			out[i] = model.Vec3{X: -2 * dx, Y: -2 * dy, Z: -2 * dz}
		}
	}
	return out
}

// indexToPos 把展平体素序号还原为坐标。
func indexToPos(index int, dims model.Dims) model.Vec3 {
	nx := dims[0]
	ny := dims[1]
	iz := index / (nx * ny)
	rem := index % (nx * ny)
	iy := rem / nx
	ix := rem % nx
	return model.Vec3{X: float64(ix), Y: float64(iy), Z: float64(iz)}
}
