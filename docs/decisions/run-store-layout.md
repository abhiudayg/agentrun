# Decision: run store layout

Status: accepted for Step 5, pending the Step 5 review.

## Options compared

**A. One directory per run.** `runs/<id>/` holds `task.md`, `base`, `status`, `logs/`, `worktree/`, and later `outcome.json`. `list` scans the `runs/` directory. This is what the Phase 0 prototype does.

**B. One index file for all runs.** A single `runs.json` holds every run's record and is rewritten on each change. Each run directory still holds its logs and worktree.

## What was tried

A throwaway comparison lives outside the repository, in the session scratchpad. It is not part of the product.

- Option B, with the index truncated mid-write, failed to load: `index unreadable: unexpected end of JSON input`. Every run is lost with it.
- Option A, with a status file written through a temporary file and `rename`, lists every run. The test checks that the scan completes. It does not show that a corrupt status file is recovered. That case is still untested.

## Decision

Use option A. Write each status update to a temporary file and rename it over `status`, so a crash leaves the old or the new value, never a partial one.

## Why

- A single index is one point of failure. A torn write loses every run, and the capability report says recovery after a service restart is not tested.
- A directory per run keeps each run's state separate. A damaged run affects only itself.
- It matches the prototype, so existing runs and tests still describe the real layout.

## Costs

- `list` reads one file per run. That is fine for a single-user local tool. If it becomes slow, add a cache that can be rebuilt from the directories. Do not make the cache the source of truth.
- Status changes need the temporary-file-and-rename step in every writer.

## Not decided here

- Whether `outcome.json` is rewritten or appended.
- Locking between the supervisor and `cancel`. This belongs with the cancel race in [../run-api.md](../run-api.md).
