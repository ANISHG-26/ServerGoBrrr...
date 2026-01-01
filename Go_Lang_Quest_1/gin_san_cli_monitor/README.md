# Gin-san CLI Monitor — Madao watcher

![Gin-san monitor](../../assets/gin-san-monitor.jpg)

Terminal dashboard that pings one or more health endpoints (built for the Madao API) and rolls up latency/error stats with a grumpy Gin-san commentary.

## Run

- Basic: `go run . --urls=http://localhost:8080/health`
- Multiple targets: `go run . --urls=http://localhost:8080/health,http://localhost:8081/health -c 12`
- Stable view (no screen clear): `go run . --urls=http://localhost:8080/health --no-clear`
- Verbose (last responses): `go run . --urls=http://localhost:8080/health -v`

## Flags

- `--urls` comma-separated health URLs to watch (default `http://localhost:8080/health`).
- `--interval` delay between ticks (default `2s`).
- `--timeout` per-request deadline (default `1200ms`).
- `-c` max concurrent checks per tick (default `8`).
- `-n` number of ticks (default `0` = run forever).
- `-v` verbose: show last responses/JSON for each target.
- `--no-clear` skip terminal clear between ticks (better for logs).

## Dashboard quick guide

- Columns: `STATE` (ok/degraded/madao_spiral/net_err/parse_err), `HTTP`, `LAT(ms)` from payload, `WALL` real duration, `OK%` per-target success rate, `URL`.
- Final line prints a one-line Gin-san comment per target.
- `-v` prints the last tick’s response/JSON or error per target.
- After exit, a compact overall summary is printed (counts + latency stats).

## Practice drills (pair with Madao API chaos)

- Tight timeout drill: `go run . --urls=http://localhost:8080/health --timeout=500ms --interval=1s -v` while Madao API runs with high `-slow-pct`.
- Error-burst drill: Madao API with `-fail-pct 40` and monitor with `--interval=2s --timeout=1500ms` to watch spiral vs. degraded counts.
- Multi-target drill: point `--urls` at two Madao instances (one chaotic, one stable) and compare `OK%` per row.

## Notes

- Works best against the Madao API health schema (`status`, `latency_ms`, messages, request_id) but any similar JSON will render.
- ANSI colors/emojis are used; disable terminal filtering to keep icons visible.
