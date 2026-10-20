package errors

import "fmt"

// Error represents a structured error complying with the AgentRun error standard.
type Error struct {
	What  string `json:"what"`
	Cause string `json:"cause"`
	Fix   string `json:"fix"`
	Docs  string `json:"docs"`
}

// Error returns the formatted error string in the 4-line standard format:
// error: <what>
//
//	cause: <cause>
//	fix:   <fix>
//	docs:  <docs>
func (e Error) Error() string {
	return fmt.Sprintf("error: %s\n  cause: %s\n  fix:   %s\n  docs:  %s", e.What, e.Cause, e.Fix, e.Docs)
}

// New constructs a new Error with immutable semantics.
func New(what, cause, fix, docs string) Error {
	return Error{
		What:  what,
		Cause: cause,
		Fix:   fix,
		Docs:  docs,
	}
}

// WithWhat returns an updated copy of the Error with a new What message.
func (e Error) WithWhat(what string) Error {
	return Error{
		What:  what,
		Cause: e.Cause,
		Fix:   e.Fix,
		Docs:  e.Docs,
	}
}

// WithCause returns an updated copy of the Error with a new Cause message.
func (e Error) WithCause(cause string) Error {
	return Error{
		What:  e.What,
		Cause: cause,
		Fix:   e.Fix,
		Docs:  e.Docs,
	}
}

// WithFix returns an updated copy of the Error with a new Fix message.
func (e Error) WithFix(fix string) Error {
	return Error{
		What:  e.What,
		Cause: e.Cause,
		Fix:   fix,
		Docs:  e.Docs,
	}
}

// WithDocs returns an updated copy of the Error with a new Docs link.
func (e Error) WithDocs(docs string) Error {
	return Error{
		What:  e.What,
		Cause: e.Cause,
		Fix:   e.Fix,
		Docs:  docs,
	}
}

// ErrorCode is a string identifier for catalog entries.
type ErrorCode string

const (
	ErrBackendMissingCapability ErrorCode = "ERR_BACKEND_MISSING_CAPABILITY"
	ErrDockerNotReachable       ErrorCode = "ERR_DOCKER_NOT_REACHABLE"
	ErrAppleContainerNotRunning ErrorCode = "ERR_APPLE_CONTAINER_NOT_RUNNING"
	ErrDirtyWorkingTree         ErrorCode = "ERR_DIRTY_WORKING_TREE"
	ErrUnmetPrerequisites       ErrorCode = "ERR_UNMET_PREREQUISITES"
	ErrRunNotFound              ErrorCode = "ERR_RUN_NOT_FOUND"
	ErrActionPolicyViolation    ErrorCode = "ERR_ACTION_POLICY_VIOLATION"
	ErrProtectedPathModified    ErrorCode = "ERR_PROTECTED_PATH_MODIFIED"
	ErrInvalidMode              ErrorCode = "ERR_INVALID_MODE"
	ErrFullModeRequired         ErrorCode = "ERR_FULL_MODE_REQUIRED"
)

// Catalog returns the curated repository of all system error definitions.
func Catalog() map[ErrorCode]Error {
	return map[ErrorCode]Error{
		ErrBackendMissingCapability: New(
			"cannot start run: required capability is not enforceable on this backend",
			"the backend has not passed capability verification for one or more requested features",
			"run `agentrun backend test` for details, or adjust policy to remove the requirement",
			"https://agentrun.dev/docs/troubleshooting#capabilities",
		),
		ErrDockerNotReachable: New(
			"cannot start run: Docker daemon is not reachable",
			"the Docker socket does not exist or is not accessible",
			"start Docker Desktop, or add your user to the docker group, then run `agentrun doctor`",
			"https://agentrun.dev/docs/troubleshooting#docker",
		),
		ErrAppleContainerNotRunning: New(
			"cannot start run: Apple container service is not running",
			"the container background daemon is stopped or not initialized",
			"start the container service or run `agentrun doctor` for instructions",
			"https://agentrun.dev/docs/troubleshooting#apple-container",
		),
		ErrDirtyWorkingTree: New(
			"cannot start run: working tree contains uncommitted changes",
			"agentrun requires a clean base commit to guarantee isolated worktree branches",
			"commit or stash your changes, or run with `--include-dirty`",
			"https://agentrun.dev/docs/troubleshooting#working-tree",
		),
		ErrUnmetPrerequisites: New(
			"preflight checks failed",
			"one or more required tools or credentials are missing",
			"run `agentrun doctor` and resolve the reported errors",
			"https://agentrun.dev/docs/troubleshooting#doctor",
		),
		ErrRunNotFound: New(
			"run not found",
			"the specified run ID does not exist in the run store",
			"check the run ID with `agentrun list`",
			"https://agentrun.dev/docs/troubleshooting#run-not-found",
		),
		ErrActionPolicyViolation: New(
			"external action refused by policy",
			"the requested action is not permitted by repository contract or organization policy",
			"inspect allowed actions in `.proofrail/contract.yaml` or contact your operator",
			"https://agentrun.dev/docs/troubleshooting#action-policy",
		),
		ErrProtectedPathModified: New(
			"change rejected: protected path modified",
			"the agent modified files restricted by the verification contract",
			"ensure the agent only edits allowed workspace files",
			"https://agentrun.dev/docs/troubleshooting#protected-paths",
		),
		ErrInvalidMode: New(
			"invalid mode",
			"the mode value is not one of lite or full",
			"set `mode: lite` or `mode: full` in configuration, or pass `--mode lite` or `--mode full`",
			"https://agentrun.dev/docs/troubleshooting#modes",
		),
		ErrFullModeRequired: New(
			"command unavailable in lite mode",
			"this command is only available in Full mode, and the current mode is not full",
			"rerun with `--mode full`, or set `mode: full` in organization configuration",
			"https://agentrun.dev/docs/troubleshooting#modes",
		),
	}
}
