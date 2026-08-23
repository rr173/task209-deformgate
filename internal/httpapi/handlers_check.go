package httpapi

import (
	"net/http"

	"task209-deformgate/internal/model"
)

type createCheckReq struct {
	ImagePairID int64 `json:"image_pair_id"`
	FieldID     int64 `json:"field_id"`
}

func (s *Server) createCheck(w http.ResponseWriter, r *http.Request) {
	var req createCheckReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, model.ErrInvalid)
		return
	}
	t, err := s.app.CreateCheck(req.ImagePairID, req.FieldID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) listChecks(w http.ResponseWriter, r *http.Request) {
	ts, err := s.app.ListChecks()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ts)
}

func (s *Server) getCheck(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	t, err := s.app.GetCheck(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) runCheck(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	res, err := s.app.RunCheck(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	res, err := s.app.GetMetrics(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) getFolds(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	folds, err := s.app.GetFolds(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, folds)
}
