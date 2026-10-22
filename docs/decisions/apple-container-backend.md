# Decision: Apple container backend (Phase A2)

## Context

The capability report (test/capability) shows three constraints:

- A detached container does not keep its exit code, so runs are attached and the supervisor records the exit code.
- The internal network blocks all egress and the default network has no per-container allowlist. Agents still need their model APIs.
- Root filesystem, non-root user, tmpfs /tmp, memory limits, and digest-pinned images all work and must be used.

## Decisions

1. **Attached runs only.** `container run` without `-d`. The backend returns the process exit code in `container.Result`.
2. **Egress through an allowlist proxy (placement under review).** The proxy accepts CONNECT only, tunnels to `host:port` entries on the allowlist, and returns 403 for everything else. Agents get no direct route out.
   - The original plan put the proxy on the host. This is not possible. A container on the internal network has no route to the host. Its gateway is 192.168.128.1, which is not a host address, and the host's only container-facing address is 192.168.64.1 on the default network, which has full egress.
   - Tested: a container can join both networks. An internal-only client reached a service on the dual-homed container at its internal address (192.168.128.x). Direct egress from the internal network stayed blocked.
   - Tested: containers on the internal network do not resolve each other by name, so the agent must be given the proxy's IP address.
   - Proposed: run the proxy as a dual-homed sidecar container on the default and internal networks. Agents run on the internal network only, with `HTTPS_PROXY` set to the sidecar's internal IP. This needs a Go proxy binary in an image, and the launcher must read the sidecar's IP after start. Not yet approved.
3. **Installers: accept drift, record the version.** Cursor and agy installers are not pinned. Each run records the installed agent CLI version in its run directory. Pinning is not part of A2.
4. **Real-prompt runs are manual.** Cursor and agy are run against a real prompt inside the Debian image by hand, with the user's go-ahead. Each run is recorded as an attempt.

## Not in A2

- Wiring the backend into the verify loop or submit. That is the gated engine step (implementation plan Step 5).
- Per-run concurrency (Step 6).
- Adapters for claude, codex, and opencode (A3).

## Open

- The address the container uses to reach the host proxy. It is found empirically in the real-prompt step and recorded here.
