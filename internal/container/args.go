package container

import (
	"maps"
	"slices"
)

func RunArgs(spec Spec) ([]string, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	args := []string{"--user", spec.User, "--read-only"}
	for _, path := range spec.Tmpfs {
		args = append(args, "--tmpfs", path)
	}
	args = append(args, "--network", spec.Network)
	if spec.Name != "" {
		args = append(args, "--name", spec.Name)
	}
	for _, key := range slices.Sorted(maps.Keys(spec.Labels)) {
		args = append(args, "--label", key+"="+spec.Labels[key])
	}
	if spec.Memory != "" {
		args = append(args, "--memory", spec.Memory)
	}
	for _, m := range spec.Mounts {
		args = append(args, "--mount", mountFlag(m))
	}
	args = append(args, spec.Image)
	args = append(args, spec.Cmd...)
	return args, nil
}

func mountFlag(m Mount) string {
	flag := "type=bind,source=" + m.Host + ",target=" + m.Guest
	if m.ReadOnly {
		flag += ",readonly"
	}
	return flag
}
