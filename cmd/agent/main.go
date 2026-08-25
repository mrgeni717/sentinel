// Command agent runs on a monitored host, collects CPU/memory/disk stats
// periodically, and pushes them to a sentinel server. It's a single
// static binary (no runtime dependencies) - cross-compile for any target
// with `GOOS=linux GOARCH=amd64 go build ./cmd/agent` and copy it over.
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/mrgeni717/sentinel/internal/model"
)

func main() {
	serverURL := envOr("SENTINEL_SERVER_URL", "http://localhost:8090")
	hostName := envOr("SENTINEL_HOST_NAME", mustHostname())
	ingestKey := os.Getenv("SENTINEL_INGEST_KEY")
	intervalSeconds := envOrInt("SENTINEL_PUSH_INTERVAL_SECONDS", 30)

	log.Printf("sentinel agent starting: host=%s server=%s interval=%ds", hostName, serverURL, intervalSeconds)

	client := &http.Client{Timeout: 10 * time.Second}
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	pushOnce(client, serverURL, hostName, ingestKey)
	for range ticker.C {
		pushOnce(client, serverURL, hostName, ingestKey)
	}
}

func pushOnce(client *http.Client, serverURL, hostName, ingestKey string) {
	metrics, err := collect()
	if err != nil {
		log.Printf("collect metrics: %v", err)
		return
	}
	metrics.HostName = hostName

	body, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("marshal metrics: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, serverURL+"/api/ingest/metrics", bytes.NewReader(body))
	if err != nil {
		log.Printf("build request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if ingestKey != "" {
		req.Header.Set("X-Ingest-Key", ingestKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("push metrics: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		log.Printf("push metrics: server returned %d", resp.StatusCode)
		return
	}
	log.Printf("pushed metrics: cpu=%.1f%% mem=%.1f%% disk=%.1f%%", metrics.CPUPercent, metrics.MemoryPercent, metrics.DiskPercent)
}

func collect() (model.MetricPushRequest, error) {
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return model.MetricPushRequest{}, err
	}
	var cpuPct float64
	if len(cpuPercents) > 0 {
		cpuPct = cpuPercents[0]
	}

	vmem, err := mem.VirtualMemory()
	if err != nil {
		return model.MetricPushRequest{}, err
	}

	diskUsage, err := disk.Usage(rootPath())
	if err != nil {
		return model.MetricPushRequest{}, err
	}

	var loadAvg1 float64
	if avg, err := load.Avg(); err == nil {
		loadAvg1 = avg.Load1
	} // load average isn't available on Windows; ignore the error there

	return model.MetricPushRequest{
		CPUPercent:    cpuPct,
		MemoryPercent: vmem.UsedPercent,
		DiskPercent:   diskUsage.UsedPercent,
		LoadAvg1:      loadAvg1,
	}, nil
}

func rootPath() string {
	if os.PathSeparator == '\\' {
		return `C:\`
	}
	return "/"
}

func mustHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown-host"
	}
	return name
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
