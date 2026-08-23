package service

import (
	"task209-deformgate/internal/gating"
	"task209-deformgate/internal/model"
	"task209-deformgate/internal/quality"
)

// CreateCheck 创建检查任务：绑定激活参数，校验影像对与形变场维度匹配。
// 同一 (影像对, 形变场) 只允许一个活动任务（部分唯一索引保证）。
func (a *App) CreateCheck(pairID, fieldID int64) (*model.CheckTask, error) {
	pair, err := a.st.GetImagePair(pairID)
	if err != nil {
		return nil, err
	}
	if pair.Status == model.PairArchived {
		return nil, model.ErrForbidden
	}
	f, err := a.st.GetField(fieldID)
	if err != nil {
		return nil, err
	}
	if f.ImagePairID != pairID {
		return nil, model.ErrInvalid
	}
	if !pair.Dims.Equal(f.Dims) {
		return nil, model.ErrInvalid
	}
	params, err := a.st.GetActiveParams()
	if err != nil {
		return nil, err
	}
	t := &model.CheckTask{
		ImagePairID: pairID,
		FieldID:     fieldID,
		ParamsID:    params.ID,
	}
	if err := a.st.CreateTask(t); err != nil {
		return nil, err
	}
	return t, nil
}

// ListChecks 返回全部检查任务。
func (a *App) ListChecks() ([]model.CheckTask, error) {
	return a.st.ListTasks()
}

// GetCheck 读取检查任务详情。
func (a *App) GetCheck(id int64) (*model.CheckTask, error) {
	return a.st.GetTask(id)
}

// RunCheck 执行质量评估：queued → running → completed，
// 计算三组指标、产出质量结果、更新形变场与影像对状态。
func (a *App) RunCheck(id int64) (*model.QualityResult, error) {
	t, err := a.st.GetTask(id)
	if err != nil {
		return nil, err
	}
	if t.Status != model.TaskQueued {
		return nil, model.ErrConflict
	}
	if err := a.st.SetTaskStatus(id, model.TaskRunning); err != nil {
		return nil, err
	}

	result, err := a.runEvaluation(t)
	if err != nil {
		_ = a.st.SetTaskStatus(id, model.TaskFailed)
		_ = a.st.SetTaskMessage(id, err.Error())
		return nil, err
	}

	if err := a.st.SetTaskStatus(id, model.TaskCompleted); err != nil {
		return nil, err
	}
	return result, nil
}

// runEvaluation 执行具体的指标计算与结果落库。
func (a *App) runEvaluation(t *model.CheckTask) (*model.QualityResult, error) {
	f, err := a.st.GetField(t.FieldID)
	if err != nil {
		return nil, err
	}
	params, err := a.st.GetParams(t.ParamsID)
	if err != nil {
		return nil, err
	}
	samples, err := a.st.ListSamplesByField(t.FieldID)
	if err != nil {
		return nil, err
	}

	res := quality.Evaluate(f.Displacements, f.Dims, samples, *params)
	res.TaskID = t.ID
	res.ImagePairID = t.ImagePairID
	res.FieldID = t.FieldID

	if err := a.st.CreateResult(&res); err != nil {
		return nil, err
	}
	if err := a.st.SaveFolds(res.ID, res.Folds); err != nil {
		return nil, err
	}

	// 更新形变场状态：折叠 / 缺边界 / 有效。
	switch {
	case res.Jacobian.FoldCount > 0:
		_ = a.st.SetFieldStatus(t.FieldID, model.FieldFolded)
	case res.Boundary.Missing > 0:
		_ = a.st.SetFieldStatus(t.FieldID, model.FieldMissingBoundary)
	default:
		_ = a.st.SetFieldStatus(t.FieldID, model.FieldValid)
	}

	// 更新影像对状态：按判定映射。
	_ = a.st.SetImagePairStatus(t.ImagePairID, gating.VerdictToPairStatus(res.Verdict))

	return &res, nil
}

// GetMetrics 返回检查任务产出的三组指标。
func (a *App) GetMetrics(taskID int64) (*model.QualityResult, error) {
	return a.st.GetResultByTask(taskID)
}

// GetFolds 返回检查任务产出的折叠证据。
func (a *App) GetFolds(taskID int64) ([]model.FoldRegion, error) {
	res, err := a.st.GetResultByTask(taskID)
	if err != nil {
		return nil, err
	}
	return a.st.ListFolds(res.ID)
}
