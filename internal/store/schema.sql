CREATE TABLE IF NOT EXISTS targets (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    name             TEXT NOT NULL,
    url              TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS uptime_checks (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    target_id        INTEGER NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    checked_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    success          INTEGER NOT NULL,
    status_code      INTEGER NOT NULL DEFAULT 0,
    response_time_ms INTEGER NOT NULL DEFAULT 0,
    error            TEXT
);
CREATE INDEX IF NOT EXISTS idx_uptime_checks_target ON uptime_checks(target_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS hosts (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT NOT NULL UNIQUE,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at  DATETIME
);

CREATE TABLE IF NOT EXISTS host_metrics (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    host_id         INTEGER NOT NULL REFERENCES hosts(id) ON DELETE CASCADE,
    collected_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cpu_percent     REAL NOT NULL,
    memory_percent  REAL NOT NULL,
    disk_percent    REAL NOT NULL,
    load_avg_1      REAL NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_host_metrics_host ON host_metrics(host_id, collected_at DESC);

CREATE TABLE IF NOT EXISTS alert_rules (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    name             TEXT NOT NULL,
    target_type      TEXT NOT NULL,
    target_ref_id    INTEGER NOT NULL,
    metric           TEXT NOT NULL,
    operator         TEXT NOT NULL,
    threshold        REAL NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    webhook_url      TEXT,
    enabled          INTEGER NOT NULL DEFAULT 1,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alerts (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id      INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'firing',
    message      TEXT NOT NULL,
    fired_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at  DATETIME
);
CREATE INDEX IF NOT EXISTS idx_alerts_rule ON alerts(rule_id, fired_at DESC);
