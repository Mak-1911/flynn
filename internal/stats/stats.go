// Package stats provides system statistics tracking for Flynn.
package stats

import (
	"runtime"
	"time"
)

// Collector collects and tracks system statistics.
type Collector struct {
	startTime     time.Time
	requestCount  int64
	tokenCount    int64
	errorCount    int64
	totalDuration int64 // nanoseconds
}

// NewCollector creates a new stats collector.
func NewCollector() *Collector {
	return &Collector{
		startTime: time.Now(),
	}
}

// Stats represents system statistics at a point in time.
type Stats struct {
	// System resources
	MemoryStats MemoryStats `json:"memory"`
	Goroutines  int         `json:"goroutines"`
	Uptime      string      `json:"uptime"`

	// Agent metrics
	RequestCount int64   `json:"request_count"`
	TokenCount   int64   `json:"token_count"`
	ErrorCount   int64   `json:"error_count"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`

	// Database info
	DBSize    int64  `json:"db_size_bytes"`
	DBSizeMB  float64 `json:"db_size_mb"`
	DBPath    string `json:"db_path,omitempty"`
}

// MemoryStats represents memory usage statistics.
type MemoryStats struct {
	// Heap memory
	HeapAlloc     int64   `json:"heap_alloc_bytes"`
	HeapAllocMB   float64 `json:"heap_alloc_mb"`
	HeapSys       int64   `json:"heap_sys_bytes"`
	HeapSysMB     float64 `json:"heap_sys_mb"`
	HeapInuse     int64   `json:"heap_inuse_bytes"`
	HeapInuseMB   float64 `json:"heap_inuse_mb"`
	HeapObjects   uint64  `json:"heap_objects"`

	// Stack memory
	StackInuse   int64   `json:"stack_inuse_bytes"`
	StackInuseMB float64 `json:"stack_inuse_mb"`

	// GC stats
	NumGC       uint32        `json:"num_gc"`
	LastGC      string        `json:"last_gc,omitempty"`
	GCPauseTotal time.Duration `json:"gc_pause_total"`
}

// Collect returns current system statistics.
func (c *Collector) Collect(dbSize int64, dbPath string) *Stats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(c.startTime)
	avgLatency := float64(0)
	if c.requestCount > 0 {
		avgLatency = float64(c.totalDuration) / float64(c.requestCount) / 1e6 // nanos to millis
	}

	return &Stats{
		MemoryStats: MemoryStats{
			HeapAlloc:     int64(m.HeapAlloc),
			HeapAllocMB:   bytesToMB(int64(m.HeapAlloc)),
			HeapSys:       int64(m.HeapSys),
			HeapSysMB:     bytesToMB(int64(m.HeapSys)),
			HeapInuse:     int64(m.HeapInuse),
			HeapInuseMB:   bytesToMB(int64(m.HeapInuse)),
			HeapObjects:   m.HeapObjects,
			StackInuse:    int64(m.StackInuse),
			StackInuseMB:  bytesToMB(int64(m.StackInuse)),
			NumGC:         m.NumGC,
			LastGC:        time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
			GCPauseTotal:  time.Duration(m.PauseTotalNs),
		},
		Goroutines:   runtime.NumGoroutine(),
		Uptime:       uptime.String(),
		RequestCount: c.requestCount,
		TokenCount:   c.tokenCount,
		ErrorCount:   c.errorCount,
		AvgLatencyMs: avgLatency,
		DBSize:       dbSize,
		DBSizeMB:     bytesToMB(dbSize),
		DBPath:       dbPath,
	}
}

// RecordRequest records a completed request.
func (c *Collector) RecordRequest(tokens int, duration time.Duration) {
	c.requestCount++
	c.tokenCount += int64(tokens)
	c.totalDuration += duration.Nanoseconds()
}

// RecordError records an error.
func (c *Collector) RecordError() {
	c.errorCount++
}

// StartTime returns when the collector started.
func (c *Collector) StartTime() time.Time {
	return c.startTime
}

// GetMetrics returns current metrics.
func (c *Collector) GetMetrics() (requests, tokens, errors int64, totalDuration time.Duration) {
	return c.requestCount, c.tokenCount, c.errorCount, time.Duration(c.totalDuration)
}

// bytesToMB converts bytes to megabytes.
func bytesToMB(b int64) float64 {
	return float64(b) / 1024 / 1024
}
