# AgentRun

AgentRun runs coding-agent tasks as isolated, recoverable jobs on your own machine. It is intended to be one binary, `agentrun`, with two modes:

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
docs/             Phase 0 reports and runbooks
```

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
