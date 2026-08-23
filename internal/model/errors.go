package model

import "errors"

// 稳定错误码：HTTP 层据此映射到 4xx/5xx 与结构化错误响应。
var (
	// ErrNotFound 目标实体不存在。
	ErrNotFound = errors.New("not found")
	// ErrConflict 与现有状态冲突（如重复活动任务、非法状态迁移）。
	ErrConflict = errors.New("conflict")
	// ErrInvalid 输入数据不合法（维度不匹配、坐标方向未声明、采样点越界等）。
	ErrInvalid = errors.New("invalid input")
	// ErrForbidden 尝试修改封存/已发布对象。
	ErrForbidden = errors.New("operation forbidden on archived or published object")
	// ErrInternal 内部错误（数据库、计算失败等）。
	ErrInternal = errors.New("internal error")
)

// StatusError 携带 HTTP 状态码与业务错误码，供 HTTP 层直接映射。
type StatusError struct {
	Code int
	Err  error
}

func (e *StatusError) Error() string { return e.Err.Error() }
func (e *StatusError) Unwrap() error { return e.Err }

// NewStatusError 构造带状态码的错误。
func NewStatusError(code int, err error) *StatusError {
	return &StatusError{Code: code, Err: err}
}
