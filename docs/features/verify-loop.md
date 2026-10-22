# Verify loop

`internal/verify` runs an agent, then runs the target repository's checks. If a check fails, it sends the failing output back to the agent and tries again.

## Reach it

```go
import "agentrun/internal/verify"
```

The entry point is `verify.Loop(ctx, verify.Config{...})`. It returns a `verify.Outcome` with a status and one entry per attempt.

| Status | Meaning |
|---|---|
| `passed` | All checks passed in some attempt. |
| `exhausted` | `MaxAttempts` ran out while checks still failed. |
| `no_progress` | The last check failed with the same result as the previous attempt, so the loop stopped early. |

`verify.SaveOutcome(dir, outcome)` writes `outcome.json` into `dir`.

## Reproduce a fake-agent run

No real agent is needed. The test below uses a fake agent that writes `done.txt` on its second call, then saves the outcome and a decision-log entry per attempt:

```sh
go test -race -count=1 -v -run TestLoopEvidenceDirectoryMatchesAttempts ./internal/verify/
```

Read the test in `internal/verify/evidence_test.go` to see how the pieces fit together.

## Gotchas

- Checks run with `sh -c` in `Config.Dir`, and the output kept is the last 4096 bytes. Failures that print early are cut off.
- `RunChecks` stops at the first failing check. Later checks never run, so their results are absent from `Attempts`.
- The repair prompt names the failing check and includes its output. It tells the agent not to weaken or delete checks, but nothing enforces that. A passing result is only as trustworthy as the checks.
- `ErrNoChecks` and `ErrInvalidAttempts` are returned before any agent call.
- `outcome.json` is written with mode `0600`.
