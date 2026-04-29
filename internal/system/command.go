package system

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
)

type CommandStatus struct {
	Name      string
	Installed bool
	Path      string
}

func DetectCommand(name string) CommandStatus {
	path, err := exec.LookPath(name)
	if err != nil {
		return CommandStatus{Name: name}
	}

	return CommandStatus{
		Name:      name,
		Installed: true,
		Path:      path,
	}
}

func RunCommand(ctx context.Context, dir string, name string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		if stdout.Len() > 0 {
			return strings.TrimSpace(stdout.String()), err
		}
		return strings.TrimSpace(stderr.String()), err
	}

	return strings.TrimSpace(stdout.String()), nil
}
