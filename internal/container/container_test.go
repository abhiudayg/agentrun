package container_test

import (
	"context"
	"errors"
	"testing"

	"agentrun/internal/container"
)

type fakeBackend struct {
	gotSpec container.Spec
	result  container.Result
	err     error
}

func (f *fakeBackend) Run(ctx context.Context, spec container.Spec) (container.Result, error) {
	f.gotSpec = spec
	return f.result, f.err
}

var _ container.Backend = (*fakeBackend)(nil)

func validSpec() container.Spec {
	return container.Spec{
		Image:    "ghcr.io/example/agent:1.2.3",
		User:     "1000:1000",
		ReadOnly: true,
		Tmpfs:    []string{"/tmp"},
		Network:  "agentrun-internal",
		Cmd:      []string{"sh", "-c", "echo hi"},
	}
}

func TestBackendRunReturnsResultAndError(t *testing.T) {
	boom := errors.New("runtime unavailable")
	tests := []struct {
		name     string
		result   container.Result
		err      error
		wantCode int
		wantErr  error
	}{
		{name: "success", result: container.Result{ExitCode: 0}, wantCode: 0},
		{name: "nonzero exit", result: container.Result{ExitCode: 7}, wantCode: 7},
		{name: "backend error", result: container.Result{ExitCode: 1}, err: boom, wantCode: 1, wantErr: boom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeBackend{result: tt.result, err: tt.err}
			spec := validSpec()

			var backend container.Backend = fake
			got, err := backend.Run(context.Background(), spec)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Run() error = %v, want %v", err, tt.wantErr)
			}
			if got.ExitCode != tt.wantCode {
				t.Errorf("Run() exit code = %d, want %d", got.ExitCode, tt.wantCode)
			}
			if fake.gotSpec.Image != spec.Image {
				t.Errorf("Run() spec image = %q, want %q", fake.gotSpec.Image, spec.Image)
			}
		})
	}
}
