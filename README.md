# pomp

**Pre-omp project picker.** A tiny terminal launcher that remembers the directories
you use with `omp` and lets you jump back into any of them from a Bubble Tea TUI.

Run `pomp` anywhere in your shell → pick a project → `omp` starts there.

## Features

- Lists recent `omp` project directories, newest first
- Current directory always pinned at the top
- Fuzzy filter (`/`)
- Filters out stale entries (deleted folders)
- Zero configuration — reads `~/.omp/agent/sessions` directly
- Propagates `omp`'s exit code
- Single static binary, Windows / macOS / Linux

## Install

```sh
go install github.com/Boyarinov/pomp@latest
```

This drops `pomp` into `$GOPATH/bin` (default `~/go/bin` or `%USERPROFILE%\go\bin`).
Make sure that directory is on your `PATH`.

### From source

```sh
git clone https://github.com/Boyarinov/pomp
cd pomp
go build -o pomp .
```

## Usage

```
pomp
```

| Key                      | Action                       |
| ------------------------ | ---------------------------- |
| `↑` / `↓` / `k` / `j`    | Navigate                     |
| `/`                      | Filter                       |
| `Enter`                  | Launch `omp` in selected dir |
| `q` / `Esc` / `Ctrl+C`   | Cancel                       |

### Environment

| Variable            | Default                 | Purpose                              |
| ------------------- | ----------------------- | ------------------------------------ |
| `POMP_SESSIONS_DIR` | `~/.omp/agent/sessions` | Override the `omp` sessions location |

## How it works

`omp` stores one subdirectory per working directory under `~/.omp/agent/sessions`,
each containing `.jsonl` session logs. The first line of every log is a JSON event
with the session's `cwd` and optional `title`. `pomp` reads those headers, dedupes
by path, drops directories that no longer exist, and sorts by the newest log's
mtime. On selection it runs `exec.Command("omp")` with `cmd.Dir` set to the chosen
path and inherits stdio, so `omp` takes over your terminal as if you'd launched it
yourself.

## Development

```sh
go vet ./...
go test ./...
go build -o pomp .
```

## License

MIT — see [LICENSE](./LICENSE).
