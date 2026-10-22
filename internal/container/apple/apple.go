package apple

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"agentrun/internal/container"
)

var _ container.Backend = (*Apple)(nil)

type Apple struct {
	Binary string
	Output io.Writer
}

func (a *Apple) Run(ctx context.Context, spec container.Spec) (container.Result, error) {
	args, err := container.RunArgs(spec)
	if err != nil {
		return container.Result{}, fmt.Errorf("build container run args: %w", err)
	}

	cmd := exec.CommandContext(ctx, a.Binary, append([]string{"run"}, args...)...)
	cmd.Stdout = a.Output
	cmd.Stderr = a.Output

	err = cmd.Run()
	if err == nil {
		return container.Result{}, nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return container.Result{}, fmt.Errorf("run container: %w", ctxErr)
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return container.Result{ExitCode: exitErr.ExitCode()}, nil
	}
	return container.Result{}, fmt.Errorf("run container: %w", err)
}
