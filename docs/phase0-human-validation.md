# Phase 0 human validation: how to run it

An agent can do the teardown and the capability checks (see `docs/phase0-teardown.md` and `docs/phase0-capability-report.md`). The remaining Phase 0 gates need people: developers who use the product, interviews, and time on real machines. This document says how to run each one, what to record, and what counts as a pass. Thresholds come from the Lite PRD v1.0 §12 Phase 0 exit gate.

## Before any human test can start

1. **Prototype.** `prototype/agentrun-proto` implements the §6.2 command shape (`submit`, `list`, `logs`, `cancel`, `result`) and is covered by `go test ./prototype/`. A facilitator runs it on each participant's machine, or on a facilitator-controlled machine, as described in the README. It runs the agent without isolation, so each participant must agree to the agent and repository used.
2. **Decide the data rules.** The PRD forbids telemetry by default. Participant logs must record command names, timestamps, and outcomes only, never prompts, paths, or code. Store logs locally under the facilitator's control, with participants identified by a code (P01, P02, ...), not by name.
3. **Write the consent text.** It must say what is recorded, that participation can stop at any time, and how the data is deleted on request.

## Gate 1: Developer research (12–15 participants)

**Recruit.** Screen for developers who run a coding-agent CLI at least weekly. Exclude anyone on the team or working on this project. Record the screener answers so the sample can be described.

**Interview.** Run 30–45 minutes per participant over video or in person. Ask about:
- how many agent tasks they run per week, and whether they run them in parallel;
- the last time a task failed, was lost, or left a mess, and how they found out;
- what they do today to isolate tasks, check status, and collect results;
- what would make them stop using a new tool.

Take notes in a shared file with one row per participant. Quote the participant's own words for each pain point. Do not paraphrase sentiment into a stronger claim.

**Pass criteria (from the PRD exit gate):**
- Median participant runs at least 3 agent tasks per week.
- At least 3 participants describe a specific, recurring failure or friction in their current workflow that the prototype removed.

## Gate 2: Prototype use (5 participants, two weeks)

Give five participants the prototype and put the remaining participants in a comparison group using their usual workflow.

**Run.**
1. Install the prototype on each participant's machine, or have them run it on a facilitator-controlled machine.
2. Ask participants to use it for their real tasks for two weeks.
3. Check in at day 3 and day 10 by message. Ask only "what did you run, and what got in the way?"
4. At the end of week two, record which participants still run tasks through it.

**Record per participant, per week:** number of runs submitted, commands used (names only), and whether the participant stopped using it, with the stated reason.

**Pass criteria:**
- At least 2 of the 5 prototype users are still running tasks through it in week two.

## Gate 3: Quickstart usability (three clean Macs)

Use three Apple silicon Macs with different macOS versions. "Clean" means a new user account with no developer tools beyond Homebrew, and no Docker Desktop unless the test is about Docker.

**Procedure.**
1. Reset the account, or use a freshly imaged machine.
2. Give the participant the quickstart from the README only, with no help from the facilitator unless they are stuck for more than 2 minutes. Record every stuck point.
3. Start a timer at first command and stop it at the first successful run. Log each step's duration.
4. Repeat on the next machine with a different participant, then the third.

**Record:** time to first success per machine, every stuck point with its duration, and every error message the participant read aloud.

**Pass criteria:**
- Median time to first success is under 10 minutes for the prototype shape (PRD v1.0 §12).
- The top five usability issues are listed, with the step each one blocked.

## Gate 4: Maintainer interviews (8–12 people)

Use the same interview structure as Gate 1, aimed at maintainers and platform engineers. Ask how they verify agent-written changes today and what they would stop doing if a receipt existed. This feeds the ProofRail Phase 0 gate in `ProofRail PRD §11`; it is not a Lite gate.

## Go/no-go memo

When the gates above are done, the facilitator writes one memo with:
- the sample, by group, with dates;
- each gate's measured result against its threshold, pass or fail;
- the frozen v0.1 command surface;
- the top five usability issues.

The PRD says a failed gate narrows scope or stops the project. The memo records the decision; it does not make it. That decision belongs to the project owner.

## What the agent can do afterward

Once the participant logs and interview notes exist, an agent can:
- compute the metrics above from the logs and check them against the thresholds;
- draft the go/no-go memo from the recorded results;
- check the memo's numbers against the raw logs.

An agent cannot run the interviews, recruit participants, or stand in for a participant's usability session. Those results need a person in the loop.
