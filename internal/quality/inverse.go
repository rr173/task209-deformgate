package quality

import "task209-deformgate/internal/model"

// ComputeInverseStats 汇总逆一致性误差统计：
// 均值、最大值、超阈值采样点数与占比。
// points 的 Error 字段应已由 field.ComputeSampleErrors 计算。
func ComputeInverseStats(points []model.SamplePoint, threshold float64) model.InverseStats {
	st := model.InverseStats{SampleCount: len(points)}
	if len(points) == 0 {
		return st
	}
	var sum, max float64
	for _, p := range points {
		e := p.Error
		sum += e
		if e > max {
			max = e
		}
		if e > threshold {
			st.ExceedCount++
		}
	}
	st.Mean = sum / float64(len(points))
	st.Max = max
	st.ExceedRatio = float64(st.ExceedCount) / float64(len(points))
	return st
}
