# AgentRun

AgentRun is designed to run agents you don't fully trust, and to accept only results that proved themselves. Each task is meant to run in its own isolated container and git worktree, with its result accepted only after the project's checks pass on the agent's changes. The engine is not built yet (see Status). It is intended to be one binary, `agentrun`, with two modes:

- **Lite** (default): one developer, one machine. Each task runs in its own isolated container and git worktree, and you can check status, read logs, cancel, and collect a reviewable branch or patch.
- **Full**: a small team shares one agent host, with a shared queue, per-user identity, quotas, an audit trail, and organization policy. Full is not built yet.

## Status

**Pre-alpha. The `agentrun` command does not exist yet.** This repository currently contains:

- shared foundations for the mode flag, policy digest, and provenance record (`internal/mode`, `internal/policy`, `internal/record`);
- the standard error format and catalog (`internal/errors`);
- an isolation capability suite that runs real containers against Apple's `container` CLI (`test/capability`);
- Phase 0 validation documents in `docs/`.

Nothing here runs a coding agent yet. The engine, run store, container backend, and CLI come after Phase 0 is complete.

## Requirements

- Go 1.27.2 (see `go.mod`)
- For the capability suite only: macOS on Apple silicon, Apple's `container` CLI with its service running (`container system status` reports `running`), and network access to pull `alpine:3.20`

## Running the tests

Unit tests run anywhere and do not need containers:

```sh
go test -race -count=1 -cover ./...
```

The capability suite runs real containers and is excluded from the default run by a build tag:

```sh
go test -tags capability -count=1 -v -timeout 10m ./test/capability/
```

It takes about 35 seconds. Each check removes its own containers and networks. If a run is interrupted, look for leftovers with `container list --all` and `container network list`; names start with `spike-` and `spike-net-`.

Check formatting and static analysis before committing:

```sh
gofmt -l .
go vet ./...
```

`gofmt -l .` must print nothing.

## Layout

```
internal/errors   standard error format (what, cause, fix, docs) and the error catalog
internal/mode     Lite and Full modes, resolution from flag and organization config, Full-only gate
internal/policy   sha256 digest of canonical policy bytes
internal/record   run provenance: mode, host, and policy digest, validated on construction and on decode
test/capability   Phase 0 spike: checks Apple container behavior the design depends on (build tag: capability)
prototype/        wizard-of-oz shell prototype for the Phase 0 study, with its Go test driver
docs/             Phase 0 reports and runbooks
```

## Prototype (Phase 0 study)

`prototype/agentrun-proto` is the shell prototype the Phase 0 study uses. It is not the product. It runs the agent directly on the host, with no isolation (see the limits below).

```sh
prototype/agentrun-proto submit "Fix the flaky auth tests" --repo ~/code/app
prototype/agentrun-proto list
prototype/agentrun-proto logs <id>
prototype/agentrun-proto result <id>
prototype/agentrun-proto cancel <id>
```

- `submit` refuses a repository with uncommitted changes unless `--include-dirty` is given. The run starts from HEAD either way; uncommitted edits are not copied in.
- Each run gets a worktree and branch `agentrun/<id>` under `$AGENTRUN_PROTO_DATA/runs/<id>/`. The agent's changes are committed on that branch, and `patch.diff` is written when there are changes.
- `AGENTRUN_AGENT_CMD` sets the agent command. It defaults to `claude -p`, and the task text is passed on stdin, never interpolated into a shell.
- Each command appends its name, a run ID, and a timestamp to `$AGENTRUN_PROTO_DATA/usage.log`. Prompt text is never logged.

Requirements: bash, and git. Tested with bash 3.2.57 (macOS system bash) and git 2.54.0. The PRD targets git 2.30 or later; that version has not been tested.

Limits:
- **No isolation.** The agent runs with your user's permissions and can read and write anything you can. A worktree's `.git` file points into the main repository, so an agent can reach its hooks directory. Run the study only with agents and repositories participants have agreed to use. The container backend (Phase 1) is what provides isolation.
- Repository hooks are disabled for the worktree creation and the run's commit.
- `cancel` marks the run cancelled, then signals the agent's process group. There is a short race: if a cancel arrives while a queued run is starting, the run can be marked running again and the agent starts. Cancel once the run shows `running` with an agent PID, and check the status again.

## Phase 0 documents

- [`docs/phase0-capability-report.md`](docs/phase0-capability-report.md): what the container CLI does and does not support, with tested versions.
- [`docs/phase0-teardown.md`](docs/phase0-teardown.md): nearby projects checked against their own READMEs, and what could not be verified.
- [`docs/phase0-human-validation.md`](docs/phase0-human-validation.md): how to run the interviews, prototype study, and usability sessions that an agent cannot do.

## Known limits of the current spike

- The container runtime does not keep a detached container's exit code. A supervisor must record it while it is attached.
- Per-container network allowlisting is not available. The planned `model-api` mode needs a host-side proxy.
- Containers run as root with a writable root filesystem by default. Every run must set the user, `--read-only`, and a tmpfs `/tmp` explicitly.

## License

Not yet chosen. The PRD lists the license as an open decision before public release.
