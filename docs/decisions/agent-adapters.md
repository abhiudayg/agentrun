# Decision: agent adapter survey (Phase A0)

Status: survey, prototype runs, fixtures (A0.3), and the Linux install survey (A0.5) done. OpenCode and Codex have no fixture; both are blocked on their accounts. Cursor and agy Linux builds are unverified.

Scope: what each agent CLI needs to run non-interactively, what structured output it can emit, how it is kept read-only, and which environment variables carry its credentials. Results come from `--help` output on this machine and from one read-only listing run per agent through `agentrun-proto`.

## Versions and invocation (A0.1)

| Agent | Binary and version | Non-interactive command | Structured output | Read-only option |
|---|---|---|---|---|
| claude | `claude` 2.1.308 | `claude -p "<prompt>"` | `--output-format json` or `stream-json` (`stream-json` needs `--verbose`) | `--permission-mode plan` |
| opencode | `opencode` v2.1.2 | `opencode run "<message>"` | `--format json` | none in help; `--auto` only auto-approves |
| agy (Antigravity) | `agy` 1.4.3 | `agy -p "<prompt>"` | `--output-format stream-json` (also `json`, `text`) | `--mode plan` (produced a plan artifact, not a listing) |
| codex | `codex` codex-cli 0.141.0 | `codex exec "<prompt>"` | `--json` (JSONL events) | `-s read-only` |
| cursor | `agent` 2026.10.19-6bbf2c4 | `agent -p "<prompt>"` | `--output-format stream-json` (also `json`, `text`) | `--mode plan` or `ask` |

Notes:
- The Cursor binary is `agent`. The `cursor-agent` name points at the same install.
- Cursor's help says `-p` has access to all tools, including write and shell. Read-only depends on `--mode plan` or `ask`.
- Cursor and Codex may need a trust flag in a new workspace. Cursor needs `--trust`, otherwise it stops at "Workspace Trust Required".

## Prototype runs (A0.2)

Each run listed the files in a scratch repo with two files, `README.md` and `calc.py`, and was told not to modify anything. The prototype passes the task on stdin and runs the agent command with `sh -c`.

| Agent | Run | Result | Log size | Notes |
|---|---|---|---|---|
| claude | `bc23b4` | succeeded | ~27 KB | Loaded the user's full config: session hooks, MCP servers, about 80 tools. Log is noisy. |
| cursor | `2d08f4` | succeeded | ~28 KB | `agent -p --output-format stream-json --mode plan --trust "$(cat)"`. Also loaded the user's agent-brain routing. |
| agy | `3feee9` | succeeded | ~10 KB | Needs `--add-dir "$PWD"`. Without it, agy listed a different directory's files. Needed `--dangerously-skip-permissions` in default mode. |
| opencode | `faed0b` | failed | 0.8 KB | `provider.free-limit`, HTTP 429, retry after about 18 hours. Blocked until a provider is connected. |
| codex | `5f819c` | failed | 0.3 KB | `The 'gpt-5.5' model is not supported when using Codex with a ChatGPT account.` Blocked on model choice. |

Failed attempts that were fixed and not counted as results:
- `ba454c` (agy): `-p` took `--output-format` as its prompt. Fixed by putting the prompt directly after `-p`.
- `547c2e` (cursor): stopped at "Workspace Trust Required". Fixed with `--trust`.
- `ec1a1f` (agy): succeeded but listed a directory outside the run. Superseded by `3feee9`.
- `15949f` (codex): hung on an invalid refresh token. Cancelled. Re-login fixed the token, and the later run failed on the model.

## Auth environment variables (A0.4)

These names were checked against the `--help` text where possible. Anything marked "binary strings" was found by grepping the binary and has not been confirmed by a run.

| Agent | Variable | Source | Status |
|---|---|---|---|
| claude | `ANTHROPIC_API_KEY` | `--help` (`--bare` mode) | Confirmed by help. OAuth login also works interactively. |
| cursor | `CURSOR_API_KEY` | `--help` (`--api-key`) | Confirmed by help. `CURSOR_API_ENDPOINT` is optional. |
| codex | `OPENAI_API_KEY` | `codex login --help` (`--with-api-key`) | Confirmed by help. ChatGPT login also works. |
| codex | `CODEX_ACCESS_TOKEN` | binary strings | Not in help. Unverified. |
| agy | `GEMINI_API_KEY`, `GOOGLE_API_KEY` | binary strings | Not in help. Unverified. Login state not checked. |
| opencode | many provider keys, including `ANTHROPIC_API_KEY` and `GEMINI_API_KEY` | binary strings | Which key is needed depends on the configured provider. Not verified. |

No variable values were printed or stored. Only names were read.

## Linux container install (A0.5)

Read-only check: npm registry metadata and vendor docs. Nothing was installed into an image. The macOS binaries on this machine can't run in a Linux container, so each agent needs a Linux build.

| Agent | Linux install | Linux arm64 | Linux x64 | musl (Alpine) | Notes |
|---|---|---|---|---|---|
| claude | npm `@anthropic-ai/claude-code` 2.1.308, with per-platform optional packages | yes (`-linux-arm64`) | yes (`-linux-x64`) | yes (`-linux-x64-musl`, `-linux-arm64-musl`) | Node is not needed at runtime if the native package is used. |
| codex | npm `@openai/codex` 0.163.0, with per-platform optional packages | yes (`-linux-arm64`) | yes (`-linux-x64`) | not listed | Registry latest is 0.163.0; this machine runs 0.141.0. |
| opencode | npm `opencode-ai` 1.19.3, with per-platform optional packages | yes | yes | yes (`-linux-x64-musl`, `-linux-arm64-musl`) | A `-baseline` build exists for CPUs without AVX2. Registry latest is 1.19.3; this machine runs v2.1.2. |
| cursor (`agent`) | Vendor installer: `curl https://cursor.com/install -fsS \| bash`. Installs to `~/.local/bin`. | unverified | unverified | unverified | Source: [Cursor CLI installation docs](https://cursor.com/docs/cli/installation). Piping a remote script into bash runs it unreviewed, so the image build should download and pin it instead. |
| agy (Antigravity) | Third-party guides: `curl -fsSL https://antigravity.google/cli/install.sh \| bash`. Installs to `~/.local/bin`. | unverified | unverified | unverified | Sources are third-party blogs, not Google's docs. A VM with a default emulated CPU lacks AES and won't start `agy`. |

Image decisions this leaves for A2:
- Use one image per architecture. Apple `container` runs Linux images, so Apple silicon hosts need the arm64 packages.
- Pin each package version, and don't use the registry's latest.
- Install the CLIs at image build time, not at run time, so a run doesn't depend on the network for its tools.
- Cursor and agy need an arm64 Linux check before they're added to the image. That check was not run here, because it needs a Linux container.

Sources for the install details: npm registry metadata (`npm view`) for claude, codex, and opencode; [Cursor CLI installation docs](https://cursor.com/docs/cli/installation) for Cursor; and third-party Antigravity install guides for agy, which are not confirmed against Google's docs.

## Blockers and open items

- **OpenCode:** needs a connected provider or a subscription before any run.
- **Codex:** needs a model name the ChatGPT account supports. The default `gpt-5.5` is rejected.
- **Antigravity:** the adapter must always pass `--add-dir` with the workspace. Its read-only mode produced a plan, not a listing, so the adapter may need default mode with `--dangerously-skip-permissions` inside an isolated container.
- **Fixtures (A0.3):** captured under `internal/agent/testdata/` from the successful runs, with paths redacted:
  - `claude/stream.jsonl`: 11 tool-call events. Includes the user's session hooks and MCP setup.
  - `cursor/stream.jsonl`: 8 `tool_call` events, with `system`, `assistant`, `result` events.
  - `agy/stream.jsonl`: no `tool_call` events. Its activity is 32 `step_update` events, and its answer is in `assistant` events. The A3 adapter must read `step_update` for agy.
  - Redaction replaced the scratch worktree paths with `/WORKTREE`, the home directory with `/HOME`, and the encoded project path with `PROJECT`. Session IDs and tool names remain. A few path fragments are left inside streamed text deltas. A scan found no key-shaped strings.
  - All three files are valid JSON per line (checked with `jq`).
  - OpenCode and Codex have no fixture until their accounts work.
- **Container install (A0.5):** done as a survey (see above). Cursor and agy have no confirmed Linux arm64 build yet.

## Decisions to make

1. Confirm the Antigravity read-only approach. Default mode with `--dangerously-skip-permissions` inside the container is the candidate.
2. Pick a Codex model that the account supports, or leave Codex out of the default set until one is chosen.
3. Decide whether the adapters should pass `--bare` (Claude) and similar flags to avoid loading the user's global hooks and MCP servers into the logs.
