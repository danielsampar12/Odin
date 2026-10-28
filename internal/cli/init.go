package cli

import (
	"fmt"
	"path/filepath"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var agentsMD bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold Odin for the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}

			result, err := doctor.Collect(cmd.Context(), cwd)
			if err != nil {
				return err
			}

			global, err := config.Load(result.GlobalConfig.Path)
			if err != nil {
				return err
			}
			companion, _ := companions.Lookup(global.Companion.Default)
			projectName := filepath.Base(cwd)

			projectDir := config.ProjectDir(cwd)
			projectGeneratedDir := config.ProjectGeneratedDir(cwd)
			projectConfigPath := config.ProjectConfigPath(cwd)
			projectRulesPath := config.ProjectRulesPath(cwd)
			agentsPath := config.ProjectAgentsPath(cwd)

			dirCreated, err := ensureDir(projectDir)
			if err != nil {
				return err
			}

			generatedDirCreated, err := ensureDir(projectGeneratedDir)
			if err != nil {
				return err
			}

			projectConfigCreated, err := config.WriteFileIfMissing(projectConfigPath, config.DefaultProjectConfig(config.ProjectSettings{
				Name:             projectName,
				AgentProvider:    "opencode",
				RuntimeProvider:  "ollama",
				MemoryProvider:   "none",
				MemoryHall:       projectName,
				CompanionDefault: companion.Key,
			}), 0o644)
			if err != nil {
				return err
			}

			projectRulesCreated, err := config.WriteFileIfMissing(projectRulesPath, config.DefaultRules(), 0o644)
			if err != nil {
				return err
			}

			agentsCreated := false
			if agentsMD {
				if agentsCreated, err = config.WriteFileIfMissing(agentsPath, config.DefaultAgents(projectName, companion.Name), 0o644); err != nil {
					return err
				}
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Initializing Odin for project %q.\n", projectName)
			fmt.Fprintln(out)
			if !result.InGitRepo {
				fmt.Fprintln(out, "Note: this directory is not inside a Git repository. Odin can still scaffold project files here.")
				fmt.Fprintln(out)
			}

			fmt.Fprintln(out, "Project files:")
			if dirCreated {
				fmt.Fprintf(out, "- Created %s\n", displayPath(projectDir))
			} else {
				fmt.Fprintf(out, "- %s already exists\n", displayPath(projectDir))
			}
			if generatedDirCreated {
				fmt.Fprintf(out, "- Created %s\n", displayPath(projectGeneratedDir))
			} else {
				fmt.Fprintf(out, "- %s already exists\n", displayPath(projectGeneratedDir))
			}
			printFileAction(out, projectConfigPath, projectConfigCreated)
			printFileAction(out, projectRulesPath, projectRulesCreated)
			if agentsMD {
				printFileAction(out, agentsPath, agentsCreated)
			}
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Default companion: %s\n", companion.Name)
			fmt.Fprintln(out, "Memory: off by default. To opt in to MemPalace, set [memory] provider = \"mempalace\" in .odin/config.toml.")
			printGitignoreHint(cmd.Context(), out, cwd)
			if !result.GlobalConfig.Exists {
				fmt.Fprintln(out, "No ~/.odin/config.toml yet: run `odin setup` so Odin picks a model for this machine.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&agentsMD, "agents-md", false, "Also create a repository-root AGENTS.md if there is none")
	return cmd
}
