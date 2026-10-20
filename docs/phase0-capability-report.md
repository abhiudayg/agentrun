# Phase 0 capability report: Apple `container` CLI

Source assumptions: AgentRun Lite PRD v1.1, §2 ("Verified-before-building assumptions") and §5 (Phase 0 capability spike).

## Tested baseline

| Item | Version |
|---|---|
| Host | macOS 27.0.1 (build 26A434), arm64 |
| Apple `container` CLI and server | 1.5.0 (release) |
| Go | 1.27.2 |
| Test image | `alpine:3.20`, digest `sha256:d9e853e87e55526f6b2917df91a2115c36dd7c696a35be12163d44e6e2a4b6bc` |

Reproduce with `go test -tags capability -count=1 -v ./test/capability/` (about 35 s; needs the container service running). The suite is excluded from the default `go test ./...`.

## Results

| PRD §2 assumption | Result | Evidence |
|---|---|---|
| Minimum macOS and Apple silicon | **Open.** Only macOS 27.0.1 arm64 was available. | Not testable on this host. |
| CLI output is machine-readable | **Pass.** `list --format json` and `inspect` return parseable JSON. | `TestListingIsMachineReadableJSON` |
| CPU and memory limits | **Partial.** Memory limit is reflected in `inspect` and an out-of-memory process is killed with exit 137. CPU limit is reflected in config only; not measured under load. A 200 MiB minimum applies, so `-m 128M` is rejected. | `TestMemoryLimitIsAppliedToContainer`, `TestMemoryBelowMinimumIsRejected`, `TestRunawayMemoryIsKilledWithExit137` |
| Read-only mount and single writable worktree | **Pass.** Writable worktree mount reaches the host; read-only mount rejects writes. | `TestWorktreeMountIsWritableAndReadOnlyMountRejectsWrites` |
| Writes blocked outside mounts | **Pass with a requirement.** `--read-only` blocks root filesystem writes. `/tmp` is then also blocked and must be provided with `--tmpfs /tmp`. | `TestReadOnlyRootfsBlocksWritesOutsideMounts`, `TestTmpfsProvidesWritableTmpUnderReadOnlyRoot` |
| Non-root workload | **Pass with a requirement.** `-u 1000:1000` is enforced. The default user is root, so the non-root setting must be explicit in every run. | `TestNonRootUserCannotWriteRootOwnedPaths` |
| Network `none` | **Pass via `--internal` network.** `container network create --internal` blocks egress, including to arbitrary internet addresses. There is no `--network none` flag. | `TestInternalNetworkBlocksEgress` |
| Restricted egress (`model-api`) | **FAIL for the design as written.** The default network reaches arbitrary destinations; no per-container allowlist exists. The PRD's fallback applies: a host-side proxy is required and adds a component. | `TestDefaultNetworkHasNoPerContainerAllowlist` |
| Labels and discovery after restart | **Partial.** Labels are stored and readable through `inspect`. `list` has no label filter, so discovery must parse JSON. A detached container keeps running after its client exits. Recovery after a host or service restart is **not tested**. | `TestLabelsAreStoredAndReadableFromInspect`, `TestDetachedContainerKeepsRunningAfterClientExits` |
| Exit code and log semantics | **FAIL for detached runs.** Foreground runs return the exit code (verified 7). Detached runs expose no exit code through `inspect`, so a supervisor that loses its attachment cannot recover it. Logs captured stdout and stderr; separation of the two streams was not asserted. | `TestForegroundExitCodePropagates`, `TestDetachedExitStatusIsUnavailableFromInspect`, `TestDetachedLogsCaptureStdoutAndStderr` |
| Signal handling | **Pass.** A process ignoring SIGTERM is killed at the grace period. A process that honors SIGTERM exits promptly. | `TestSIGTERMIgnoringProcessIsKilledAtGracePeriod`, `TestSIGTERMHonoringProcessExitsPromptly` |
| Pull by digest | **Pass.** Digest-pinned references run. Pulls fetch a multi-platform index by default, so arm64 must be selected explicitly. | `TestDigestPinnedImageRuns` |
| Host sleep and service restart | **Open.** Not tested. A sleep-wake test and a service-restart test are needed; the restart test disrupts a running `buildkit` container. | None |

## Findings that change the design

1. **Model-API egress needs a host-side proxy.** Per-container allowlisting is unavailable. Make `model-api` a P1 feature that depends on the proxy, as the PRD anticipates.
2. **Exit status must be recorded by the supervisor.** The container runtime does not keep it for detached containers. Runs whose attachment is lost must be marked `interrupted` (RUN-10), not guessed.
3. **Every run must set user, read-only root, and tmpfs `/tmp` explicitly.** The defaults (root user, writable root) fail the PRD's isolation requirements.

## Gate status

Lite PRD v1.1 §5 says: "If any P0 item fails, the decision in §5 of this revision is re-evaluated before further work." Items 2 and 3 above, and the egress row, are failures of P0 items. **The go/no-go decision therefore needs a human decision before Phase 1 continues.**

The Phase 0 exit gates in the Lite PRD v1.0 §12 are not met by this report:

- Developer research with 12–15 participants: not started.
- Two-week prototype use by 5 participants: not started.
- Usability sessions on three clean machines: not started.
- Interviews with 8–12 maintainers or platform engineers: not started.
- Competitive teardown: incomplete. Warren (`jayminwest/warren`, MIT) exists and covers isolation, lifecycle, recovery, and git delivery, which overlaps the Lite promise. Its listed providers do not include Apple's `container`. Agentry's repository is empty. Details and sources are in `docs/phase0-teardown.md`. The "capability gap" exit criterion is not met until the Apple `container` gap is verified by running Warren on a Mac.

## PRD inconsistencies to resolve

- The Lite v1.1 spike is described as "V1 to V15", but §2 defines 10 rows with no numbering.
- Phase 0 is weeks 0–3 in v1.0 and weeks 0–4 in v3.0.
