package httpapi

import (
	"net/http"

	"task209-deformgate/internal/model"
)

func (s *Server) listParams(w http.ResponseWriter, r *http.Request) {
	ps, err := s.app.ListParams()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

func (s *Server) getActiveParams(w http.ResponseWriter, r *http.Request) {
	p, err := s.app.GetActiveParams()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type createParamsReq struct {
	Name              string  `json:"name"`
	JacMin            float64 `json:"jac_min"`
	JacMax            float64 `json:"jac_max"`
	ICEThreshold      float64 `json:"ice_threshold"`
	CoverageThreshold float64 `json:"coverage_threshold"`
}

func (s *Server) createParams(w http.ResponseWriter, r *http.Request) {
	var req createParamsReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, model.ErrInvalid)
		return
	}
	p, err := s.app.CreateParams(req.Name, req.JacMin, req.JacMax, req.ICEThreshold, req.CoverageThreshold)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) activateParams(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.app.ActivateParams(id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}
