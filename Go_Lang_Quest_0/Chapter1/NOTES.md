# Chapter1 Notes

- Entry point: `Chapter1/main.go` defines `package main` for an executable.
- Imports `fmt` solely for console output.
- `func main()` is the program entry and calls `fmt.Println` to print `Go Infra Quest: Level 0. GG Noobs!` with a newline.
- Run from repo root after module init: `go run "./Chapter1"` or `go run "Chapter1/main.go"`.
- Module init (one-time if missing): `cd Chapter1 && go mod init Chapter1` to create `go.mod` for module-aware builds.
- Build binary: `cd Chapter1 && go build` (produces `Chapter1.exe` in that folder on Windows).
