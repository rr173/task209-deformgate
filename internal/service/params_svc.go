package service

import (
	"task209-deformgate/internal/model"
)

// GetActiveParams 返回当前激活的检查参数版本。
func (a *App) GetActiveParams() (*model.CheckParams, error) {
	return a.st.GetActiveParams()
}

// ListParams 返回全部参数版本。
func (a *App) ListParams() ([]model.CheckParams, error) {
	return a.st.ListParams()
}

// CreateParams 创建新的参数版本（默认未激活）。
func (a *App) CreateParams(name string, jacMin, jacMax, iceThreshold, coverageThreshold float64) (*model.CheckParams, error) {
	if jacMin <= 0 || jacMax <= jacMin {
		return nil, model.ErrInvalid
	}
	if iceThreshold < 0 {
		return nil, model.ErrInvalid
	}
	if coverageThreshold < 0 || coverageThreshold > 1 {
		return nil, model.ErrInvalid
	}
	p := &model.CheckParams{
		Name:              name,
		JacMin:            jacMin,
		JacMax:            jacMax,
		ICEThreshold:      iceThreshold,
		CoverageThreshold: coverageThreshold,
	}
	if err := a.st.CreateParams(p); err != nil {
		return nil, err
	}
	return p, nil
}

// ActivateParams 激活指定参数版本（锁定后不可改，仅新建版本再激活）。
func (a *App) ActivateParams(id int64) error {
	return a.st.ActivateParams(id)
}
