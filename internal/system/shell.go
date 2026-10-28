package system

import (
	"os"
	"path/filepath"
)

type ShellInfo struct {
	Name string
	Path string
}

func DetectShell() ShellInfo {
	path := os.Getenv("SHELL")
	if path == "" {
		return ShellInfo{}
	}

	return ShellInfo{
		Name: filepath.Base(path),
		Path: path,
	}
}
