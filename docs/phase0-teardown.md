# Phase 0 teardown: nearby projects

Scope: the Lite PRD v1.0 §12 Phase 0 teardown. Each row is checked against the project's own README or repository page where it could be fetched. Anything not checked is marked as such. Checked on 2026-10-20.

## Verified against primary sources

| Project | Source | What it says it does | Isolation | Recovery after restart | Container backends | Single-user local mode | License |
|---|---|---|---|---|---|---|---|
| **Warren** (`jayminwest/warren`) | [README](https://github.com/jayminwest/warren) | Self-hosted control plane that runs coding agents as managed workloads. Handles isolation, lifecycle, spend limits, recovery, and git delivery. | Fresh worktree or clone per run, inside a sandbox the operator chooses | "Watchdogs reconcile lost processes and pods." Server restart recovery not described. | `local`, `docker`, `k8s`. No Apple `container`. | Partly: `warren up` on one machine with `local` runtime. Not a named mode. | MIT |
| **OrchestratorInc/agent-orchestrator** | [repository](https://github.com/OrchestratorInc/agent-orchestrator) | "Give every coding task its own agent, workspace, and feedback loop." Local desktop workspace for planning, running, and monitoring agents. | Per-worker branch and worktree for git-backed work. No sandbox or container. | Not described. | None described. | Yes (local desktop app). | Apache-2.0 |

Notes:
- Warren's README states its scope as "one operator or small trusted team sharing one deployment with one bearer credential." It is therefore not a multi-user product in the sense of PRD Full, and it overlaps Lite directly.
- Warren reports version 0.19.3 and "pre-1.0". Those figures come from its README text; the repository's release tags were not checked.

## Could not be verified

| Project | Why |
|---|---|
| **Agentry** (`cubxxw/agentry`) | The repository page returned "This repository is empty." A search listed it as an agent orchestration CLI in Go, but no README was available. The PRD's reference to "Agentry" cannot be checked from this source. |
| **ComposioHQ/agent-orchestrator** | The URL resolved to `OrchestratorInc/agent-orchestrator` instead. The ComposioHQ repository identity is unconfirmed. The earlier claim that it provides per-session worktrees was attributed to the wrong project. |
| Other search hits (AgentWrapper, CodeMachine, nekocode/agent-worktree) | Listed by search only. Not fetched. |

## What this means for the Phase 0 gap criterion

The PRD's exit gate requires "at least one capability gap that existing tools do not cover well." The verified evidence points the other way for the core promise:

- Warren covers isolation, lifecycle, recovery, git delivery, and a local runtime. A user who wants that on one machine can run it today.
- A possible gap remains in the **Apple `container` backend**. Warren's listed providers are local, Docker, and Kubernetes. Its `local` provider is not documented as using VM-per-container isolation. This must be verified by running Warren on a Mac before it counts as a gap.
- Label-based discovery after restart and the reconciliation behavior after a server restart are not described in Warren's README.

Until those points are tested, the gap criterion is **not met**.
