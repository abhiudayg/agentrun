package container

import (
	"context"
	"errors"
)

var (
	ErrRootUser      = errors.New("container user must not be root")
	ErrEmptyUser     = errors.New("container user is empty")
	ErrWritableRoot  = errors.New("container root filesystem must be read-only")
	ErrMissingTmpfs  = errors.New("container must mount tmpfs at /tmp")
	ErrNoNetwork     = errors.New("container network is empty")
	ErrGitDirMounted = errors.New("container must not mount a .git directory")
	ErrLatestImage   = errors.New("container image must not use the latest tag or no tag")
	ErrEmptyImage    = errors.New("container image is empty")
	ErrEmptyCommand  = errors.New("container command is empty")
)

type Mount struct {
	Host     string
	Guest    string
	ReadOnly bool
}

type Spec struct {
	Image    string
	User     string
	ReadOnly bool
	Tmpfs    []string
	Mounts   []Mount
	Network  string
	Cmd      []string
}

type Result struct {
	ExitCode int
}

type Backend interface {
	Run(ctx context.Context, spec Spec) (Result, error)
}
