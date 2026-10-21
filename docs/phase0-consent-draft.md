# Phase 0 participant consent and setup: DRAFT

Status: draft for review. Not approved for use. Review the wording with whoever owns the study before any participant sees it.

## Part 1: Consent text

**What this study is.** We are testing a prototype command-line tool that runs a coding agent on a task in a separate git worktree and records the result. We want to learn whether the tool fits how you already work with coding agents. We are not testing you.

**What the tool does on your machine.**
- It creates a git worktree and a branch for each task you submit, named `agentrun/<id>`. It does not change your active working tree.
- It runs the coding agent you choose with your own user permissions. The agent can read and write anything your user account can. The tool does not sandbox it.
- It stores each task locally under `~/.local/share/agentrun-proto/runs/<id>/`. That folder contains the task text you submitted, the agent's log, the status, and a patch of any changes. These files stay on your machine unless you send them to us.

**What we record for the study.**
- The commands you run (submit, list, logs, result, cancel), with a run ID and a timestamp.
- Whether each run succeeded, failed, or was cancelled.
- Your answers to the interview and check-in questions.

We do not record the text of your tasks, your code, file paths, or the agent's output in the study log. Study records use a participant code (for example P01), not your name.

**Your choices.**
- Participation is voluntary. You can stop at any time, and you do not have to give a reason.
- You can ask us to delete your study records and your local run folders at any time. We will confirm when that is done.
- Declining or stopping has no effect on anything else you do with us.

**Risks.** The agent you run may change files, run commands, or send your task text to its model provider, as it would without this tool. Please use it only on repositories and tasks you have agreed to use in this study. Do not put secrets or credentials in task text.

**Contact.** [name and address for questions, to be filled in by the study owner]

**Consent.** I have read this, I understand the tool runs without isolation, and I agree to take part.

Participant code: ______   Date: ______   Signature or typed name confirming consent: ______

## Part 2: Participant setup note

Read this before you install anything.

1. Requirements: macOS, bash, git 2.30 or later (tested on 2.54), and the Claude CLI (`claude --version` must work).
2. Install: [facilitator to add the exact install command and the release or commit to use].
3. Choose a repository you have agreed to use in this study. It must have at least one commit, and its working tree must be clean. Otherwise `submit` refuses to start, and it says why.
4. Submit a task:
   ```
   agentrun-proto submit "describe the change you want" --repo /path/to/repo
   ```
5. Check progress and results:
   ```
   agentrun-proto list
   agentrun-proto logs <id>
   agentrun-proto result <id>
   ```
6. Cancel a run you no longer want: `agentrun-proto cancel <id>`. Cancel only after the run shows `running` with an agent PID. Check its status again afterwards. A cancel that arrives during startup can leave the run running.
7. Review any change before you use it. Changes are on the branch `agentrun/<id>` in the repository you chose, and the patch is in the run folder. Nothing is merged for you.
8. Your check-in question is: "What did you run, and what got in the way?" Answer by message. Do not include task text or code.

## Open items before this can be used

- The consent text needs review by the study owner, including the contact details and any requirements for the study's ethics approval.
- The install command in step 2 is not written. The clean-machine install check (runbook Gate 3) has not been run.
- The default data location and the cancel race are known limits. They are disclosed in the README, and the consent text should point to them.
