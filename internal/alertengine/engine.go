package alertengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mrgeni717/sentinel/internal/model"
	"github.com/mrgeni717/sentinel/internal/store"
)

type Engine struct {
	store *store.Store
}

func New(s *store.Store) *Engine {
	return &Engine{store: s}
}

// EvaluateTarget checks every enabled rule attached to this uptime target
// against its most recent check result.
func (e *Engine) EvaluateTarget(targetID int64) {
	check, err := e.store.LatestUptimeCheck(targetID)
	if err != nil {
		return // no checks recorded yet
	}

	rules, err := e.store.ListAlertRules()
	if err != nil {
		log.Printf("alertengine: list rules: %v", err)
		return
	}

	for _, r := range rules {
		if !r.Enabled || r.TargetType != model.AlertTargetUptime || r.TargetRefID != targetID {
			continue
		}
		if r.Metric != model.MetricUptimeDown {
			continue
		}

		breached := !check.Success
		e.applyRuleResult(r, breached, fmt.Sprintf("Uptime check failed for target %d (status %d): %s", targetID, check.StatusCode, check.Error))
	}
}

// EvaluateHost checks every enabled rule attached to this host against its
// most recent pushed metric.
func (e *Engine) EvaluateHost(hostID int64) {
	metric, err := e.store.LatestHostMetric(hostID)
	if err != nil {
		return
	}

	rules, err := e.store.ListAlertRules()
	if err != nil {
		log.Printf("alertengine: list rules: %v", err)
		return
	}

	for _, r := range rules {
		if !r.Enabled || r.TargetType != model.AlertTargetHost || r.TargetRefID != hostID {
			continue
		}

		var value float64
		switch r.Metric {
		case model.MetricCPU:
			value = metric.CPUPercent
		case model.MetricMemory:
			value = metric.MemoryPercent
		case model.MetricDisk:
			value = metric.DiskPercent
		default:
			continue
		}

		breached := evaluateThreshold(r.Operator, value, r.Threshold)
		e.applyRuleResult(r, breached, fmt.Sprintf("%s is %.1f on host %d (threshold %.1f)", r.Metric, value, hostID, r.Threshold))
	}
}

func evaluateThreshold(op model.AlertOperator, value, threshold float64) bool {
	switch op {
	case model.OpGreaterThan:
		return value > threshold
	case model.OpLessThan:
		return value < threshold
	default:
		return false
	}
}

// applyRuleResult opens a new alert if the rule just started breaching and
// nothing is currently firing for it, or resolves the open alert if the
// rule is no longer breaching. DurationSeconds ("breach must hold for N
// seconds before firing") is intentionally not implemented in v1 - noted
// as a known simplification, since it would need tracking breach-start
// time per rule rather than evaluating point-in-time on each new reading.
func (e *Engine) applyRuleResult(r model.AlertRule, breached bool, message string) {
	open, err := e.store.OpenAlertForRule(r.ID)
	if err != nil {
		log.Printf("alertengine: open alert lookup for rule %d: %v", r.ID, err)
		return
	}

	switch {
	case breached && open == nil:
		if err := e.store.FireAlert(r.ID, message); err != nil {
			log.Printf("alertengine: fire alert for rule %d: %v", r.ID, err)
			return
		}
		e.sendWebhook(r, message, true)

	case !breached && open != nil:
		if err := e.store.ResolveAlert(open.ID); err != nil {
			log.Printf("alertengine: resolve alert %d: %v", open.ID, err)
			return
		}
		e.sendWebhook(r, fmt.Sprintf("RESOLVED: %s", message), false)
	}
	// breached && open != nil: already firing, nothing to do.
	// !breached && open == nil: healthy, nothing to do.
}

func (e *Engine) sendWebhook(r model.AlertRule, message string, firing bool) {
	if r.WebhookURL == "" {
		return
	}

	icon := "🔴"
	if !firing {
		icon = "✅"
	}

	// Slack's Incoming Webhooks require a JSON body with a "text" field -
	// other fields are accepted but ignored by Slack, kept here so a
	// non-Slack webhook receiver (e.g. a custom endpoint) still gets
	// structured data instead of only a formatted string.
	payload := map[string]any{
		"text":    fmt.Sprintf("%s *%s*\n%s", icon, r.Name, message),
		"rule":    r.Name,
		"firing":  firing,
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(r.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("alertengine: webhook for rule %d failed: %v", r.ID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("alertengine: webhook for rule %d returned status %d", r.ID, resp.StatusCode)
	}
}
