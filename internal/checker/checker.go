package checker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/mrgeni717/sentinel/internal/model"
	"github.com/mrgeni717/sentinel/internal/store"
)

type Checker struct {
	store      *store.Store
	httpClient *http.Client
	onResult   func(targetID int64) // called after each check, e.g. to evaluate alert rules
}

func New(s *store.Store, onResult func(targetID int64)) *Checker {
	return &Checker{
		store:      s,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		onResult:   onResult,
	}
}

// Run starts one goroutine per target's own interval and blocks until ctx
// is cancelled. Each target ticks independently rather than sharing a
// single global tick, so a slow-interval target doesn't get checked more
// often than configured just because a fast one shares its loop.
func (c *Checker) Run(ctx context.Context) {
	knownTargets := map[int64]context.CancelFunc{}

	// Poll the target list periodically for newly added/removed targets
	// rather than requiring a restart - simple and sufficient at this scale.
	refresh := time.NewTicker(15 * time.Second)
	defer refresh.Stop()

	c.syncTargets(ctx, knownTargets)

	for {
		select {
		case <-ctx.Done():
			for _, cancel := range knownTargets {
				cancel()
			}
			return
		case <-refresh.C:
			c.syncTargets(ctx, knownTargets)
		}
	}
}

func (c *Checker) syncTargets(ctx context.Context, knownTargets map[int64]context.CancelFunc) {
	targets, err := c.store.ListTargets()
	if err != nil {
		log.Printf("checker: list targets: %v", err)
		return
	}

	seen := map[int64]bool{}
	for _, t := range targets {
		seen[t.ID] = true
		if _, exists := knownTargets[t.ID]; !exists {
			targetCtx, cancel := context.WithCancel(ctx)
			knownTargets[t.ID] = cancel
			go c.watchTarget(targetCtx, t)
		}
	}

	for id, cancel := range knownTargets {
		if !seen[id] {
			cancel()
			delete(knownTargets, id)
		}
	}
}

func (c *Checker) watchTarget(ctx context.Context, t model.Target) {
	interval := time.Duration(t.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	c.checkOnce(t)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkOnce(t)
		}
	}
}

func (c *Checker) checkOnce(t model.Target) {
	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, t.URL, nil)

	check := model.UptimeCheck{TargetID: t.ID}

	if err != nil {
		check.Success = false
		check.Error = err.Error()
	} else {
		resp, err := c.httpClient.Do(req)
		check.ResponseTimeMs = time.Since(start).Milliseconds()
		if err != nil {
			check.Success = false
			check.Error = err.Error()
		} else {
			defer resp.Body.Close()
			check.StatusCode = resp.StatusCode
			check.Success = resp.StatusCode >= 200 && resp.StatusCode < 400
			if !check.Success {
				check.Error = "unexpected status code"
			}
		}
	}

	if err := c.store.RecordUptimeCheck(check); err != nil {
		log.Printf("checker: record check for target %d: %v", t.ID, err)
		return
	}

	if c.onResult != nil {
		c.onResult(t.ID)
	}
}
