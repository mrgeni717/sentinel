package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mrgeni717/sentinel/internal/model"
)

type createTargetRequest struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds int    `json:"intervalSeconds"`
}

func (s *Server) listTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := s.store.ListTargets()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Enrich each target with its latest check so the dashboard doesn't
	// need a second round trip per row.
	type targetWithStatus struct {
		model.Target
		Latest *model.UptimeCheck `json:"latest,omitempty"`
	}
	result := make([]targetWithStatus, 0, len(targets))
	for _, t := range targets {
		latest, _ := s.store.LatestUptimeCheck(t.ID) // nil if no checks yet, fine to ignore
		result = append(result, targetWithStatus{Target: t, Latest: latest})
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) createTarget(w http.ResponseWriter, r *http.Request) {
	var req createTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" || req.URL == "" {
		writeError(w, http.StatusBadRequest, "name and url are required")
		return
	}
	if req.IntervalSeconds <= 0 {
		req.IntervalSeconds = 60
	}

	target, err := s.store.CreateTarget(req.Name, req.URL, req.IntervalSeconds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, target)
}

func (s *Server) deleteTarget(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.store.DeleteTarget(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) targetHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	history, err := s.store.UptimeHistory(id, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, history)
}
