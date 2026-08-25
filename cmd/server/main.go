package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrgeni717/sentinel/internal/alertengine"
	"github.com/mrgeni717/sentinel/internal/api"
	"github.com/mrgeni717/sentinel/internal/checker"
	"github.com/mrgeni717/sentinel/internal/store"
)

func main() {
	dbPath := envOr("SENTINEL_DB_PATH", "sentinel.db")
	port := envOr("SENTINEL_PORT", "8090")
	staticDir := envOr("SENTINEL_STATIC_DIR", "web/static")
	ingestKey := os.Getenv("SENTINEL_INGEST_KEY") // empty = ingest endpoint is unauthenticated (fine for local dev)

	s, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer s.Close()

	engine := alertengine.New(s)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := checker.New(s, func(targetID int64) {
		engine.EvaluateTarget(targetID)
	})
	go c.Run(ctx)

	server := api.NewServer(s, engine, ingestKey)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: server.Routes(staticDir),
	}

	go func() {
		log.Printf("sentinel listening on :%s (db: %s)", port, dbPath)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
