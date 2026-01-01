package monitor

import "time"

// HealthResponse matches Madao API JSON.
type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	LatencyMS int64  `json:"latency_ms"`
	Time      string `json:"time"`
	MessageJP string `json:"message_jp"`
	MessageEN string `json:"message_en"`
	RequestID string `json:"request_id"`
}

// OneResult is one check outcome.
type OneResult struct {
	URL       string
	HTTPCode  int
	Err       error
	Parsed    *HealthResponse
	WallTime  time.Duration
	CheckedAt time.Time
}

// Stats tracks rolling stats across all checks.
type Stats struct {
	TotalChecks int

	OKCount       int
	DegradedCount int
	SpiralCount   int

	NetFail   int
	HTTPFail  int
	ParseFail int

	LatenciesMS []float64
}

// TargetStats holds per-target rolling stats.
type TargetStats struct {
	Total int
	OK    int
	Deg   int
	Spi   int

	NetErr   int
	HTTPErr  int
	ParseErr int

	LastStatus  string
	LastCode    int
	LastLatency int64
	LastWall    time.Duration
	LastAt      time.Time

	LatenciesMS []float64
}
