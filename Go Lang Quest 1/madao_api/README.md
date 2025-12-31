# Madao API — SRE Chaos Playground

![Madao spiral](data:image/svg+xml;utf8,%3Csvg%20xmlns%3D%27http%3A//www.w3.org/2000/svg%27%20viewBox%3D%270%200%20420%20220%27%3E%0A%20%20%3Cdefs%3E%0A%20%20%20%20%3ClinearGradient%20id%3D%27g%27%20x1%3D%270%27%20y1%3D%270%27%20x2%3D%271%27%20y2%3D%271%27%3E%0A%20%20%20%20%20%20%3Cstop%20offset%3D%270%25%27%20stop-color%3D%27%231d3557%27/%3E%0A%20%20%20%20%20%20%3Cstop%20offset%3D%27100%25%27%20stop-color%3D%27%23457b9d%27/%3E%0A%20%20%20%20%3C/linearGradient%3E%0A%20%20%3C/defs%3E%0A%20%20%3Crect%20width%3D%27420%27%20height%3D%27220%27%20fill%3D%27%230b132b%27/%3E%0A%20%20%3Ccircle%20cx%3D%2790%27%20cy%3D%27110%27%20r%3D%2764%27%20fill%3D%27url%28%23g%29%27%20stroke%3D%27%23f1faee%27%20stroke-width%3D%276%27/%3E%0A%20%20%3Ccircle%20cx%3D%2790%27%20cy%3D%27110%27%20r%3D%2714%27%20fill%3D%27%23f1faee%27/%3E%0A%20%20%3Cpath%20d%3D%27M168%2070%20h200%20q20%200%2020%2020%20v40%20q0%2020-20%2020%20h-80%20l-30%2030%20v-30%20h-90%20z%27%20fill%3D%27%23a8dadc%27%20stroke%3D%27%23f1faee%27%20stroke-width%3D%276%27%20stroke-linejoin%3D%27round%27/%3E%0A%20%20%3Ctext%20x%3D%27188%27%20y%3D%27108%27%20font-family%3D%27Verdana%2C%20sans-serif%27%20font-size%3D%2732%27%20fill%3D%27%230b132b%27%20font-weight%3D%27bold%27%3EMadao%3C/text%3E%0A%20%20%3Ctext%20x%3D%27188%27%20y%3D%27140%27%20font-family%3D%27Verdana%2C%20sans-serif%27%20font-size%3D%2718%27%20fill%3D%27%231d3557%27%3EChaos%20%26amp%3B%20Recovery%3C/text%3E%0A%3C/svg%3E)

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
  "message_jp": "...",
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
