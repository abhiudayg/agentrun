# Verification steps

Run these steps before a push, after an agent run is merged, and when a change touches a container boundary. Each step names the command, so a result can be recorded and reproduced.

## 1. Before a push

Run from the repository root on the branch being pushed:

```sh
agent-brain verify -- sh -c 'gofmt -l . && go vet ./... && go test -race -count=1 ./...'
```

- `gofmt -l .` must print nothing.
- `go vet ./...` must pass.
- `go test -race -count=1 ./...` must pass.
- Record the verify id. A push without a recorded passing run is not verified.

When `internal/container/` or `test/capability/` changed, also run the capability suite. It needs the Apple `container` service running:

```sh
go test -tags capability -count=1 -v -timeout 10m ./test/capability/
```

## 2. After an agent run

Before merging a run branch `agentrun/<id>`, check each of these:

1. The result is `succeeded`: `agentrun-proto list` shows it.
2. The diff touches only the files the task named: `git diff --stat self-build agentrun/<id>`.
3. The new code has no comments: `grep -n "//" <changed files>` returns nothing new.
4. The checks pass in the run's worktree, using the same command as step 1.
5. New regression cases fail on the old code. Check this by running the new tests against the previous commit.
6. Record the token use and cost from the run's `result` event (`num_turns`, `total_cost_usd`, `usage`).

## 3. After a push

Confirm the remote matches the local branch:

```sh
git fetch origin
git rev-parse self-build origin/self-build
```

Both commands must print the same hash. If the push was a force push, the remote tip must equal the lease value that was used.

## 4. Known gaps

- A run's "tests failed before the fix" claim is not checked automatically. Step 2.5 is a manual check until it is scripted.
- The prototype runs agents without isolation. Step 2 is the only review before code is merged.
