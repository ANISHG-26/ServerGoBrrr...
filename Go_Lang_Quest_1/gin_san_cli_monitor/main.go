package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"gin_san_cli_monitor/internal/monitor"
)

func main() {
	urlsFlag := flag.String("urls", "http://localhost:8080/health", "Comma-separated list of health URLs")
	interval := flag.Duration("interval", 2*time.Second, "Interval between checks (e.g. 2s, 500ms)")
	timeout := flag.Duration("timeout", 1200*time.Millisecond, "Per-request timeout (e.g. 800ms, 2s)")
	concurrency := flag.Int("c", 8, "Max concurrent checks per tick")
	iterations := flag.Int("n", 0, "Number of ticks (0 = run forever)")
	verbose := flag.Bool("v", false, "Verbose: show Madao messages per target (prints last response details)")
	noClear := flag.Bool("no-clear", false, "Do not clear screen (better for logs/history)")
	flag.Parse()

	urls := monitor.ParseURLs(*urlsFlag)
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "No URLs provided. Use --urls=http://localhost:8080/health")
		os.Exit(2)
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
	}

	overall := &monitor.Stats{}
	perTarget := make(map[string]*monitor.TargetStats, len(urls))
	for _, u := range urls {
		perTarget[u] = &monitor.TargetStats{}
	}

	started := time.Now()

	fmt.Printf("%s%sgin-san%s  🛰️  keeping an eye on Madao%s\n", monitor.ANSIBold, monitor.ANSICyan, monitor.ANSIReset, monitor.ANSIReset)
	fmt.Printf("%sTargets:%s %s\n", monitor.ANSIBold, monitor.ANSIReset, join(urls, ", "))
	fmt.Printf("%sInterval:%s %s  %sTimeout:%s %s  %sConcurrency:%s %d  %sIterations:%s %d\n\n",
		monitor.ANSIBold, monitor.ANSIReset, interval,
		monitor.ANSIBold, monitor.ANSIReset, timeout,
		monitor.ANSIBold, monitor.ANSIReset, *concurrency,
		monitor.ANSIBold, monitor.ANSIReset, *iterations)

	tick := 0
	for {
		tick++
		if *iterations > 0 && tick > *iterations {
			break
		}

		tickStart := time.Now()
		results := monitor.RunTick(urls, client, *timeout, *concurrency)

		monitor.UpdateStats(overall, perTarget, results)

		if !*noClear {
			monitor.ClearScreen()
		}
		monitor.RenderDashboard(tick, started, urls, overall, perTarget, results, *verbose)

		elapsed := time.Since(tickStart)
		sleepFor := *interval - elapsed
		if sleepFor > 0 {
			time.Sleep(sleepFor)
		}
	}

	fmt.Printf("\n%sFinal summary:%s\n", monitor.ANSIBold, monitor.ANSIReset)
	monitor.RenderOverallSummary(overall)
}

// small helper to avoid importing strings just for join in banner
func join(vals []string, sep string) string {
	if len(vals) == 0 {
		return ""
	}
	out := vals[0]
	for i := 1; i < len(vals); i++ {
		out += sep + vals[i]
	}
	return out
}
