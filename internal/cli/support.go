package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/danielsampar12/odin/internal/plugins"
	"github.com/danielsampar12/odin/internal/system"
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

// printGitignoreHint suggests ignoring .odin/generated/, which is machine-specific. Odin never edits .gitignore.
func printGitignoreHint(ctx context.Context, out io.Writer, cwd string) {
	// check-ignore exits 0 when the path is ignored, 1 when it is not, 128 outside a repository.
	// --no-index also catches a generated file that was committed before it was ignored.
	_, err := system.RunCommand(ctx, cwd, "git", "check-ignore", "-q", "--no-index", filepath.Join(".odin", "generated", "opencode.jsonc"))
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		fmt.Fprintln(out, "Tip: .odin/generated/ is machine-specific. Add this line to .gitignore: .odin/generated/")
	}
}

func printWarnings(out io.Writer, warnings []string) {
	for _, warning := range warnings {
		fmt.Fprintf(out, "Warning: %s\n", warning)
	}
	if len(warnings) > 0 {
		fmt.Fprintln(out)
	}
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

func printToolStatusLine(writer io.Writer, label string, tool plugins.Status) {
	fmt.Fprintf(writer, "- %s: %s", label, installedLabel(tool.Installed))
	if tool.Path != "" {
		fmt.Fprintf(writer, " (%s", tool.Path)
		if tool.Details != "" {
			fmt.Fprintf(writer, ", %s", tool.Details)
		}
		fmt.Fprintln(writer, ")")
		return
	}
	if tool.Details != "" {
		fmt.Fprintf(writer, " (%s)", tool.Details)
	}
	fmt.Fprintln(writer)
}

func configValueOrUnset(value string) string {
	if strings.TrimSpace(value) == "" {
		return "not set"
	}

	return value
}

func derivedSuffix(derived bool) string {
	if derived {
		return " (derived)"
	}

	return ""
}

func joinCommand(parts []string) string {
	return strings.Join(parts, " ")
}
