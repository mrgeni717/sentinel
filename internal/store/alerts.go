package store

import (
	"database/sql"

	"github.com/mrgeni717/sentinel/internal/model"
)

func (s *Store) CreateAlertRule(r model.AlertRule) (*model.AlertRule, error) {
	res, err := s.DB.Exec(
		`INSERT INTO alert_rules (name, target_type, target_ref_id, metric, operator, threshold, duration_seconds, webhook_url, enabled)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.Name, r.TargetType, r.TargetRefID, r.Metric, r.Operator, r.Threshold, r.DurationSeconds, nullIfEmpty(r.WebhookURL), r.Enabled,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetAlertRule(id)
}

func (s *Store) GetAlertRule(id int64) (*model.AlertRule, error) {
	row := s.DB.QueryRow(
		`SELECT id, name, target_type, target_ref_id, metric, operator, threshold, duration_seconds, COALESCE(webhook_url, ''), enabled, created_at
		 FROM alert_rules WHERE id = ?`, id,
	)
	return scanAlertRule(row)
}

func (s *Store) ListAlertRules() ([]model.AlertRule, error) {
	rows, err := s.DB.Query(
		`SELECT id, name, target_type, target_ref_id, metric, operator, threshold, duration_seconds, COALESCE(webhook_url, ''), enabled, created_at
		 FROM alert_rules ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []model.AlertRule{}
	for rows.Next() {
		r, err := scanAlertRuleRows(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *r)
	}
	return rules, rows.Err()
}

func (s *Store) DeleteAlertRule(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM alert_rules WHERE id = ?`, id)
	return err
}

// OpenAlertForRule returns the currently-firing alert for a rule, if any.
func (s *Store) OpenAlertForRule(ruleID int64) (*model.Alert, error) {
	row := s.DB.QueryRow(
		`SELECT a.id, a.rule_id, r.name, a.status, a.message, a.fired_at, a.resolved_at
		 FROM alerts a JOIN alert_rules r ON r.id = a.rule_id
		 WHERE a.rule_id = ? AND a.status = 'firing' ORDER BY a.fired_at DESC LIMIT 1`,
		ruleID,
	)
	a, err := scanAlert(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return a, err
}

func (s *Store) FireAlert(ruleID int64, message string) error {
	_, err := s.DB.Exec(
		`INSERT INTO alerts (rule_id, status, message) VALUES (?, 'firing', ?)`,
		ruleID, message,
	)
	return err
}

func (s *Store) ResolveAlert(alertID int64) error {
	_, err := s.DB.Exec(
		`UPDATE alerts SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP WHERE id = ?`,
		alertID,
	)
	return err
}

func (s *Store) ListAlerts(limit int) ([]model.Alert, error) {
	rows, err := s.DB.Query(
		`SELECT a.id, a.rule_id, r.name, a.status, a.message, a.fired_at, a.resolved_at
		 FROM alerts a JOIN alert_rules r ON r.id = a.rule_id
		 ORDER BY a.fired_at DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []model.Alert{}
	for rows.Next() {
		a, err := scanAlertRows(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, *a)
	}
	return alerts, rows.Err()
}

// --- scan helpers (database/sql.Row and .Rows share a Scan signature via this interface) ---

type scanner interface {
	Scan(dest ...any) error
}

func scanAlertRule(row scanner) (*model.AlertRule, error) {
	var r model.AlertRule
	err := row.Scan(&r.ID, &r.Name, &r.TargetType, &r.TargetRefID, &r.Metric, &r.Operator, &r.Threshold, &r.DurationSeconds, &r.WebhookURL, &r.Enabled, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}
func scanAlertRuleRows(rows *sql.Rows) (*model.AlertRule, error) { return scanAlertRule(rows) }

func scanAlert(row scanner) (*model.Alert, error) {
	var a model.Alert
	var resolvedAt sql.NullTime
	err := row.Scan(&a.ID, &a.RuleID, &a.RuleName, &a.Status, &a.Message, &a.FiredAt, &resolvedAt)
	if err != nil {
		return nil, err
	}
	if resolvedAt.Valid {
		a.ResolvedAt = &resolvedAt.Time
	}
	return &a, nil
}
func scanAlertRows(rows *sql.Rows) (*model.Alert, error) { return scanAlert(rows) }
