package cli

import (
	"fmt"
	"path/filepath"

	"github.com/danielsampar12/odin/internal/companions"
	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/models"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
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

			profile := detectProfileFromConfig(result)
			companion := companions.DefaultForProfile(profile)
			projectName := filepath.Base(cwd)
			modelRecommendation := models.RecommendCodingModel(result.RAMGB, result.GPU)

			projectDir := config.ProjectDir(cwd)
			projectConfigPath := config.ProjectConfigPath(cwd)
			projectRulesPath := config.ProjectRulesPath(cwd)
			agentsPath := config.ProjectAgentsPath(cwd)

			dirCreated, err := ensureDir(projectDir)
			if err != nil {
				return err
			}

			projectConfigCreated, err := writeFileIfMissing(projectConfigPath, config.DefaultProjectConfig(config.ProjectSettings{
				Name:             projectName,
				AgentProvider:    "opencode",
				RuntimeProvider:  "ollama",
				MemoryProvider:   "mempalace",
				MemoryHall:       projectName,
				ModelDefault:     modelRecommendation.Recommended,
				CompanionDefault: companion.Key,
			}), 0o644)
			if err != nil {
				return err
			}

			projectRulesCreated, err := writeFileIfMissing(projectRulesPath, config.DefaultRules(projectName, companion.Name), 0o644)
			if err != nil {
				return err
			}

			agentsCreated, err := writeFileIfMissing(agentsPath, config.DefaultAgents(projectName, companion.Name), 0o644)
			if err != nil {
				return err
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
			printFileAction(out, projectConfigPath, projectConfigCreated)
			printFileAction(out, projectRulesPath, projectRulesCreated)
			printFileAction(out, agentsPath, agentsCreated)
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Default companion: %s\n", companion.Name)
			fmt.Fprintf(out, "Memory hall: %s\n", projectName)
			fmt.Fprintln(out, "MemPalace project integration is intentionally scaffolded for now.")

			return nil
		},
	}
}

func detectProfileFromConfig(result doctor.Result) string {
	if result.GlobalConfig.Exists {
		profile, err := config.ReadGlobalProfile(result.GlobalConfig.Path)
		if err == nil && profile != "" {
			return profile
		}
	}

	return detectProfile(result.CurrentDir, result.InGitRepo)
}
