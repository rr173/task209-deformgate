// Package field 负责形变场的内容处理：确定性哈希（幂等键）、
// 位移向量有效性检查与逆一致性采样误差计算，以及测试/自检用的场合成。
package field

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"task209-deformgate/internal/model"
)

// HashField 计算形变场的内容哈希（SHA-256）。
// 输入为维度与位移向量序列，序列化后哈希，用作幂等去重键：
// 相同内容的形变场重复上传返回同一记录，不重复入库。
func HashField(dims model.Dims, disp []model.Vec3) string {
	payload := struct {
		Dims model.Dims   `json:"dims"`
		Disp []model.Vec3 `json:"disp"`
	}{Dims: dims, Disp: disp}
	b, err := json.Marshal(payload)
	if err != nil {
		// 不可能发生（仅含数字）；兜底返回空哈希。
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ValidateDisplacements 校验位移向量序列：数量与体素总数一致，且均为有限值。
func ValidateDisplacements(dims model.Dims, disp []model.Vec3) error {
	if len(disp) != dims.Volume() {
		return model.ErrInvalid
	}
	for _, d := range disp {
		if !d.IsFinite() {
			return model.ErrInvalid
		}
	}
	return nil
}

// InverseConsistencyError 计算采样点的逆一致性误差：
// 正变换位移 f 与映射点处逆变换位移 b 之和的范数。
// 理想配准下 b = -f，误差为 0；误差越大说明正逆变换越不自洽。
func InverseConsistencyError(f, b model.Vec3) float64 {
	return f.Add(b).Norm()
}
