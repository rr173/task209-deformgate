package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"task209-deformgate/internal/model"
)

// errorResponse 是结构化错误响应体。
type errorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 把业务错误映射为 HTTP 状态码并输出结构化错误。
func writeError(w http.ResponseWriter, err error) {
	status, code := mapError(err)
	writeJSON(w, status, errorResponse{Error: err.Error(), Code: code})
}

// mapError 把领域错误映射为 HTTP 状态码与业务错误码。
func mapError(err error) (int, string) {
	var se *model.StatusError
	if errors.As(err, &se) {
		return se.Code, "status_error"
	}
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, "not_found"
	case errors.Is(err, model.ErrInvalid):
		return http.StatusBadRequest, "invalid_input"
	case errors.Is(err, model.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, model.ErrForbidden):
		return http.StatusForbidden, "forbidden"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}

// decodeJSON 解析请求体 JSON，并限制体大小（1 MiB）。
func decodeJSON(r *http.Request, v any) error {
	r.Body = io.NopCloser(io.LimitReader(r.Body, 1<<20))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// pathID 从路径变量提取 int64 ID。
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, model.ErrInvalid
	}
	return id, nil
}
