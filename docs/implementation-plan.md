# Implementation plan: pstack-style delivery for AgentRun

Source: "The Complete Guide to pstack" (parts 1–3). This plan maps its ideas onto AgentRun. Phase 0 gates come first; nothing here changes the frozen Phase 0 prototype.

## Current state

- `internal/verify` exists: fail-fast checks, bounded repair loop, no-progress stop. Tests pass (92.2% coverage). Uncommitted.
- `prototype/agentrun-proto` is the Phase 0 study instrument and its command surface is frozen.
- The Apple `container` capability report has open P0 items (egress allowlist, detached exit codes, host sleep and service restart). The go/no-go decision needs a human.

## Article idea → AgentRun

| Article idea | AgentRun meaning | Gated on Phase 0? |
|---|---|---|
| Verification is the bottleneck | `internal/verify` runs the target repo's checks and feeds failures back | No |
| Build the lever | Dev CLI for this repo: `doctor`, `verify`, `evidence`; `--json`, `--dry-run`, actionable errors | No (dev-only) |
| Evidence beats explanations | Each attempt saves check output, exit codes, and diff under a run directory | No |
| Show-me-your-work | Append-only decision log, checked against the agent transcript | No |
| Feature Map | `docs/features/`: one page per feature, how to reach it, gotchas | No |
| Stop the line | Stages submit → worktree → agent → verify → repair → commit; any stage ends the run with a reason | Partly |
| Parallel cloud agents | Per-run isolation via Apple `container` | Yes (Phase 1) |
| Readme-driven design | Write run API docs before engine code | No |
| Prototype before committing | Compare two run-store layouts with throwaway code; record the choice | No |
| Plan with proof | Each plan item names the check that must run, not just tests | No |
| Recall / teach | Decision notes so later sessions start with context | No |
| Build the lever for migrations | Codemod plus checker, first unit done by hand | Not needed yet |

## Steps

1. **Library evidence and decision log (no gate).** Save per-attempt evidence to disk. Add an append-only JSON-lines decision log. Done when a fake-agent test produces an evidence directory and the log matches the attempts.
2. **Dev CLI (no gate).** `cmd/agentrun-dev` with `verify`, `doctor`, `evidence <run>`. `--json` output, `--dry-run` for writes, errors in the 4-line format from `internal/errors`. Done when each command has a test and runs on this repo.
3. **Feature map and design docs (no gate).** `docs/features/` pages. A readme-first sketch of the run API. Done when a reader can reproduce a fake-agent run from the docs.
4. **Prototype two run-store layouts (no gate).** Throwaway code, a decision note, and a recorded choice.
5. **Engine (gated).** Wire `Loop` into submit with stop-the-line stages. Run each agent in an isolated container using the capability report's constraints: explicit non-root user, read-only root, tmpfs `/tmp`, exit code recorded while attached. Done when the capability suite and a submit-to-patch run both pass.
6. **Parallel runs (after 5).** Several isolated agents, each with its own evidence directory.

## Verification rule

Every step ends with a run of the real check, recorded with `agent-brain verify -- <cmd>`. A step is done when that check passes after its last edit, not when the code compiles.

## Decisions needed

1. Approve Steps 1–4 to start now. They do not change any frozen artifact.
2. Dev CLI as a separate binary, or a subcommand of the future `agentrun` binary.
3. Whether Step 5 waits for the Phase 0 human gates or starts once the P0 isolation items are resolved.
