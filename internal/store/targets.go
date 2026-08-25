package store

import (
	"github.com/mrgeni717/sentinel/internal/model"
)

func (s *Store) CreateTarget(name, url string, intervalSeconds int) (*model.Target, error) {
	res, err := s.DB.Exec(
		`INSERT INTO targets (name, url, interval_seconds) VALUES (?, ?, ?)`,
		name, url, intervalSeconds,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetTarget(id)
}

func (s *Store) GetTarget(id int64) (*model.Target, error) {
	row := s.DB.QueryRow(`SELECT id, name, url, interval_seconds, created_at FROM targets WHERE id = ?`, id)
	var t model.Target
	if err := row.Scan(&t.ID, &t.Name, &t.URL, &t.IntervalSeconds, &t.CreatedAt); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) ListTargets() ([]model.Target, error) {
	rows, err := s.DB.Query(`SELECT id, name, url, interval_seconds, created_at FROM targets ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := []model.Target{}
	for rows.Next() {
		var t model.Target
		if err := rows.Scan(&t.ID, &t.Name, &t.URL, &t.IntervalSeconds, &t.CreatedAt); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

func (s *Store) DeleteTarget(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM targets WHERE id = ?`, id)
	return err
}

func (s *Store) RecordUptimeCheck(c model.UptimeCheck) error {
	_, err := s.DB.Exec(
		`INSERT INTO uptime_checks (target_id, success, status_code, response_time_ms, error) VALUES (?, ?, ?, ?, ?)`,
		c.TargetID, c.Success, c.StatusCode, c.ResponseTimeMs, nullIfEmpty(c.Error),
	)
	return err
}

func (s *Store) LatestUptimeCheck(targetID int64) (*model.UptimeCheck, error) {
	row := s.DB.QueryRow(
		`SELECT id, target_id, checked_at, success, status_code, response_time_ms, COALESCE(error, '')
		 FROM uptime_checks WHERE target_id = ? ORDER BY checked_at DESC LIMIT 1`,
		targetID,
	)
	var c model.UptimeCheck
	if err := row.Scan(&c.ID, &c.TargetID, &c.CheckedAt, &c.Success, &c.StatusCode, &c.ResponseTimeMs, &c.Error); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UptimeHistory(targetID int64, limit int) ([]model.UptimeCheck, error) {
	rows, err := s.DB.Query(
		`SELECT id, target_id, checked_at, success, status_code, response_time_ms, COALESCE(error, '')
		 FROM uptime_checks WHERE target_id = ? ORDER BY checked_at DESC LIMIT ?`,
		targetID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := []model.UptimeCheck{}
	for rows.Next() {
		var c model.UptimeCheck
		if err := rows.Scan(&c.ID, &c.TargetID, &c.CheckedAt, &c.Success, &c.StatusCode, &c.ResponseTimeMs, &c.Error); err != nil {
			return nil, err
		}
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
