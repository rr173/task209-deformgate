package gating

import (
	"fmt"

	"task209-deformgate/internal/model"
)

// VersionChain 描述一个影像对的发布版本链：当前发布结果与全部历史结果。
type VersionChain struct {
	ImagePairID int64               `json:"image_pair_id"`
	Published   *model.QualityResult `json:"published,omitempty"`
	History     []model.QualityResult `json:"history"`
}

// BuildVersionChain 依据结果列表构建版本链：当前 published 结果 +
// 按时间序的历史结果（含 superseded / draft）。
func BuildVersionChain(pairID int64, results []model.QualityResult) VersionChain {
	chain := VersionChain{ImagePairID: pairID, History: []model.QualityResult{}}
	for _, r := range results {
		switch r.Status {
		case model.ResultPublished:
			cp := r
			chain.Published = &cp
		case model.ResultSuperseded:
			chain.History = append(chain.History, r)
		default:
			chain.History = append(chain.History, r)
		}
	}
	return chain
}

// ValidateVersionTransition 校验版本状态迁移是否合法：
// draft → published（发布）；published → superseded（替代）；
// 其它迁移一律拒绝，保证发布链不可变。
func ValidateVersionTransition(from, to string) error {
	switch {
	case from == model.ResultDraft && to == model.ResultPublished:
		return nil
	case from == model.ResultPublished && to == model.ResultSuperseded:
		return nil
	case from == to:
		return nil
	default:
		return fmt.Errorf("%w: 非法版本状态迁移 %s → %s", model.ErrConflict, from, to)
	}
}
