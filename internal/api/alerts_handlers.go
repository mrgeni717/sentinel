package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mrgeni717/sentinel/internal/model"
)

func (s *Server) listAlertRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.ListAlertRules()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

func (s *Server) createAlertRule(w http.ResponseWriter, r *http.Request) {
	var req model.AlertRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Name == "" || req.TargetType == "" || req.Metric == "" || req.Operator == "" {
		writeError(w, http.StatusBadRequest, "name, targetType, metric, and operator are required")
		return
	}
	req.Enabled = true

	rule, err := s.store.CreateAlertRule(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (s *Server) deleteAlertRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.store.DeleteAlertRule(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, err := s.store.ListAlerts(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, alerts)
}
