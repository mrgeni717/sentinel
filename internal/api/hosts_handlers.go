package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mrgeni717/sentinel/internal/model"
)

func (s *Server) listHosts(w http.ResponseWriter, r *http.Request) {
	hosts, err := s.store.ListHosts()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type hostWithStatus struct {
		model.Host
		Latest *model.HostMetric `json:"latest,omitempty"`
	}
	result := make([]hostWithStatus, 0, len(hosts))
	for _, h := range hosts {
		latest, _ := s.store.LatestHostMetric(h.ID)
		result = append(result, hostWithStatus{Host: h, Latest: latest})
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) hostHistory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	history, err := s.store.HostMetricHistory(id, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, history)
}

// ingestMetrics is what the agent binary POSTs to. Protected by a shared
// secret (X-Ingest-Key header) rather than full user auth, since this is
// a machine-to-machine endpoint, not something a browser user calls.
func (s *Server) ingestMetrics(w http.ResponseWriter, r *http.Request) {
	if s.ingestKey != "" && r.Header.Get("X-Ingest-Key") != s.ingestKey {
		writeError(w, http.StatusUnauthorized, "invalid or missing ingest key")
		return
	}

	var req model.MetricPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.HostName == "" {
		writeError(w, http.StatusBadRequest, "hostName is required")
		return
	}

	host, err := s.store.UpsertHost(req.HostName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metric := model.HostMetric{
		HostID:        host.ID,
		CPUPercent:    req.CPUPercent,
		MemoryPercent: req.MemoryPercent,
		DiskPercent:   req.DiskPercent,
		LoadAvg1:      req.LoadAvg1,
	}
	if err := s.store.RecordHostMetric(metric); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if s.engine != nil {
		go s.engine.EvaluateHost(host.ID)
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "recorded"})
}
