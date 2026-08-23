package service

import (
	"errors"

	"task209-deformgate/internal/field"
	"task209-deformgate/internal/gating"
	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// AttachField 向影像对追加形变场：校验维度与位移、计算内容哈希，
// 幂等去重后入库，并把影像对推进到 pending_validation。
func (a *App) AttachField(pairID int64, dims model.Dims, displacements []model.Vec3) (*model.DeformField, error) {
	pair, err := a.st.GetImagePair(pairID)
	if err != nil {
		return nil, err
	}
	if !gating.CanAttachField(pair.Status) {
		return nil, model.ErrForbidden
	}
	if err := imaging.ValidateFieldDims(pair.Dims, dims, len(displacements)); err != nil {
		return nil, err
	}
	if err := field.ValidateDisplacements(dims, displacements); err != nil {
		return nil, err
	}

	hash := field.HashField(dims, displacements)
	if hash == "" {
		return nil, model.ErrInternal
	}
	// 幂等：相同内容形变场返回既有记录。
	if existing, err := a.st.FieldByHash(hash); err == nil {
		return existing, nil
	} else if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	f := &model.DeformField{
		ImagePairID:   pairID,
		Hash:          hash,
		Dims:          dims,
		Displacements: displacements,
	}
	if err := a.st.CreateField(f); err != nil {
		return nil, err
	}
	// 影像对进入待校验状态。
	if pair.Status == model.PairPending {
		_ = a.st.SetImagePairStatus(pairID, model.PairPending)
	}
	return f, nil
}

// ListFields 返回影像对下全部形变场。
func (a *App) ListFields(pairID int64) ([]model.DeformField, error) {
	if _, err := a.st.GetImagePair(pairID); err != nil {
		return nil, err
	}
	return a.st.ListFieldsByPair(pairID)
}

// GetField 读取形变场详情。
func (a *App) GetField(id int64) (*model.DeformField, error) {
	return a.st.GetField(id)
}

// AddSample 向形变场添加逆一致性采样点：校验采样位置越界，
// 计算逆一致性误差后入库。
func (a *App) AddSample(fieldID int64, pos, fwd, inv model.Vec3) (*model.SamplePoint, error) {
	f, err := a.st.GetField(fieldID)
	if err != nil {
		return nil, err
	}
	if err := imaging.ValidateSamplePosition(pos, f.Dims); err != nil {
		return nil, err
	}
	sp := &model.SamplePoint{
		FieldID: fieldID,
		Pos:     pos,
		Forward: fwd,
		Inverse: inv,
		Error:   field.InverseConsistencyError(fwd, inv),
		Status:  model.SampleChecked,
	}
	if err := a.st.CreateSample(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

// ListSamples 返回形变场下全部采样点。
func (a *App) ListSamples(fieldID int64) ([]model.SamplePoint, error) {
	if _, err := a.st.GetField(fieldID); err != nil {
		return nil, err
	}
	return a.st.ListSamplesByField(fieldID)
}
