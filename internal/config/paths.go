package config

import (
	"os"
	"path/filepath"
)

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func GlobalDir() string {
	home := HomeDir()
	if home == "" {
		return ".odin"
	}
	return filepath.Join(home, ".odin")
}

func GlobalConfigPath() string {
	return filepath.Join(GlobalDir(), "config.toml")
}

func ProjectDir(cwd string) string {
	return filepath.Join(cwd, ".odin")
}

func ProjectConfigPath(cwd string) string {
	return filepath.Join(ProjectDir(cwd), "config.toml")
}

func ProjectRulesPath(cwd string) string {
	return filepath.Join(ProjectDir(cwd), "rules.md")
}

func ProjectAgentsPath(cwd string) string {
	return filepath.Join(cwd, "AGENTS.md")
}
