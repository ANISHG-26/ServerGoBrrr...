# Madao API — SRE Chaos Playground

![Madao spiral](../../assets/madao-spiral.png)

Tiny HTTP service to practice SRE basics: golden signals, chaos drills, and graceful shutdown.

## Run

- Healthy default: `go run main.go -name "Madao API" -port 8080`
- Chaos on: `go run main.go -chaos -fail-pct 20 -slow-pct 30 -max-delay-ms 1200 -slow-threshold-ms 800 -persistent=true`
- Flags: `-fail-pct` failure rate, `-slow-pct` slow rate, `-max-delay-ms` added latency, `-slow-threshold-ms` mark degraded, `-persistent` keep diagnostic JSON on failure.

## Endpoints

- `GET /` welcome text.
- `GET /health` JSON with status (`ok`, `degraded`, `madao_spiral`), latency, time, bilingual message, and request id. Returns `503` when in `madao_spiral`, `200` otherwise.

## Sample /health

```json
{
  "status": "degraded",
  "service": "Madao API",
  "version": "0.1.0",
  "latency_ms": 910,
  "time": "2024-12-30T12:00:00Z",
  "message_jp": "心が折れそうだ…でも動く (マダオ)",
  "message_en": "Slow and struggling but still responding.",
  "request_id": "madao-123456"
}
```

## Chaos patterns (CLI + expected behavior)

- Latency heavy, no failures:  
  `go run main.go -chaos -slow-pct 80 -max-delay-ms 1800 -fail-pct 0 -slow-threshold-ms 700`  
  Expect many `degraded` responses; good for testing timeouts, retries, and backoff.
- Spiky errors, low latency:  
  `go run main.go -chaos -fail-pct 40 -slow-pct 5 -max-delay-ms 200 -slow-threshold-ms 900`  
  Expect frequent `503` (`madao_spiral`) without much delay; good for circuit breakers.
- Mixed brownout:  
  `go run main.go -chaos -fail-pct 15 -slow-pct 50 -max-delay-ms 1500 -slow-threshold-ms 850 -persistent=true`  
  Mix of `degraded` and occasional failures; practice distinguishing soft vs. hard incidents.
- Silent outages (poor diagnostics):  
  `go run main.go -chaos -fail-pct 50 -slow-pct 20 -persistent=false`  
  Failures return minimal JSON; observe how limited telemetry slows incident response.
- Shutdown drill:  
  `go run main.go -chaos -fail-pct 10 -slow-pct 20 -shutdown-timeout-sec 5` then hit `Ctrl+C`  
  Watch logs for draining and ensure clients handle in-flight cancellations.

## SRE practice ideas

- Treat `degraded` as an early warning; alert on sustained latency before full failure.
- Compare runs with `-persistent` on/off to see how diagnostics affect MTTR.
- Use `request_id` from responses to trace log lines and correlate chaos-induced symptoms.
- Plot latency and error rate to rehearse SLO/SLA thinking and error budgets.

## Notes

- Server listens on `:port` and drains for `-shutdown-timeout-sec` before exiting.
- Logs print chaos configuration at startup and per-request status/latency for quick observability.
