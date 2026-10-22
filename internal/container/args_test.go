package container_test

import (
	"errors"
	"slices"
	"testing"

	"agentrun/internal/container"
)

func TestRunArgs(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*container.Spec)
		want    []string
		wantErr error
	}{
		{
			name:   "minimal spec",
			mutate: func(*container.Spec) {},
			want: []string{
				"--user", "1000:1000",
				"--read-only",
				"--tmpfs", "/tmp",
				"--network", "agentrun-internal",
				"ghcr.io/example/agent:1.2.3",
				"sh", "-c", "echo hi",
			},
		},
		{
			name: "multiple tmpfs and mounts keep order",
			mutate: func(s *container.Spec) {
				s.Tmpfs = []string{"/tmp", "/run"}
				s.Mounts = []container.Mount{
					{Host: "/work/ws", Guest: "/work"},
					{Host: "/work/ro", Guest: "/ro", ReadOnly: true},
				}
			},
			want: []string{
				"--user", "1000:1000",
				"--read-only",
				"--tmpfs", "/tmp",
				"--tmpfs", "/run",
				"--network", "agentrun-internal",
				"--mount", "type=bind,source=/work/ws,target=/work",
				"--mount", "type=bind,source=/work/ro,target=/ro,readonly",
				"ghcr.io/example/agent:1.2.3",
				"sh", "-c", "echo hi",
			},
		},
		{
			name: "command elements with spaces stay single arguments",
			mutate: func(s *container.Spec) {
				s.Cmd = []string{"python", "-c", "print('a b')", "two words"}
			},
			want: []string{
				"--user", "1000:1000",
				"--read-only",
				"--tmpfs", "/tmp",
				"--network", "agentrun-internal",
				"ghcr.io/example/agent:1.2.3",
				"python", "-c", "print('a b')", "two words",
			},
		},
		{
			name: "invalid spec returns validation error",
			mutate: func(s *container.Spec) {
				s.ReadOnly = false
			},
			wantErr: container.ErrWritableRoot,
		},
		{
			name: "git dir mount is rejected",
			mutate: func(s *container.Spec) {
				s.Mounts = []container.Mount{{Host: "/repo/.git", Guest: "/git"}}
			},
			wantErr: container.ErrGitDirMounted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := validSpec()
			tt.mutate(&spec)

			got, err := container.RunArgs(spec)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("RunArgs() error = %v, want %v", err, tt.wantErr)
				}
				if got != nil {
					t.Errorf("RunArgs() args = %q, want nil on error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("RunArgs() error = %v, want nil", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("RunArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}
