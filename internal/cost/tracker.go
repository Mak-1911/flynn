// Package cost tracks token usage and costs for transparency.
package cost

import "time"

// Tracker monitors AI usage and calculates costs.
type Tracker struct {
	localFree  bool // Local models are free
	hourlyRate map[string]float64 // Cost per 1M tokens by model
	daily      *DailyStats
	monthly    *MonthlyStats
}

// DailyStats tracks cost for a single day.
type DailyStats struct {
	Date        string
	LocalTokens int
	CloudTokens int
	CloudCost   float64
	Requests    int
}

// MonthlyStats tracks cost for a month.
type MonthlyStats struct {
	Month       string
	LocalTokens int
	CloudTokens int
	CloudCost   float64
	Requests    int
	LocalRate   float64 // Percentage handled locally
}

// NewTracker creates a new cost tracker.
func NewTracker() *Tracker {
	return &Tracker{
		localFree:  true,
		hourlyRate: make(map[string]float64),
		daily:      &DailyStats{},
		monthly:    &MonthlyStats{},
	}
}

// Record records a model inference request.
func (t *Tracker) Record(model string, isLocal bool, tokens int, cost float64) {
	if isLocal {
		t.daily.LocalTokens += tokens
		t.monthly.LocalTokens += tokens
	} else {
		t.daily.CloudTokens += tokens
		t.monthly.CloudTokens += tokens
		t.daily.CloudCost += cost
		t.monthly.CloudCost += cost
	}
	t.daily.Requests++
	t.monthly.Requests++
}

// Savings returns the savings compared to using cloud for everything.
// Assumes $0.50 per 1M tokens for cloud baseline.
func (t *Tracker) Savings() float64 {
	const cloudCostPerMillion = 0.50
	totalTokens := t.daily.LocalTokens + t.daily.CloudTokens
	if totalTokens == 0 {
		return 0
	}
	allCloudCost := float64(totalTokens) / 1_000_000 * cloudCostPerMillion
	return allCloudCost - t.daily.CloudCost
}

// LocalRate returns the percentage of requests handled locally.
func (t *Tracker) LocalRate() float64 {
	total := t.daily.LocalTokens + t.daily.CloudTokens
	if total == 0 {
		return 0
	}
	return float64(t.daily.LocalTokens) / float64(total) * 100
}

// GetDailyStats returns the current daily statistics.
func (t *Tracker) GetDailyStats() *DailyStats {
	return t.daily
}

// GetMonthlyStats returns the current monthly statistics.
func (t *Tracker) GetMonthlyStats() *MonthlyStats {
	return t.monthly
}

// ResetDaily resets daily stats (call at midnight).
func (t *Tracker) ResetDaily() {
	t.daily = &DailyStats{Date: time.Now().Format("2006-01-02")}
}

// ResetMonthly resets monthly stats (call on 1st of month).
func (t *Tracker) ResetMonthly() {
	t.monthly = &MonthlyStats{Month: time.Now().Format("2006-01")}
}
