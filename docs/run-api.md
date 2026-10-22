# Run API (sketch)

This is a sketch for the engine that comes after Phase 0. Nothing here is implemented. It starts from the command shape the Phase 0 prototype already uses (`submit`, `list`, `logs`, `cancel`, `result`), so the CLI stays familiar. Where the engine needs more, the addition is marked.

## Commands

```sh
agentrun submit "<task>" [--repo PATH] [--include-dirty]
agentrun list
agentrun logs <id>
agentrun cancel <id>
agentrun result <id>
agentrun evidence <id>      # new: per-attempt check results, from outcome.json
```

| Command | Behavior |
|---|---|
| `submit` | Creates a run from committed HEAD of `--repo`, on branch `agentrun/<id>`. Refuses a dirty tree unless `--include-dirty` is given. Prints the run ID and the next command. |
| `list` | One line per run: ID, status, task. Scans the run directory; see [decisions/run-store-layout.md](decisions/run-store-layout.md). |
| `logs <id>` | Prints the agent log captured so far. |
| `cancel <id>` | Marks the run cancelled and stops its agent. A run that has already finished is left alone. |
| `result <id>` | Prints the final status and the branch to review. |
| `evidence <id>` | New. Prints the verify outcome (`passed`, `exhausted`, `no_progress`) and each attempt's checks. Reads the same `outcome.json` the verify loop writes. |

## Run states

`queued` → `running` → one of `succeeded`, `failed`, `cancelled`.

A run that reaches `running` and then loses its supervisor must be recovered from its last written state. The capability report says a detached container does not expose its exit code, so the supervisor records the exit code while it is attached.

## Errors

Every error prints the four-line format from `internal/errors`: what happened, the cause, the fix, and a docs link.

## Open questions

- Whether `evidence` belongs on the product binary or only on `agentrun-dev`. Decide before Step 5.
- How `cancel` reports the race the prototype README describes, where a cancel can arrive while a run is starting.
