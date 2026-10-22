package containertest_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"agentrun/internal/container"
	"agentrun/internal/container/containertest"
)

var _ container.Backend = (*containertest.Fake)(nil)

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

func TestRunReturnsScriptedResultAndRecordsSpec(t *testing.T) {
	spec := validSpec()
	f := &containertest.Fake{Results: []container.Result{{ExitCode: 0}}}

	got, err := f.Run(context.Background(), spec)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if want := (container.Result{ExitCode: 0}); got != want {
		t.Errorf("Run() result = %+v, want %+v", got, want)
	}
	if len(f.Specs) != 1 {
		t.Fatalf("recorded specs = %d, want 1", len(f.Specs))
	}
	if !reflect.DeepEqual(f.Specs[0], spec) {
		t.Errorf("recorded spec = %+v, want %+v", f.Specs[0], spec)
	}
}

func TestRunReturnsNonZeroExitAsResultWithNilError(t *testing.T) {
	f := &containertest.Fake{Results: []container.Result{{ExitCode: 3}}}

	got, err := f.Run(context.Background(), validSpec())
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if want := (container.Result{ExitCode: 3}); got != want {
		t.Errorf("Run() result = %+v, want %+v", got, want)
	}
}

func TestRunStartErrReturnedAndSpecRecorded(t *testing.T) {
	boom := errors.New("runtime unavailable")
	spec := validSpec()
	f := &containertest.Fake{
		Results:  []container.Result{{ExitCode: 0}},
		StartErr: boom,
	}

	got, err := f.Run(context.Background(), spec)
	if !errors.Is(err, boom) {
		t.Fatalf("Run() error = %v, want %v", err, boom)
	}
	if want := (container.Result{}); got != want {
		t.Errorf("Run() result = %+v, want %+v", got, want)
	}
	if len(f.Specs) != 1 {
		t.Fatalf("recorded specs = %d, want 1", len(f.Specs))
	}
	if !reflect.DeepEqual(f.Specs[0], spec) {
		t.Errorf("recorded spec = %+v, want %+v", f.Specs[0], spec)
	}
	if len(f.Results) != 1 {
		t.Errorf("remaining results = %d, want 1 (StartErr must not consume results)", len(f.Results))
	}
}

func TestRunInvalidSpecReturnsValidationErrorAndIsNotRecorded(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*container.Spec)
		wantErr error
	}{
		{
			name:    "root user",
			mutate:  func(s *container.Spec) { s.User = "0:0" },
			wantErr: container.ErrRootUser,
		},
		{
			name:    "empty image",
			mutate:  func(s *container.Spec) { s.Image = "" },
			wantErr: container.ErrEmptyImage,
		},
		{
			name:    "writable root",
			mutate:  func(s *container.Spec) { s.ReadOnly = false },
			wantErr: container.ErrWritableRoot,
		},
		{
			name:    "empty command",
			mutate:  func(s *container.Spec) { s.Cmd = nil },
			wantErr: container.ErrEmptyCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := validSpec()
			tt.mutate(&spec)
			f := &containertest.Fake{Results: []container.Result{{ExitCode: 0}}}

			got, err := f.Run(context.Background(), spec)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Run() error = %v, want %v", err, tt.wantErr)
			}
			if want := (container.Result{}); got != want {
				t.Errorf("Run() result = %+v, want %+v", got, want)
			}
			if len(f.Specs) != 0 {
				t.Errorf("recorded specs = %d, want 0", len(f.Specs))
			}
			if len(f.Results) != 1 {
				t.Errorf("remaining results = %d, want 1", len(f.Results))
			}
		})
	}
}

func TestRunCancelledContextReturnsCanceledAndIsNotRecorded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f := &containertest.Fake{Results: []container.Result{{ExitCode: 0}}}

	got, err := f.Run(ctx, validSpec())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}
	if want := (container.Result{}); got != want {
		t.Errorf("Run() result = %+v, want %+v", got, want)
	}
	if len(f.Specs) != 0 {
		t.Errorf("recorded specs = %d, want 0", len(f.Specs))
	}
	if len(f.Results) != 1 {
		t.Errorf("remaining results = %d, want 1", len(f.Results))
	}
}

func TestRunPopsResultsInOrder(t *testing.T) {
	tests := []struct {
		name    string
		results []container.Result
		want    []container.Result
	}{
		{
			name:    "scripted results come back in order",
			results: []container.Result{{ExitCode: 1}, {ExitCode: 2}, {ExitCode: 0}},
			want:    []container.Result{{ExitCode: 1}, {ExitCode: 2}, {ExitCode: 0}},
		},
		{
			name:    "exhausted results return zero value with nil error",
			results: []container.Result{{ExitCode: 7}},
			want:    []container.Result{{ExitCode: 7}, {}, {}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &containertest.Fake{Results: tt.results}
			var got []container.Result
			for i := range tt.want {
				r, err := f.Run(context.Background(), validSpec())
				if err != nil {
					t.Fatalf("Run() call %d error = %v, want nil", i+1, err)
				}
				got = append(got, r)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() results = %+v, want %+v", got, tt.want)
			}
			if len(f.Specs) != len(tt.want) {
				t.Errorf("recorded specs = %d, want %d", len(f.Specs), len(tt.want))
			}
		})
	}
}
