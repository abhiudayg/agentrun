package prototype_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	scriptName = "agentrun-proto"
	waitLimit  = 15 * time.Second
	pollEvery  = 50 * time.Millisecond
)

var createdPattern = regexp.MustCompile(`Run ([0-9a-f]{6}) created`)

type session struct {
	t     *testing.T
	data  string
	agent string
}

type outcome struct {
	stdout string
	stderr string
	exit   int
}

func newSession(t *testing.T, agentCommand string) *session {
	t.Helper()
	return &session{t: t, data: t.TempDir(), agent: agentCommand}
}

func (s *session) run(args ...string) outcome {
	s.t.Helper()
	script, err := filepath.Abs(scriptName)
	if err != nil {
		s.t.Fatal(err)
	}
	cmd := exec.Command(script, args...)
	cmd.Env = append(os.Environ(),
		"AGENTRUN_PROTO_DATA="+s.data,
		"AGENTRUN_AGENT_CMD="+s.agent,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err == nil {
		return outcome{stdout: stdout.String(), stderr: stderr.String()}
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return outcome{stdout: stdout.String(), stderr: stderr.String(), exit: exitErr.ExitCode()}
	}
	s.t.Fatalf("%s %v could not run: %v", scriptName, args, err)
	return outcome{}
}

func (s *session) submit(repo, prompt string, extra ...string) string {
	s.t.Helper()
	args := append([]string{"submit", prompt, "--repo", repo}, extra...)
	got := s.run(args...)
	if got.exit != 0 {
		s.t.Fatalf("submit failed (exit %d): %s", got.exit, got.stderr)
	}
	match := createdPattern.FindStringSubmatch(got.stdout)
	if match == nil {
		s.t.Fatalf("submit did not report a run ID:\n%s", got.stdout)
	}
	return match[1]
}

func (s *session) runDir(id string) string {
	return filepath.Join(s.data, "runs", id)
}

func (s *session) status(id string) string {
	raw, _ := os.ReadFile(filepath.Join(s.runDir(id), "status"))
	return strings.TrimSpace(string(raw))
}

func (s *session) waitFor(id string, want ...string) string {
	s.t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got := s.status(id)
		for _, w := range want {
			if got == w {
				return got
			}
		}
		if time.Now().After(deadline) {
			s.t.Fatalf("run %s status %q, want one of %v", id, got, want)
		}
		time.Sleep(pollEvery)
	}
}

func (s *session) worktree(id string) string {
	return filepath.Join(s.runDir(id), "worktree")
}

func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	gitIn(t, dir, "config", "user.name", "Test")
	gitIn(t, dir, "config", "user.email", "test@example.com")
	writeFile(t, filepath.Join(dir, "README.md"), "base\n")
	gitIn(t, dir, "add", "README.md")
	gitIn(t, dir, "commit", "-q", "-m", "base")
	return dir
}

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func assertStandardError(t *testing.T, got outcome) {
	t.Helper()
	if got.exit == 0 {
		t.Fatalf("expected failure, got success:\n%s", got.stdout)
	}
	for _, field := range []string{"error: ", "  cause: ", "  fix:   ", "  docs:  "} {
		if !strings.Contains(got.stderr, field) {
			t.Errorf("error is missing %q:\n%s", field, got.stderr)
		}
	}
}

func TestSubmitCreatesIsolatedWorktreeAndBranch(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "true")

	id := s.submit(repo, "Fix the flaky test")
	s.waitFor(id, "succeeded")

	if branches := gitIn(t, repo, "branch", "--list", "agentrun/"+id); !strings.Contains(branches, "agentrun/"+id) {
		t.Errorf("branch agentrun/%s was not created; branches: %q", id, branches)
	}
	if _, err := os.Stat(s.worktree(id)); err != nil {
		t.Errorf("worktree was not created: %v", err)
	}
	if got := readFile(t, filepath.Join(repo, "README.md")); got != "base\n" {
		t.Errorf("main working tree was modified: %q", got)
	}
}

func TestSuccessfulRunCommitsChangesAndWritesPatch(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "cat > agent-output.txt")

	id := s.submit(repo, "Write the greeting")
	s.waitFor(id, "succeeded")

	if got := readFile(t, filepath.Join(s.worktree(id), "agent-output.txt")); !strings.Contains(got, "Write the greeting") {
		t.Errorf("agent did not receive the task: %q", got)
	}
	if n := gitIn(t, repo, "rev-list", "--count", "main..agentrun/"+id); n != "1" {
		t.Errorf("branch has %s commits ahead of main, want 1", n)
	}
	patch := readFile(t, filepath.Join(s.runDir(id), "patch.diff"))
	if !strings.Contains(patch, "agent-output.txt") {
		t.Errorf("patch does not contain the change:\n%s", patch)
	}
	res := s.run("result", id)
	for _, want := range []string{"Outcome:  succeeded", "agent-output.txt"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("result missing %q:\n%s", want, res.stdout)
		}
	}
}

func TestFailedAgentIsMarkedFailedAndKeepsPartialChanges(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "echo partial > partial.txt; exit 3")

	id := s.submit(repo, "Migrate the config loader")
	s.waitFor(id, "failed")

	if got := strings.TrimSpace(readFile(t, filepath.Join(s.runDir(id), "exit_code"))); got != "3" {
		t.Errorf("exit code file = %q, want 3", got)
	}
	patch := readFile(t, filepath.Join(s.runDir(id), "patch.diff"))
	if !strings.Contains(patch, "partial.txt") {
		t.Errorf("partial changes were not preserved in the patch:\n%s", patch)
	}
}

func TestRunWithoutChangesSucceedsWithoutPatch(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "true")

	id := s.submit(repo, "Check the build")
	s.waitFor(id, "succeeded")

	if _, err := os.Stat(filepath.Join(s.runDir(id), "patch.diff")); !os.IsNotExist(err) {
		t.Errorf("patch.diff exists for a run with no changes (err=%v)", err)
	}
	if res := s.run("result", id); !strings.Contains(res.stdout, "Changed:  none") {
		t.Errorf("result does not report no changes:\n%s", res.stdout)
	}
}

func TestSubmitRefusesDirtyRepositoryUnlessIncluded(t *testing.T) {
	repo := newRepo(t)
	writeFile(t, filepath.Join(repo, "README.md"), "edited\n")
	s := newSession(t, "true")

	refused := s.run("submit", "task", "--repo", repo)
	assertStandardError(t, refused)
	if !strings.Contains(refused.stderr, "uncommitted changes") {
		t.Errorf("refusal does not name the dirty tree:\n%s", refused.stderr)
	}

	id := s.submit(repo, "task", "--include-dirty")
	s.waitFor(id, "succeeded", "failed")
}

func TestRepositoryPathWithSpacesIsSupported(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "my repo with spaces")
	if err := os.Mkdir(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q", "-b", "main")
	gitIn(t, repo, "config", "user.name", "Test")
	gitIn(t, repo, "config", "user.email", "test@example.com")
	writeFile(t, filepath.Join(repo, "README.md"), "base\n")
	gitIn(t, repo, "add", "README.md")
	gitIn(t, repo, "commit", "-q", "-m", "base")
	s := newSession(t, "cat > agent-output.txt")

	id := s.submit(repo, "Task in a spaced path")
	s.waitFor(id, "succeeded")

	if _, err := os.Stat(filepath.Join(s.worktree(id), "agent-output.txt")); err != nil {
		t.Errorf("agent output missing for a repository path with spaces: %v", err)
	}
}

func TestSubmitRejectsDirectoryThatIsNotARepository(t *testing.T) {
	s := newSession(t, "true")

	assertStandardError(t, s.run("submit", "task", "--repo", t.TempDir()))
}

func TestSubmitRejectsRepositoryWithoutCommits(t *testing.T) {
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	s := newSession(t, "true")

	got := s.run("submit", "task", "--repo", dir)
	assertStandardError(t, got)
	if !strings.Contains(got.stderr, "no commits") {
		t.Errorf("error does not name the missing commit:\n%s", got.stderr)
	}
}

func TestPromptIsTreatedAsDataNotShell(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "cat > agent-output.txt")
	prompt := "$(touch INJECTED) `touch INJECTED_BACKTICK`"

	id := s.submit(repo, prompt)
	s.waitFor(id, "succeeded")

	for _, path := range []string{
		filepath.Join(repo, "INJECTED"),
		filepath.Join(repo, "INJECTED_BACKTICK"),
		filepath.Join(s.worktree(id), "INJECTED"),
		filepath.Join(s.worktree(id), "INJECTED_BACKTICK"),
	} {
		if _, err := os.Stat(path); err == nil {
			t.Errorf("prompt was executed as shell: %s exists", path)
		}
	}
	if got := readFile(t, filepath.Join(s.worktree(id), "agent-output.txt")); !strings.Contains(got, "$(touch INJECTED)") {
		t.Errorf("prompt was altered before reaching the agent: %q", got)
	}
}

func TestRepositoryHooksDoNotRunOnHost(t *testing.T) {
	repo := newRepo(t)
	marker := filepath.Join(t.TempDir(), "hook-ran")
	hook := "#!/bin/sh\ntouch " + marker + "\n"
	for _, name := range []string{"post-checkout", "pre-commit"} {
		path := filepath.Join(repo, ".git", "hooks", name)
		writeFile(t, path, hook)
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	s := newSession(t, "cat > agent-output.txt")

	id := s.submit(repo, "Change the file")
	s.waitFor(id, "succeeded")

	if _, err := os.Stat(marker); err == nil {
		t.Errorf("a repository hook executed on the host during submit or commit")
	}
}

func TestCancelStopsAgentKeepsWorktreeAndIsRepeatSafe(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "sleep 60")

	id := s.submit(repo, "Long task")
	s.waitFor(id, "running")
	pid := s.waitForPID(id)

	got := s.run("cancel", id)
	if got.exit != 0 {
		t.Fatalf("cancel failed (exit %d): %s", got.exit, got.stderr)
	}
	if status := s.status(id); status != "cancelled" {
		t.Errorf("status = %q after cancel, want cancelled", status)
	}
	if !processGone(pid) {
		t.Errorf("agent process group %d is still alive after cancel", pid)
	}
	if _, err := os.Stat(s.worktree(id)); err != nil {
		t.Errorf("worktree was removed by cancel: %v", err)
	}

	again := s.run("cancel", id)
	if again.exit != 0 || !strings.Contains(again.stdout, "already cancelled") {
		t.Errorf("second cancel was not a safe no-op (exit %d): %s", again.exit, again.stdout)
	}
}

func (s *session) waitForPID(id string) int {
	s.t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		raw, err := os.ReadFile(filepath.Join(s.runDir(id), "agent.pid"))
		if err == nil {
			pid, convErr := strconv.Atoi(strings.TrimSpace(string(raw)))
			if convErr != nil {
				s.t.Fatalf("agent.pid is not a number: %q", raw)
			}
			return pid
		}
		if time.Now().After(deadline) {
			s.t.Fatalf("agent.pid never appeared for run %s", id)
		}
		time.Sleep(pollEvery)
	}
}

func processGone(pid int) bool {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(-pid, 0); errors.Is(err, syscall.ESRCH) {
			return true
		}
		time.Sleep(pollEvery)
	}
	return false
}

func TestRunIDsAreValidatedBeforeUse(t *testing.T) {
	s := newSession(t, "true")
	badIDs := []string{"../../etc", "ABCDEF", "abc", "abcdef0", "abc/de"}
	commands := []string{"logs", "result", "cancel"}

	for _, cmd := range commands {
		for _, id := range badIDs {
			t.Run(cmd+" "+id, func(t *testing.T) {
				got := s.run(cmd, id)
				assertStandardError(t, got)
				if !strings.Contains(got.stderr, "invalid run id") {
					t.Errorf("error does not name the invalid ID:\n%s", got.stderr)
				}
			})
		}
	}
}

func TestUnknownRunIsReportedWithStandardError(t *testing.T) {
	s := newSession(t, "true")

	got := s.run("logs", "000000")

	assertStandardError(t, got)
	if !strings.Contains(got.stderr, "run not found") {
		t.Errorf("error does not say the run is missing:\n%s", got.stderr)
	}
}

func TestUsageLogRecordsCommandNamesNotPrompts(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "true")

	id := s.submit(repo, "classified-note-7731")
	s.waitFor(id, "succeeded")
	s.run("result", id)

	usage := readFile(t, filepath.Join(s.data, "usage.log"))
	if !strings.Contains(usage, "submit") || !strings.Contains(usage, "result") {
		t.Errorf("usage log is missing command names:\n%s", usage)
	}
	if strings.Contains(usage, "classified-note") {
		t.Errorf("usage log records prompt text:\n%s", usage)
	}
}

func TestListShowsRunsWithStatus(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "true")

	first := s.submit(repo, "first task")
	s.waitFor(first, "succeeded")
	second := s.submit(repo, "second task")
	s.waitFor(second, "succeeded")

	got := s.run("list")
	for _, id := range []string{first, second} {
		if !strings.Contains(got.stdout, id) {
			t.Errorf("list is missing run %s:\n%s", id, got.stdout)
		}
	}
	if !strings.Contains(got.stdout, "succeeded") {
		t.Errorf("list does not show status:\n%s", got.stdout)
	}
}

func TestLogsShowAgentOutput(t *testing.T) {
	repo := newRepo(t)
	s := newSession(t, "echo hello-agent")

	id := s.submit(repo, "say hello")
	s.waitFor(id, "succeeded")

	if got := s.run("logs", id); !strings.Contains(got.stdout, "hello-agent") {
		t.Errorf("logs do not contain agent output:\n%s", got.stdout)
	}
}
