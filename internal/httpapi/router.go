// Package httpapi 是 HTTP 层：路由（前缀 /api）、JSON 编解码、错误映射与中间件。
package httpapi

import (
	"net/http"

	"task209-deformgate/internal/service"
)

// Server 是 HTTP 服务，持有编排层引用。
type Server struct {
	app *service.App
}

// New 构造 HTTP 服务。
func New(app *service.App) *Server {
	return &Server{app: app}
}

// Handler 注册全部路由并套用中间件。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 影像对。
	mux.HandleFunc("POST /api/imagepairs", s.createImagePair)
	mux.HandleFunc("GET /api/imagepairs", s.listImagePairs)
	mux.HandleFunc("GET /api/imagepairs/{id}", s.getImagePair)
	mux.HandleFunc("PUT /api/imagepairs/{id}", s.updateImagePair)
	mux.HandleFunc("POST /api/imagepairs/{id}/archive", s.archiveImagePair)
	mux.HandleFunc("GET /api/imagepairs/{id}/results", s.listPairResults)
	mux.HandleFunc("GET /api/imagepairs/{id}/version-chain", s.versionChain)

	// 形变场。
	mux.HandleFunc("POST /api/imagepairs/{id}/fields", s.attachField)
	mux.HandleFunc("GET /api/imagepairs/{id}/fields", s.listFields)
	mux.HandleFunc("GET /api/fields/{id}", s.getField)
	mux.HandleFunc("GET /api/fields/{id}/hash", s.getFieldHash)
	mux.HandleFunc("POST /api/fields/{id}/samples", s.addSample)
	mux.HandleFunc("GET /api/fields/{id}/samples", s.listSamples)

	// 检查任务。
	mux.HandleFunc("POST /api/checks", s.createCheck)
	mux.HandleFunc("GET /api/checks", s.listChecks)
	mux.HandleFunc("GET /api/checks/{id}", s.getCheck)
	mux.HandleFunc("POST /api/checks/{id}/run", s.runCheck)
	mux.HandleFunc("GET /api/checks/{id}/metrics", s.getMetrics)
	mux.HandleFunc("GET /api/checks/{id}/folds", s.getFolds)

	// 检查参数。
	mux.HandleFunc("GET /api/params", s.listParams)
	mux.HandleFunc("GET /api/params/active", s.getActiveParams)
	mux.HandleFunc("POST /api/params", s.createParams)
	mux.HandleFunc("POST /api/params/{id}/activate", s.activateParams)

	// 质量结果。
	mux.HandleFunc("POST /api/results/{id}/publish", s.publishResult)
	mux.HandleFunc("GET /api/results", s.listResults)
	mux.HandleFunc("GET /api/results/{id}", s.getResult)

	// 元信息。
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/stats", s.stats)
	mux.HandleFunc("GET /api/version", s.version)

	return withRecover(withLogging(mux))
}
