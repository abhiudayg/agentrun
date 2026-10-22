# Developer CLI

`agentrun-dev` is a development tool for this repository. It is not the `agentrun` product binary.

## Reach it

```sh
go run ./cmd/agentrun-dev <command> [args]
```

## Commands

| Command | What it does | Example |
|---|---|---|
| `doctor` | Reports whether `git`, `go`, and `sh` are on PATH. Exits 1 and prints the four-line error if one is missing. | `go run ./cmd/agentrun-dev doctor` |
| `verify <dir> name=cmd ...` | Runs each check with `sh -c` in `<dir>`, stopping at the first failure. Exits 1 on failure. | `go run ./cmd/agentrun-dev verify . vet="go vet ./..." test="go test ./..."` |
| `evidence <dir>` | Prints the status and per-attempt results from `<dir>/outcome.json`. | `go run ./cmd/agentrun-dev evidence /path/to/run` |

Add `--json` before the command name to get machine-readable output from `verify`.

```sh
go run ./cmd/agentrun-dev --json verify . vet="go vet ./..."
```

## Gotchas

- `--json` is a global flag. Put it before the command name, not after it.
- A check is `name=cmd`. Both parts must be non-empty, and the split is on the first `=`, so the command may contain `=`.
- `verify` runs the checks only. It does not call an agent. The agent loop is in `internal/verify` (see [verify-loop.md](verify-loop.md)).
- `evidence` reads `outcome.json`, which only `verify.SaveOutcome` writes. A run directory without that file gives the four-line error.
- Error messages point to `https://agentrun.dev/docs/troubleshooting#commands`. That anchor does not exist yet.
