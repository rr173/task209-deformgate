package httpapi

import "net/http"

// versionInfo 是服务版本信息。
type versionInfo struct {
	Service string `json:"service"`
	Version string `json:"version"`
	Go      string `json:"go"`
	SQLite  string `json:"sqlite"`
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	st, err := s.app.Store().Stats()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, versionInfo{
		Service: "task209-deformgate",
		Version: "1.0.0",
		Go:      "1.26.3",
		SQLite:  "3.46.1",
	})
}
