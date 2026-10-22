# Self-build backlog

Each item is one prototype run (`agentrun-proto submit`) against this repository. A run starts from committed HEAD and leaves its result on a branch `agentrun/<id>`. A person and Claude review the diff, run the checks, and merge before the next item starts. Items run in order because later items depend on earlier ones.

Shared rules for every item: test first, no new comments, `gofmt -l .` prints nothing, `go vet ./...` passes, and `go test -race -count=1 ./...` passes. Touch only the files the item names.

| # | Item | Depends on | Done when |
|---|---|---|---|
| 1 | Save verify outcome to disk | none | `verify.SaveOutcome(dir, outcome)` writes `outcome.json`; a test reads it back and checks status and attempt count |
| 2 | Append-only decision log | none | `internal/decisionlog` appends one JSON line per entry in order; empty decision is rejected; tests read the file back |
| 3 | `agentrun-dev` skeleton with `--json` | none | `cmd/agentrun-dev` dispatches subcommands; unknown command prints the 4-line error; `--json` flag is parsed globally |
| 4 | `agentrun-dev verify` | 3 | `agentrun-dev verify <dir> name=cmd ...` runs checks and exits 0 on pass, 1 on fail; `--json` prints the results |
| 5 | `agentrun-dev doctor` | 3 | Reports `git`, `go`, and `sh` on PATH; a missing tool produces the 4-line error from `internal/errors` |
| 6 | `agentrun-dev evidence <dir>` | 1, 3 | Prints the status and attempts from `outcome.json`; a missing file gives the 4-line error |
| 7 | Name, labels, and memory in `RunArgs` | none | `Spec` gains `Name`, `Labels`, and `Memory`; `RunArgs` emits `--name`, sorted `--label k=v`, and `--memory` when set; a label key or value with `=` or `,` is rejected; `internal/container/args.go`, `container.go`, `validate.go`, and their tests only |
| 8 | Apple `container` backend | 7 | `internal/container/apple` implements `container.Backend` by running `container run` attached; exit code comes back in `Result`; output goes to an `io.Writer`; a fake `container` script in the test temp dir checks args, exit codes, and cancellation |
| 9 | Egress allowlist proxy | none | `internal/egress` is a CONNECT-only HTTP proxy that tunnels to allowed host names and returns 403 for others, including IP literals; non-CONNECT methods get 405; tests use a local TCP echo server and `localhost` |

Items 7–9 implement `docs/decisions/apple-container-backend.md`. Live container checks (`go test -tags capability ./test/capability/...` and a real prompt) are run by hand after these items merge.

## Checking a run

```sh
AGENTRUN_PROTO_DATA=<scratch>/proto-data \
AGENTRUN_AGENT_CMD='claude -p --permission-mode acceptEdits' \
prototype/agentrun-proto submit "<item text>" --repo .
```

Then `result <id>`, `logs <id>`, and read the diff on `agentrun/<id>`. Run the checks from the rules above on that branch before merging.
