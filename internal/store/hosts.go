package store

import "github.com/mrgeni717/sentinel/internal/model"

// UpsertHost finds a host by name or creates it, and updates last_seen_at.
// Called every time an agent pushes metrics.
func (s *Store) UpsertHost(name string) (*model.Host, error) {
	_, err := s.DB.Exec(
		`INSERT INTO hosts (name, last_seen_at) VALUES (?, CURRENT_TIMESTAMP)
		 ON CONFLICT(name) DO UPDATE SET last_seen_at = CURRENT_TIMESTAMP`,
		name,
	)
	if err != nil {
		return nil, err
	}
	return s.GetHostByName(name)
}

func (s *Store) GetHostByName(name string) (*model.Host, error) {
	row := s.DB.QueryRow(`SELECT id, name, created_at, last_seen_at FROM hosts WHERE name = ?`, name)
	var h model.Host
	if err := row.Scan(&h.ID, &h.Name, &h.CreatedAt, &h.LastSeenAt); err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *Store) ListHosts() ([]model.Host, error) {
	rows, err := s.DB.Query(`SELECT id, name, created_at, last_seen_at FROM hosts ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hosts := []model.Host{}
	for rows.Next() {
		var h model.Host
		if err := rows.Scan(&h.ID, &h.Name, &h.CreatedAt, &h.LastSeenAt); err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, rows.Err()
}

func (s *Store) RecordHostMetric(m model.HostMetric) error {
	_, err := s.DB.Exec(
		`INSERT INTO host_metrics (host_id, cpu_percent, memory_percent, disk_percent, load_avg_1) VALUES (?, ?, ?, ?, ?)`,
		m.HostID, m.CPUPercent, m.MemoryPercent, m.DiskPercent, m.LoadAvg1,
	)
	return err
}

func (s *Store) LatestHostMetric(hostID int64) (*model.HostMetric, error) {
	row := s.DB.QueryRow(
		`SELECT id, host_id, collected_at, cpu_percent, memory_percent, disk_percent, load_avg_1
		 FROM host_metrics WHERE host_id = ? ORDER BY collected_at DESC LIMIT 1`,
		hostID,
	)
	var m model.HostMetric
	if err := row.Scan(&m.ID, &m.HostID, &m.CollectedAt, &m.CPUPercent, &m.MemoryPercent, &m.DiskPercent, &m.LoadAvg1); err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *Store) HostMetricHistory(hostID int64, limit int) ([]model.HostMetric, error) {
	rows, err := s.DB.Query(
		`SELECT id, host_id, collected_at, cpu_percent, memory_percent, disk_percent, load_avg_1
		 FROM host_metrics WHERE host_id = ? ORDER BY collected_at DESC LIMIT ?`,
		hostID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := []model.HostMetric{}
	for rows.Next() {
		var m model.HostMetric
		if err := rows.Scan(&m.ID, &m.HostID, &m.CollectedAt, &m.CPUPercent, &m.MemoryPercent, &m.DiskPercent, &m.LoadAvg1); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}
