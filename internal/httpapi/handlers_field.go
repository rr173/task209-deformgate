package httpapi

import (
	"net/http"

	"task209-deformgate/internal/model"
)

type attachFieldReq struct {
	Dims          model.Dims   `json:"dims"`
	Displacements []model.Vec3 `json:"displacements"`
}

func (s *Server) attachField(w http.ResponseWriter, r *http.Request) {
	pairID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req attachFieldReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, model.ErrInvalid)
		return
	}
	f, err := s.app.AttachField(pairID, req.Dims, req.Displacements)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

func (s *Server) listFields(w http.ResponseWriter, r *http.Request) {
	pairID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	fs, err := s.app.ListFields(pairID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, fs)
}

func (s *Server) getField(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	f, err := s.app.GetField(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (s *Server) getFieldHash(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	f, err := s.app.GetField(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"hash": f.Hash})
}

type addSampleReq struct {
	Pos     model.Vec3 `json:"pos"`
	Forward model.Vec3 `json:"forward"`
	Inverse model.Vec3 `json:"inverse"`
}

func (s *Server) addSample(w http.ResponseWriter, r *http.Request) {
	fieldID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req addSampleReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, model.ErrInvalid)
		return
	}
	sp, err := s.app.AddSample(fieldID, req.Pos, req.Forward, req.Inverse)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sp)
}

func (s *Server) listSamples(w http.ResponseWriter, r *http.Request) {
	fieldID, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	sps, err := s.app.ListSamples(fieldID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sps)
}
