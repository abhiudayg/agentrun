# Decision: agent adapter survey (Phase A0)

Status: A0.1 to A0.5 done. OpenCode and Codex have no fixture because their accounts are blocked. Codex's model is unresolved.

Scope: what each agent CLI needs to run non-interactively, what structured output it emits, how it is kept read-only, which environment variables carry its credentials, and whether it runs in a Linux container. Results come from `--help` output on this machine, one read-only listing run per agent through `agentrun-proto`, vendor docs, and runs inside Apple `container` images.

## A0.1: versions and invocation

| Agent | Binary and version here | Non-interactive command | Structured output | Read-only option |
|---|---|---|---|---|
| claude | `claude` 2.1.308 | `claude -p "<prompt>"` | `--output-format json` or `stream-json` (`stream-json` needs `--verbose`) | `--permission-mode plan` |
| opencode | `opencode` v2.1.2 | `opencode run "<message>"` | `--format json` | none in help; `--auto` only auto-approves |
| agy (Antigravity) | `agy` 1.4.3 | `agy -p "<prompt>"` | `--output-format stream-json` (also `json`, `text`) | `--mode plan` (produced a plan artifact, not a listing) |
| codex | codex-cli 0.141.0 | `codex exec "<prompt>"` | `--json` (JSONL events) | `-s read-only` |
| cursor | `agent` 2026.10.19-6bbf2c4 | `agent -p "<prompt>"` | `--output-format stream-json` (also `json`, `text`) | `--mode plan` or `ask` |

Notes:
- The Cursor binary is `agent`. The `cursor-agent` name points at the same install.
- Cursor's help says `-p` has access to all tools, including write and shell. Read-only depends on `--mode plan` or `ask`.
- Cursor needs `--trust` in a new workspace, or it stops at "Workspace Trust Required".

## A0.2: prototype runs

Each run listed the files in a scratch repo with two files, `README.md` and `calc.py`, and was told not to modify anything. The prototype passes the task on stdin and runs the agent command with `sh -c`.

| Agent | Run | Result | Log size | Notes |
|---|---|---|---|---|
| claude | `bc23b4` | succeeded | ~27 KB | Loaded the user's full config: session hooks, MCP servers, about 80 tools. Log is noisy. |
| cursor | `2d08f4` | succeeded | ~28 KB | `agent -p --output-format stream-json --mode plan --trust "$(cat)"`. Also loaded the user's agent-brain routing. |
| agy | `3feee9` | succeeded | ~10 KB | Needs `--add-dir "$PWD"`. Without it, agy listed a different directory's files. Needed `--dangerously-skip-permissions` in default mode. |
| opencode | `faed0b` | failed | 0.8 KB | `provider.free-limit`, HTTP 429, retry after about 18 hours. Blocked until a provider is connected. |
| codex | `5f819c` | failed | 0.3 KB | `The 'gpt-5.5' model is not supported when using Codex with a ChatGPT account.` Blocked on model choice. |

Failed attempts that were fixed and are not counted as results:
- `ba454c` (agy): `-p` took `--output-format` as its prompt. Fixed by putting the prompt directly after `-p`.
- `547c2e` (cursor): stopped at "Workspace Trust Required". Fixed with `--trust`.
- `ec1a1f` (agy): succeeded but listed a directory outside the run. Superseded by `3feee9`.
- `15949f` (codex): hung on an invalid refresh token. Cancelled. Re-login fixed the token; the later run failed on the model.

## A0.3: fixtures

Captured from the successful runs and stored under `internal/agent/testdata/`:

| File | Source run | What it contains |
|---|---|---|
| `claude/stream.jsonl` | `bc23b4` | 11 tool-call events. Includes the user's session hooks and MCP setup. |
| `cursor/stream.jsonl` | `2d08f4` | 8 `tool_call` events, plus `system`, `assistant`, and `result` events. |
| `agy/stream.jsonl` | `3feee9` | No `tool_call` events. 32 `step_update` events, and the answer in `assistant` events. The A3 adapter must read `step_update` for agy. |

Redaction:
- Scratch worktree paths were replaced with `/WORKTREE`, the home directory with `/HOME`, and the encoded project path with `PROJECT`.
- Session IDs and tool names remain.
- A few path fragments are left inside streamed text deltas. A scan found no key-shaped strings.
- Each file is valid JSON per line (checked with `jq`).

No fixture exists for OpenCode or Codex. Both need working accounts first.

## A0.4: auth environment variables

Confirmed by `--help` or vendor docs:

| Agent | Variable | Source | Behavior |
|---|---|---|---|
| claude | `ANTHROPIC_API_KEY` | `--help` (`--bare` mode) | Bare mode accepts only this key. OAuth login works without it, in normal mode. |
| cursor | `CURSOR_API_KEY` | `--help` (`--api-key`) | `CURSOR_API_ENDPOINT` is optional. `agent login` is the interactive route. |
| codex | `OPENAI_API_KEY` | `codex login --help` (`--with-api-key`) | ChatGPT login also works. |
| agy | `GEMINI_API_KEY` | [Antigravity install and auth docs](https://antigravity.google/docs/install?tab=cli) | Required only when `"modelProvider": "gemini"` is set in `~/.gemini/antigravity-cli/settings.json`. Without that setting, the variable has no effect. |
| agy | `AGY_ADC_AUTH=true` | same docs | Uses Application Default Credentials from `gcloud auth application-default login`. Alternative to the key. |
| agy | `GOOGLE_GEMINI_BASE_URL` | same docs | Optional custom Gemini-compatible endpoint. |

Corrections:
- `GOOGLE_API_KEY` appears in the agy binary, but the docs say a key set through `GOOGLE_API_KEY` or a `.env` file has no effect. It is not a supported variable.
- For headless agy, the settings file must exist with `modelProvider` set. The adapter or the image must write it. An environment variable alone is not enough.

OpenCode's binary names many provider keys, such as `ANTHROPIC_API_KEY` and `GEMINI_API_KEY`. Which one is needed depends on the configured provider, and this is not verified.

No variable values were printed or stored.

## A0.5: Linux container install

### Container check

Run with Apple `container` 1.6.0, against Linux arm64 images. Each agent was installed inside a container and run with `--version`. No prompts were sent, so no quota was spent.

| Agent | Debian (`node:22-bookworm-slim`, glibc) | Alpine (`node:22-alpine`, musl) | Install method |
|---|---|---|---|
| claude | 2.1.308 ✓ | 2.1.308 ✓ | `npm i -g @anthropic-ai/claude-code@2.1.308` |
| codex | 0.163.0 ✓ | 0.163.0 ✓ | `npm i -g @openai/codex@0.163.0` |
| opencode | 1.19.3 ✓ | 1.19.3 ✓ | `npm i -g opencode-ai@1.19.3` |
| cursor (`agent`) | 2026.10.19-6bbf2c4 ✓ | install ✓, **run fails**: `required file not found` | `curl https://cursor.com/install -fsS \| bash` |
| agy | 1.4.5 ✓ | 1.4.5 ✓ | `curl -fsSL https://antigravity.google/cli/install.sh \| bash` |

What the check shows:
- Use a glibc (Debian-based) image. Cursor's bundled Node binary doesn't run on musl.
- The npm packages are pinned and ran as expected.
- Cursor and agy installers fetch the latest build, with no version pin. The Cursor build matches this machine. agy is 1.4.5 in the container and 1.4.3 on this machine, so the agy build differs from the one that produced the fixtures.
- Node 22.23.4 was in the Debian image. Node is needed only to install the npm packages, not to run the native builds.

### Earlier survey (npm registry and vendor docs)

Still accurate for the packages that were not run:
- claude, codex, and opencode ship per-platform optional packages for Linux x64 and arm64, and for musl.
- Registry latest is codex 0.163.0 and opencode 1.19.3. Both are newer than the local versions, codex 0.141.0 and opencode v2.1.2.

## Log hygiene (decisions)

1. **Antigravity read-only approach: accepted.** agy runs in default mode with `--dangerously-skip-permissions`, inside the container only. Its `--mode plan` produced a plan artifact instead of an answer, so it is not used. The adapter always passes `--add-dir` with the workspace.
2. **Keep hooks and MCP servers out of agent logs: accepted, with a Claude-specific flag set.**
   - `--bare` is not usable. The probe with `claude --bare` returned `Not logged in · Please run /login`. Bare mode accepts only `ANTHROPIC_API_KEY`, and the current login is OAuth.
   - Claude adapter: `claude -p --output-format stream-json --verbose --setting-sources project --strict-mcp-config --mcp-config '{"mcpServers":{}}'`. The probe with these flags returned `OK`, 7 log lines, and no `hook_started` events. Login still works.
   - Codex: `--ignore-user-config` exists in the help text. Not tested, because the account's model is blocked.
   - Cursor (`agent`), agy, and OpenCode: no flag in their help skips user hooks or MCP servers. Cursor has `--approve-mcps`, which approves servers and does not skip them. These adapters need an isolated config directory, or a container without the host's config. Open for A3.
3. **Codex model: open.** Codex stays out of the default set until an account-supported model is chosen.

## Blockers and open items

- **OpenCode:** needs a connected provider or a subscription before any run, and so a fixture.
- **Codex:** needs a model name the ChatGPT account supports. The default `gpt-5.5` is rejected. Needs a fixture too.
- **Antigravity:** the adapter must always pass `--add-dir` with the workspace, and must write `~/.gemini/antigravity-cli/settings.json` with `modelProvider` set, plus `GEMINI_API_KEY` or ADC, for headless runs.
- **Cursor on Alpine:** not supported. Use the Debian image.
- **Unpinned installers:** Cursor and agy installers fetch the latest build. Decide how to pin them in A2, or accept the drift and record the version in each run.
- **Cursor and agy in the container:** installed and run, but not yet run against a real prompt inside a container. That belongs to A2.
