package model

import "time"

type Target struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	URL             string    `json:"url"`
	IntervalSeconds int       `json:"intervalSeconds"`
	CreatedAt       time.Time `json:"createdAt"`
}

type UptimeCheck struct {
	ID             int64     `json:"id"`
	TargetID       int64     `json:"targetId"`
	CheckedAt      time.Time `json:"checkedAt"`
	Success        bool      `json:"success"`
	StatusCode     int       `json:"statusCode"`
	ResponseTimeMs int64     `json:"responseTimeMs"`
	Error          string    `json:"error,omitempty"`
}

type Host struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
}

// HostMetric is one collection snapshot pushed by an agent.
type HostMetric struct {
	ID            int64     `json:"id"`
	HostID        int64     `json:"hostId"`
	CollectedAt   time.Time `json:"collectedAt"`
	CPUPercent    float64   `json:"cpuPercent"`
	MemoryPercent float64   `json:"memoryPercent"`
	DiskPercent   float64   `json:"diskPercent"`
	LoadAvg1      float64   `json:"loadAvg1"`
}

// MetricPushRequest is the payload an agent POSTs to the collector.
type MetricPushRequest struct {
	HostName      string  `json:"hostName"`
	CPUPercent    float64 `json:"cpuPercent"`
	MemoryPercent float64 `json:"memoryPercent"`
	DiskPercent   float64 `json:"diskPercent"`
	LoadAvg1      float64 `json:"loadAvg1"`
}

type AlertTargetType string

const (
	AlertTargetUptime AlertTargetType = "target" // an uptime Target
	AlertTargetHost   AlertTargetType = "host"   // a monitored Host
)

type AlertMetric string

const (
	MetricUptimeDown AlertMetric = "uptime_down"
	MetricCPU        AlertMetric = "cpu_percent"
	MetricMemory     AlertMetric = "memory_percent"
	MetricDisk       AlertMetric = "disk_percent"
)

type AlertOperator string

const (
	OpGreaterThan AlertOperator = "gt"
	OpLessThan    AlertOperator = "lt"
)

type AlertRule struct {
	ID              int64           `json:"id"`
	Name            string          `json:"name"`
	TargetType      AlertTargetType `json:"targetType"`
	TargetRefID     int64           `json:"targetRefId"`
	Metric          AlertMetric     `json:"metric"`
	Operator        AlertOperator   `json:"operator"`
	Threshold       float64         `json:"threshold"`
	DurationSeconds int             `json:"durationSeconds"`
	WebhookURL      string          `json:"webhookUrl,omitempty"`
	Enabled         bool            `json:"enabled"`
	CreatedAt       time.Time       `json:"createdAt"`
}

type AlertStatus string

const (
	AlertFiring   AlertStatus = "firing"
	AlertResolved AlertStatus = "resolved"
)

type Alert struct {
	ID         int64       `json:"id"`
	RuleID     int64       `json:"ruleId"`
	RuleName   string      `json:"ruleName"`
	Status     AlertStatus `json:"status"`
	Message    string      `json:"message"`
	FiredAt    time.Time   `json:"firedAt"`
	ResolvedAt *time.Time  `json:"resolvedAt,omitempty"`
}
