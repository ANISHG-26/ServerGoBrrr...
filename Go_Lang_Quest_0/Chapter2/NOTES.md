# Chapter2 Notes

- Uses the standard library `flag` package to parse command-line flags.
- `level := flag.Int("level", 0, "Level number to display")` declares an int flag named `level` with default 0 and returns a pointer to store the parsed value.
- `flag.Parse()` must run before reading any flag values; afterwards `*level` gives the integer passed from the CLI.
- Output prints the chosen level: `fmt.Println("Go Infra Quest: Level", *level, ". GG Noobs!")`.
- Run with a custom level: `go run "./Chapter2" -level 3`; without the flag the default level 0 is used.
- Build the binary: `cd Chapter2 && go build` (creates `Chapter2.exe` in the same folder).
