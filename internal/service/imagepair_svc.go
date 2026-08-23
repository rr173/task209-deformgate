package service

import (
	"task209-deformgate/internal/gating"
	"task209-deformgate/internal/imaging"
	"task209-deformgate/internal/model"
)

// CreateImagePair 登记影像对：校验几何元数据后入库（状态 registered）。
func (a *App) CreateImagePair(name string, dims model.Dims, spacing []float64, axes, modality string) (*model.ImagePair, error) {
	p := &model.ImagePair{
		Name:     name,
		Dims:     dims,
		Spacing:  spacing,
		Axes:     axes,
		Modality: modality,
	}
	if err := imaging.ValidatePair(p); err != nil {
		return nil, err
	}
	if err := a.st.CreateImagePair(p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListImagePairs 返回全部影像对。
func (a *App) ListImagePairs() ([]model.ImagePair, error) {
	return a.st.ListImagePairs()
}

// GetImagePair 读取影像对详情。
func (a *App) GetImagePair(id int64) (*model.ImagePair, error) {
	return a.st.GetImagePair(id)
}

// UpdateImagePairMeta 更新影像对可编辑元数据（封存后拒绝）。
func (a *App) UpdateImagePairMeta(id int64, name, axes, modality string) error {
	p, err := a.st.GetImagePair(id)
	if err != nil {
		return err
	}
	if !gating.CanModifyPair(p.Status) {
		return model.ErrForbidden
	}
	if err := imaging.ValidateAxes(axes); err != nil {
		return err
	}
	return a.st.UpdateImagePairMeta(id, name, axes, modality)
}

// ArchiveImagePair 封存影像对：终态，不可再修改或追加形变场。
func (a *App) ArchiveImagePair(id int64) error {
	p, err := a.st.GetImagePair(id)
	if err != nil {
		return err
	}
	if p.Status == model.PairArchived {
		return model.ErrConflict
	}
	return a.st.SetImagePairStatus(id, model.PairArchived)
}
