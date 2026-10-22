# Decision: Apple container backend (Phase A2)

## Context

The capability report (test/capability) shows three constraints:

- A detached container does not keep its exit code, so runs are attached and the supervisor records the exit code.
- The internal network blocks all egress and the default network has no per-container allowlist. Agents still need their model APIs.
- Root filesystem, non-root user, tmpfs /tmp, memory limits, and digest-pinned images all work and must be used.

## Decisions

1. **Attached runs only.** `container run` without `-d`. The backend returns the process exit code in `container.Result`.
2. **Egress through a host-side allowlist proxy.** The container runs on the internal network with `HTTPS_PROXY` pointing at a proxy on the host. The proxy accepts CONNECT only, tunnels to hosts on the allowlist, and returns 403 for everything else, including IP literals. No direct route out is given to agents.
3. **Installers: accept drift, record the version.** Cursor and agy installers are not pinned. Each run records the installed agent CLI version in its run directory. Pinning is not part of A2.
4. **Real-prompt runs are manual.** Cursor and agy are run against a real prompt inside the Debian image by hand, with the user's go-ahead. Each run is recorded as an attempt.

## Not in A2

- Wiring the backend into the verify loop or submit. That is the gated engine step (implementation plan Step 5).
- Per-run concurrency (Step 6).
- Adapters for claude, codex, and opencode (A3).

## Open

- The address the container uses to reach the host proxy. It is found empirically in the real-prompt step and recorded here.
