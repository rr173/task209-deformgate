package service

import (
	"task209-deformgate/internal/gating"
	"task209-deformgate/internal/model"
)

// PublishResult 发布质量结果：校验前置条件后，把该影像对既有 published
// 结果显式替代，新结果绑定 supersedes 指向被替代结果。
func (a *App) PublishResult(id int64) (*model.QualityResult, error) {
	res, err := a.st.GetResult(id)
	if err != nil {
		return nil, err
	}
	pair, err := a.st.GetImagePair(res.ImagePairID)
	if err != nil {
		return nil, err
	}
	if err := gating.ValidatePublish(*res, pair.Status); err != nil {
		return nil, err
	}
	if err := a.st.PublishResult(id); err != nil {
		return nil, err
	}
	// 发布后影像对状态与判定保持一致。
	_ = a.st.SetImagePairStatus(res.ImagePairID, gating.VerdictToPairStatus(res.Verdict))
	return a.st.GetResult(id)
}

// ListResultsByPair 返回影像对下全部质量结果。
func (a *App) ListResultsByPair(pairID int64) ([]model.QualityResult, error) {
	if _, err := a.st.GetImagePair(pairID); err != nil {
		return nil, err
	}
	return a.st.ListResultsByPair(pairID)
}

// ListResults 返回全部质量结果。
func (a *App) ListResults() ([]model.QualityResult, error) {
	return a.st.ListResults()
}

// GetResult 读取质量结果详情。
func (a *App) GetResult(id int64) (*model.QualityResult, error) {
	return a.st.GetResult(id)
}

// VersionChain 返回影像对的版本链（当前发布 + 历史）。
func (a *App) VersionChain(pairID int64) (gating.VersionChain, error) {
	results, err := a.st.ListResultsByPair(pairID)
	if err != nil {
		return gating.VersionChain{}, err
	}
	return gating.BuildVersionChain(pairID, results), nil
}
