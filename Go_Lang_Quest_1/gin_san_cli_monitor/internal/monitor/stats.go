package monitor

import "math"

// UpdateStats updates overall and per-target counters based on results.
func UpdateStats(overall *Stats, perTarget map[string]*TargetStats, results []OneResult) {
	for _, r := range results {
		overall.TotalChecks++
		t := perTarget[r.URL]
		if t == nil {
			t = &TargetStats{}
			perTarget[r.URL] = t
		}
		t.Total++
		t.LastAt = r.CheckedAt
		t.LastWall = r.WallTime

		// Network error (timeouts, refused, DNS) => no HTTP code
		if r.Err != nil && r.Parsed == nil && r.HTTPCode == 0 {
			overall.NetFail++
			t.NetErr++
			t.LastStatus = "net_err"
			t.LastCode = 0
			t.LastLatency = 0
			continue
		}

		// Parse error (HTTP came back, but JSON failed)
		if r.Err != nil && r.Parsed == nil && r.HTTPCode != 0 {
			overall.ParseFail++
			overall.HTTPFail++
			t.ParseErr++
			t.HTTPErr++
			t.LastStatus = "parse_err"
			t.LastCode = r.HTTPCode
			t.LastLatency = 0
			continue
		}

		// Parsed OK
		hr := r.Parsed
		t.LastStatus = hr.Status
		t.LastCode = r.HTTPCode
		t.LastLatency = hr.LatencyMS

		overall.LatenciesMS = append(overall.LatenciesMS, float64(hr.LatencyMS))
		t.LatenciesMS = append(t.LatenciesMS, float64(hr.LatencyMS))

		switch hr.Status {
		case "ok":
			overall.OKCount++
			t.OK++
		case "degraded":
			overall.DegradedCount++
			t.Deg++
		case "madao_spiral":
			overall.SpiralCount++
			t.Spi++
		default:
			overall.DegradedCount++
			t.Deg++
		}

		if r.HTTPCode >= 400 {
			overall.HTTPFail++
			t.HTTPErr++
		}
	}
}

// LatencyStats computes avg, p50, p95 and max for a slice of latencies.
func LatencyStats(values []float64) (avg, p50, p95, max float64) {
	if len(values) == 0 {
		return math.NaN(), math.NaN(), math.NaN(), math.NaN()
	}

	cp := make([]float64, len(values))
	copy(cp, values)

	// Simple O(n^2) sort (fine for small learning data).
	for i := 0; i < len(cp); i++ {
		for j := i + 1; j < len(cp); j++ {
			if cp[j] < cp[i] {
				cp[i], cp[j] = cp[j], cp[i]
			}
		}
	}

	sum := 0.0
	for _, v := range cp {
		sum += v
	}
	avg = sum / float64(len(cp))
	max = cp[len(cp)-1]
	p50 = percentileSorted(cp, 50)
	p95 = percentileSorted(cp, 95)
	return avg, p50, p95, max
}

func percentileSorted(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	rank := int(math.Ceil(float64(p) / 100.0 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}
