package httpapi

import (
	"net/http"

	"task209-deformgate/internal/model"
)

type createImagePairReq struct {
	Name     string     `json:"name"`
	Dims     model.Dims `json:"dims"`
	Spacing  []float64  `json:"spacing"`
	Axes     string     `json:"axes"`
	Modality string     `json:"modality"`
}

func (s *Server) createImagePair(w http.ResponseWriter, r *http.Request) {
	var req createImagePairReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, model.ErrInvalid)
		return
	}
	p, err := s.app.CreateImagePair(req.Name, req.Dims, req.Spacing, req.Axes, req.Modality)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) listImagePairs(w http.ResponseWriter, r *http.Request) {
	ps, err := s.app.ListImagePairs()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

func (s *Server) getImagePair(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	p, err := s.app.GetImagePair(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

type updateImagePairReq struct {
	Name     string `json:"name"`
	Axes     string `json:"axes"`
	Modality string `json:"modality"`
}

func (s *Server) updateImagePair(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req updateImagePairReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, model.ErrInvalid)
		return
	}
	if err := s.app.UpdateImagePairMeta(id, req.Name, req.Axes, req.Modality); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

func (s *Server) archiveImagePair(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := s.app.ArchiveImagePair(id); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

func (s *Server) listPairResults(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	results, err := s.app.ListResultsByPair(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *Server) versionChain(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	chain, err := s.app.VersionChain(id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, chain)
}
