package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mrgeni717/sentinel/internal/alertengine"
	"github.com/mrgeni717/sentinel/internal/store"
)

type Server struct {
	store     *store.Store
	engine    *alertengine.Engine
	ingestKey string
}

func NewServer(s *store.Store, e *alertengine.Engine, ingestKey string) *Server {
	return &Server{store: s, engine: e, ingestKey: ingestKey}
}

// Routes returns the full mux: API handlers plus the static dashboard.
// staticDir is the filesystem path to serve the web dashboard from.
func (s *Server) Routes(staticDir string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/targets", s.listTargets)
	mux.HandleFunc("POST /api/targets", s.createTarget)
	mux.HandleFunc("DELETE /api/targets/{id}", s.deleteTarget)
	mux.HandleFunc("GET /api/targets/{id}/checks", s.targetHistory)

	mux.HandleFunc("GET /api/hosts", s.listHosts)
	mux.HandleFunc("GET /api/hosts/{id}/metrics", s.hostHistory)
	mux.HandleFunc("POST /api/ingest/metrics", s.ingestMetrics)

	mux.HandleFunc("GET /api/alert-rules", s.listAlertRules)
	mux.HandleFunc("POST /api/alert-rules", s.createAlertRule)
	mux.HandleFunc("DELETE /api/alert-rules/{id}", s.deleteAlertRule)

	mux.HandleFunc("GET /api/alerts", s.listAlerts)

	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	return logMiddleware(mux)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s -> %d", r.Method, r.URL.Path, rec.status)
	})
}

// statusRecorder wraps http.ResponseWriter to capture the status code that
// was actually written, so the request log can show it - the standard
// ResponseWriter interface has no way to read this back otherwise.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	if status >= 500 {
		log.Printf("server error: %s", message)
	}
	writeJSON(w, status, map[string]string{"error": message})
}
