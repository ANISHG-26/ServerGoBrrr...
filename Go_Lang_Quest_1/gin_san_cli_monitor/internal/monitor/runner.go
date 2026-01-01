package monitor

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RunTick checks all URLs once with concurrency limit and returns results.
func RunTick(urls []string, client *http.Client, timeout time.Duration, concurrency int) []OneResult {
	results := make([]OneResult, 0, len(urls))

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, u := range urls {
		wg.Add(1)
		sem <- struct{}{}

		go func(url string) {
			defer wg.Done()
			defer func() { <-sem }()

			r := checkOnce(client, url, timeout)
			mu.Lock()
			results = append(results, r)
			mu.Unlock()
		}(u)
	}

	wg.Wait()
	return results
}

// checkOnce performs a single GET using context timeout and decodes JSON.
func checkOnce(client *http.Client, url string, timeout time.Duration) OneResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return OneResult{URL: url, Err: err, CheckedAt: time.Now()}
	}

	start := time.Now()
	resp, err := client.Do(req)
	wall := time.Since(start)
	checkedAt := time.Now()

	if err != nil {
		return OneResult{URL: url, Err: err, WallTime: wall, CheckedAt: checkedAt}
	}
	defer resp.Body.Close()

	var parsed HealthResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&parsed)
	if decodeErr != nil {
		return OneResult{URL: url, HTTPCode: resp.StatusCode, Err: decodeErr, Parsed: nil, WallTime: wall, CheckedAt: checkedAt}
	}

	return OneResult{URL: url, HTTPCode: resp.StatusCode, Err: nil, Parsed: &parsed, WallTime: wall, CheckedAt: checkedAt}
}
