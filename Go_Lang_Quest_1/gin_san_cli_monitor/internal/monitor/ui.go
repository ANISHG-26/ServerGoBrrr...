package monitor

import (
	"fmt"
	"runtime"
	"time"
)

// ginsanCommentary maps statuses to messages.
func ginsanCommentary(status string) string {
	switch status {
	case "ok":
		return "You still breathing?"
	case "degraded":
		return "You look like crap, but youƒ?Tll live."
	case "madao_spiral":
		return "Alright, thatƒ?Ts enough."
	case "net_err":
		return "I canƒ?Tt even reach you."
	case "parse_err":
		return "Youƒ?Tre talking nonsense."
	default:
		return "ƒ?ÝIƒ?Tm watching."
	}
}

func stateVisual(status string, httpCode int) (label string, color string, icon string) {
	switch status {
	case "ok":
		return "ok", ANSIGreen, "dYY› "
	case "degraded":
		return "degraded", ANSIYellow, "dYY­ "
	case "madao_spiral":
		return "spiral", ANSIRed, "dYO? "
	case "net_err":
		return "net_err", ANSIRed, "ƒ?O "
	case "parse_err":
		return "parse_err", ANSIYellow, "ƒsÿ‹,? "
	default:
		if httpCode == 0 && status == "" {
			return "init", ANSIDim, "ƒ?Ý  "
		}
		return "unknown", ANSIGray, "ƒs¦ "
	}
}

// RenderOverallSummary prints a compact overall summary.
func RenderOverallSummary(s *Stats) {
	total := s.TotalChecks
	if total == 0 {
		fmt.Printf("%sNo checks yet.%s\n", ANSIDim, ANSIReset)
		return
	}

	okPct := 100 * float64(s.OKCount) / float64(total)
	degPct := 100 * float64(s.DegradedCount) / float64(total)
	spiPct := 100 * float64(s.SpiralCount) / float64(total)

	avg, p50, p95, max := LatencyStats(s.LatenciesMS)

	fmt.Printf("%sOverall%s  checks=%d  ", ANSIBold, ANSIReset, total)
	fmt.Printf("%sok%s=%d(%.1f%%)  ", ANSIGreen, ANSIReset, s.OKCount, okPct)
	fmt.Printf("%sdegraded%s=%d(%.1f%%)  ", ANSIYellow, ANSIReset, s.DegradedCount, degPct)
	fmt.Printf("%sspirals%s=%d(%.1f%%)\n", ANSIRed, ANSIReset, s.SpiralCount, spiPct)

	fmt.Printf("          net_err=%d  http_fail=%d  parse_fail=%d\n", s.NetFail, s.HTTPFail, s.ParseFail)

	if len(s.LatenciesMS) > 0 {
		fmt.Printf("          latency_ms: avg=%.0f  p50=%.0f  p95=%.0f  max=%.0f\n", avg, p50, p95, max)
	} else {
		fmt.Printf("          latency_ms: (no data)\n")
	}
}

// RenderDashboard draws the main dashboard.
func RenderDashboard(tick int, started time.Time, urls []string, overall *Stats, perTarget map[string]*TargetStats, results []OneResult, verbose bool) {
	uptime := time.Since(started).Round(time.Second)

	fmt.Printf("%s%sGIN-SAN MONITOR%s  %s| tick=%d | uptime=%s | go=%s | os=%s%s\n",
		ANSIBold, ANSICyan, ANSIReset,
		ANSIGray, tick, uptime, runtime.Version(), runtime.GOOS, ANSIReset)

	RenderOverallSummary(overall)

	fmt.Printf("\n%sTargets%s\n", ANSIBold, ANSIReset)

	// Header
	fmt.Printf("%s\n", ANSIGray+
		PadRight("STATE", 10)+"  "+
		PadRight("HTTP", 6)+"  "+
		PadRight("LAT(ms)", 8)+"  "+
		PadRight("WALL", 10)+"  "+
		PadRight("OK%", 6)+"  "+
		PadRight("URL", 45)+
		ANSIReset)

	for _, u := range urls {
		ts := perTarget[u]
		if ts == nil {
			continue
		}

		label, color, icon := stateVisual(ts.LastStatus, ts.LastCode)

		httpStr := "-"
		if ts.LastCode != 0 {
			httpStr = fmt.Sprintf("%d", ts.LastCode)
		}

		okPct := 0.0
		if ts.Total > 0 {
			okPct = 100.0 * float64(ts.OK) / float64(ts.Total)
		}

		latStr := "-"
		if ts.LastLatency > 0 {
			latStr = fmt.Sprintf("%d", ts.LastLatency)
		}

		wallStr := "-"
		if !ts.LastAt.IsZero() {
			wallStr = ts.LastWall.Round(time.Millisecond).String()
		}

		urlShort := u
		if len(urlShort) > 45 {
			urlShort = urlShort[:42] + "..."
		}

		stateCol := PadRight(icon+label, 10)
		fmt.Printf("%s%s%s  %s  %s  %s  %5.1f  %s%s\n",
			color, stateCol, ANSIReset,
			PadRight(httpStr, 6),
			PadRight(latStr, 8),
			PadRight(wallStr, 10),
			okPct,
			PadRight(urlShort, 45),
			ANSIReset)

		fmt.Printf("   %s%sGin-san:%s %s%s\n",
			ANSIDim, ANSICyan, ANSIReset, ginsanCommentary(ts.LastStatus), ANSIReset)
	}

	if verbose {
		fmt.Printf("\n%sDetails (last tick)%s\n", ANSIBold, ANSIReset)
		for _, r := range results {
			if r.Parsed != nil {
				fmt.Printf("%s%s%s\n", ANSIBlue, r.URL, ANSIReset)
				fmt.Printf("  status=%s http=%d latency=%dms wall=%s\n", r.Parsed.Status, r.HTTPCode, r.Parsed.LatencyMS, r.WallTime.Round(time.Millisecond))
				fmt.Printf("  jp=%s\n", r.Parsed.MessageJP)
				fmt.Printf("  en=%s\n", r.Parsed.MessageEN)
				fmt.Printf("  id=%s time=%s\n", r.Parsed.RequestID, r.Parsed.Time)
			} else if r.Err != nil {
				fmt.Printf("%s%s%s\n", ANSIBlue, r.URL, ANSIReset)
				fmt.Printf("  error=%v\n", r.Err)
			}
		}
	}

	fmt.Printf("\n%sHint:%s If you see lots of %sdYO?madao_spiral%s, try raising --timeout or reducing the service --fail-pct.\n",
		ANSIBold, ANSIReset, ANSIRed, ANSIReset)
}
