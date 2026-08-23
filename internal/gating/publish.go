// Package gating 负责质量门的裁决与发布规则：把质量判定映射为影像对状态，
// 并校验结果发布的前置条件，保证发布链不可变、可追溯。
package gating

import (
	"fmt"

	"task209-deformgate/internal/model"
)

// VerdictToPairStatus 把质量判定映射为影像对状态。
func VerdictToPairStatus(verdict string) string {
	switch verdict {
	case model.VerdictPass:
		return model.PairPassed
	case model.VerdictReview:
		return model.PairReview
	case model.VerdictReject:
		return model.PairRejected
	default:
		return model.PairPending
	}
}

// ValidatePublish 校验结果能否发布：
// 结果必须为 draft；关联影像对不得处于封存状态。
// 违反时返回携带 model.ErrForbidden / model.ErrConflict 的错误。
func ValidatePublish(result model.QualityResult, pairStatus string) error {
	if result.Status != model.ResultDraft {
		return fmt.Errorf("%w: 结果状态为 %s，仅草稿可发布", model.ErrConflict, result.Status)
	}
	if pairStatus == model.PairArchived {
		return fmt.Errorf("%w: 影像对已封存，禁止发布新结果", model.ErrForbidden)
	}
	return nil
}

// ValidateSupersede 校验显式替代操作：被替代结果必须为 published 状态。
func ValidateSupersede(target model.QualityResult) error {
	if target.Status != model.ResultPublished {
		return fmt.Errorf("%w: 被替代结果状态为 %s，仅已发布结果可被替代", model.ErrConflict, target.Status)
	}
	return nil
}

// CanModifyPair 判断影像对是否仍可编辑元数据（封存后禁止）。
func CanModifyPair(status string) bool {
	return status != model.PairArchived
}

// CanAttachField 判断影像对是否仍可追加形变场（封存后禁止）。
func CanAttachField(status string) bool {
	return status != model.PairArchived
}
