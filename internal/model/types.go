// Package model 定义医学影像配准形变场质量门服务的领域实体、状态常量、
// 稳定错误与基础向量类型。
package model

import "math"

// Vec3 表示三维向量或坐标；二维场时 Z 分量恒为 0。
// 位移场中每个体素对应一个 Vec3（dx, dy, dz），单位为毫米（mm）。
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// IsFinite 报告向量三分量是否均为有限值（排除 NaN 与 ±Inf）。
// 质量门据此把"缺值/越界"的位移体素判为无效覆盖。
func (v Vec3) IsFinite() bool {
	return !math.IsNaN(v.X) && !math.IsInf(v.X, 0) &&
		!math.IsNaN(v.Y) && !math.IsInf(v.Y, 0) &&
		!math.IsNaN(v.Z) && !math.IsInf(v.Z, 0)
}

// Add 返回向量相加结果。
func (v Vec3) Add(o Vec3) Vec3 { return Vec3{X: v.X + o.X, Y: v.Y + o.Y, Z: v.Z + o.Z} }

// Sub 返回向量相减结果。
func (v Vec3) Sub(o Vec3) Vec3 { return Vec3{X: v.X - o.X, Y: v.Y - o.Y, Z: v.Z - o.Z} }

// Norm 返回欧氏范数。
func (v Vec3) Norm() float64 { return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z) }

// Dims 表示影像体素维度，例如 [nx, ny, nz] 或二维 [nx, ny]。
type Dims []int

// Volume 返回体素总数（各维乘积）。
func (d Dims) Volume() int {
	if len(d) == 0 {
		return 0
	}
	n := 1
	for _, s := range d {
		if s <= 0 {
			return 0
		}
		n *= s
	}
	return n
}

// Rank 返回维度（2 或 3）。
func (d Dims) Rank() int {
	if len(d) < 2 {
		return 0
	}
	if len(d) >= 3 && d[2] > 1 {
		return 3
	}
	return 2
}

// Equal 报告两个维度序列是否完全一致。
func (d Dims) Equal(o Dims) bool {
	if len(d) != len(o) {
		return false
	}
	for i := range d {
		if d[i] != o[i] {
			return false
		}
	}
	return true
}
