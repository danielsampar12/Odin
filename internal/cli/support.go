package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func workingDir() (string, error) {
	return os.Getwd()
}

func ensureDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return false, fmt.Errorf("%s exists and is not a directory", path)
		}
		return false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, os.MkdirAll(path, 0o755)
}

func writeFileIfMissing(path, content string, perm os.FileMode) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	return true, os.WriteFile(path, []byte(content), perm)
}

func installedLabel(installed bool) string {
	if installed {
		return "installed"
	}
	return "not installed"
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func displayPath(path string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}

	cwd, err := os.Getwd()
	if err == nil && cwd != "" {
		if rel, relErr := filepath.Rel(cwd, path); relErr == nil && rel != ".." && !strings.HasPrefix(rel, "../") {
			return rel
		}
	}

	return path
}

func fileStatusLabel(path string, exists bool) string {
	if exists {
		return fmt.Sprintf("found (%s)", displayPath(path))
	}
	return fmt.Sprintf("missing (%s)", displayPath(path))
}

func shellDisplay(path, name string) string {
	if name == "" && path == "" {
		return "unknown"
	}
	if name == "" {
		return path
	}
	if path == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, path)
}

func detectProfile(cwd string, inGitRepo bool) string {
	if inGitRepo || hasProjectMarkers(cwd) {
		return "developer"
	}
	return "beginner"
}

func hasProjectMarkers(dir string) bool {
	markers := []string{
		"go.mod",
		"package.json",
		"pyproject.toml",
		"Cargo.toml",
		".git",
		"Makefile",
	}

	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}

	return false
}

func printFileAction(writer io.Writer, path string, created bool) {
	if created {
		fmt.Fprintf(writer, "- Created %s\n", displayPath(path))
		return
	}

	fmt.Fprintf(writer, "- Kept existing %s\n", displayPath(path))
}

func confirmAction(reader io.Reader, writer io.Writer, prompt string) (bool, error) {
	if _, err := fmt.Fprint(writer, prompt); err != nil {
		return false, err
	}

	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes", nil
}
