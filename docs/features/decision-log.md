# Decision log

`internal/decisionlog` appends one JSON object per line to a file. It is meant to record what was decided and what evidence supports it, so a later reader can check it against the agent transcript.

## Reach it

```go
import "agentrun/internal/decisionlog"

err := decisionlog.Append(path, decisionlog.Entry{
	Decision: "use JSON lines",
	Reason:   "append-friendly",
	Evidence: "backlog item 2",
	Outcome:  "accepted",
})
```

## Fields

| Field | Set by | Notes |
|---|---|---|
| `time` | `Append` | Overwritten with the current UTC time. Callers cannot set it. |
| `decision` | caller | Required. Whitespace-only is rejected with `ErrEmptyDecision`. |
| `reason`, `evidence`, `outcome` | caller | Free text. |

## Gotchas

- Each entry is written with one `Write` call on a file opened with `O_APPEND`. The package does not lock, so concurrent writers from several processes are not coordinated.
- A missing decision returns `ErrEmptyDecision` and writes nothing.
- The file is created with mode `0600` if it does not exist.
