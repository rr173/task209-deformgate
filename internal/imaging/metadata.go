// Package imaging 负责影像几何元数据与坐标系统的校验，
// 是质量门的第一道输入把关：维度、体素间距、坐标方向约定。
package imaging

import (
	"fmt"
	"strings"

	"task209-deformgate/internal/model"
)

// axisLetters 是合法的坐标方向字母集合（解剖学标准方向）。
// R/L 左右、A/P 前后、S/I 上下。
var axisLetters = map[rune]bool{
	'R': true, 'L': true,
	'A': true, 'P': true,
	'S': true, 'I': true,
}

// ValidateDims 校验影像维度：2 或 3 维，且各维为正整数。
func ValidateDims(d model.Dims) error {
	if len(d) < 2 || len(d) > 3 {
		return fmt.Errorf("%w: dims 必须为 2 或 3 维，得到 %d 维", model.ErrInvalid, len(d))
	}
	for i, s := range d {
		if s <= 0 {
			return fmt.Errorf("%w: dims[%d]=%d 必须为正", model.ErrInvalid, i, s)
		}
	}
	return nil
}

// ValidateSpacing 校验体素间距：长度与维度一致且为正。
func ValidateSpacing(dims model.Dims, spacing []float64) error {
	if len(spacing) != len(dims) {
		return fmt.Errorf("%w: spacing 长度 %d 与 dims 维度 %d 不一致", model.ErrInvalid, len(spacing), len(dims))
	}
	for i, v := range spacing {
		if v <= 0 {
			return fmt.Errorf("%w: spacing[%d]=%v 必须为正", model.ErrInvalid, i, v)
		}
	}
	return nil
}

// ValidateAxes 校验坐标方向约定：必须显式声明，且为合法方向字母组合。
// 坐标方向未声明属于 REQ 明确要求的拒绝条件。
func ValidateAxes(axes string) error {
	axes = strings.TrimSpace(strings.ToUpper(axes))
	if axes == "" {
		return fmt.Errorf("%w: 坐标方向 axes 未声明", model.ErrInvalid)
	}
	for _, c := range axes {
		if !axisLetters[c] {
			return fmt.Errorf("%w: 非法坐标方向字母 %q", model.ErrInvalid, c)
		}
	}
	return nil
}

// ValidatePair 综合校验影像对几何元数据。
func ValidatePair(p *model.ImagePair) error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: 影像对名称不能为空", model.ErrInvalid)
	}
	if err := ValidateDims(p.Dims); err != nil {
		return err
	}
	if err := ValidateSpacing(p.Dims, p.Spacing); err != nil {
		return err
	}
	return ValidateAxes(p.Axes)
}
