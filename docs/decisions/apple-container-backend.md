# Decision: Apple container backend (Phase A2)

## Context

The capability report (test/capability) shows three constraints:

- A detached container does not keep its exit code, so runs are attached and the supervisor records the exit code.
- The internal network blocks all egress and the default network has no per-container allowlist. Agents still need their model APIs.
- Root filesystem, non-root user, tmpfs /tmp, memory limits, and digest-pinned images all work and must be used.

## Decisions

1. **Attached runs only.** `container run` without `-d`. The backend returns the process exit code in `container.Result`.
2. **Egress through an allowlist sidecar.** The proxy accepts CONNECT only, tunnels to `host:port` entries on the allowlist, and returns 403 for everything else. Agents get no direct route out.
   - The original plan put the proxy on the host. This is not possible. A container on the internal network has no route to the host. Its gateway is 192.168.128.1, which is not a host address, and the host's only container-facing address is 192.168.64.1 on the default network, which has full egress.
   - Tested: a container can join both networks. An internal-only client reached a service on the dual-homed container at its internal address (192.168.128.x). Direct egress from the internal network stayed blocked.
   - Tested: containers on the internal network do not resolve each other by name, so the agent must be given the proxy's IP address.
   - Adopted: the proxy runs as a dual-homed sidecar (`images/egress`, binary `cmd/agentrun-egress`). Agents run on the internal network only, with `HTTPS_PROXY` set to the sidecar's internal address.
   - The sidecar must list the default network first. With the internal network first, its default route and DNS point at the internal gateway, and the upstream dial fails with 502.
   - Confirmed: agy and the Cursor CLI both honor `HTTPS_PROXY`. Through the sidecar, a request to an allowed host reaches the upstream and a request to any other host is refused with 403 and logged.
   - Endpoints seen: agy's model API is `generativelanguage.googleapis.com:443`. agy also tried `play.googleapis.com:443` and `antigravity-unleash.goog:443`; these were not allowlisted. Cursor's endpoints are `api2.cursor.sh:443` and `agentn.global.api5.cursor.sh:443`.
   - Not yet proven: a full authenticated prompt. Both CLIs need an API key, and neither key is available here (see below).
   - Auth: the host agy and Cursor logins live in the macOS keychain, not in files. The agy OAuth file copied into scratch did not log agy in. Real runs need `GEMINI_API_KEY` and `CURSOR_API_KEY`, or the user's approval for another method.
3. **Installers: accept drift, record the version.** Cursor and agy installers are not pinned. Each run records the installed agent CLI version in its run directory. Pinning is not part of A2.
4. **Real-prompt runs are manual.** Cursor and agy are run against a real prompt inside the Debian image by hand, with the user's go-ahead. Each run is recorded as an attempt.

## Not in A2

- Wiring the backend into the verify loop or submit. That is the gated engine step (implementation plan Step 5).
- Per-run concurrency (Step 6).
- Adapters for claude, codex, and opencode (A3).

## Open

- An authenticated prompt through the sidecar for agy and for Cursor. Needs API keys.
- The sidecar's internal address is assigned at start. The launcher reads it from `container inspect`. The launcher is not written yet, and the Apple backend's `Spec` has one network and no environment, so both need a change before submit can use the sidecar.
