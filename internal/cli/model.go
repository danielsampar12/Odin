package cli

import (
	"flag"
	"fmt"

	"github.com/danielsampar12/odin/internal/config"
	"github.com/danielsampar12/odin/internal/doctor"
	"github.com/danielsampar12/odin/internal/models"
	ollamaplugin "github.com/danielsampar12/odin/internal/plugins/ollama"
	"github.com/spf13/cobra"
)

func newModelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model",
		Short: "Inspect Odin's local model state and recommendations",
	}

	cmd.AddCommand(
		newModelListCmd(),
		newModelRecommendCmd(),
		newModelPullCmd(),
	)

	return cmd
}

func newModelListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed local Ollama models",
		RunE: func(cmd *cobra.Command, args []string) error {
			globalConfigPath := config.GlobalConfigPath()
			baseURL := config.ResolveGlobalRuntimeBaseURL(globalConfigPath, ollamaplugin.DefaultBaseURL)
			commandStatus := ollamaplugin.Detect()
			apiStatus := ollamaplugin.Probe(cmd.Context(), baseURL)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Ollama models (%s)\n", apiStatus.BaseURL)
			fmt.Fprintln(out)

			if !apiStatus.APIAvailable {
				if !commandStatus.Installed {
					fmt.Fprintln(out, "Ollama is not installed and the local API is not reachable.")
					fmt.Fprintln(out, "Install Ollama, start it, then rerun `odin model list`.")
					return nil
				}
				fmt.Fprintf(out, "Ollama API is not reachable: %s\n", apiStatus.Error)
				fmt.Fprintln(out, "Start Ollama and rerun this command. For local use, the official default is `http://localhost:11434`.")
				return nil
			}

			if !commandStatus.Installed {
				fmt.Fprintln(out, "Note: the Ollama API is reachable, but the `ollama` command is not on PATH.")
				fmt.Fprintln(out)
			}

			if len(apiStatus.Models) == 0 {
				fmt.Fprintln(out, "No local models are currently installed in Ollama.")
				return nil
			}

			for _, model := range apiStatus.Models {
				fmt.Fprintf(out, "- %s\n", formatModelEntry(model))
			}

			return nil
		},
	}
}

func newModelRecommendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "recommend",
		Short: "Recommend a local coding model for this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}

			result, err := doctor.Collect(cmd.Context(), cwd)
			if err != nil {
				return err
			}

			recommendation := models.RecommendCodingModel(result.RAMGB, result.GPU)
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Odin model recommendation")
			fmt.Fprintln(out)
			if result.RAMGB > 0 {
				fmt.Fprintf(out, "Hardware: %dGB RAM, %s\n", result.RAMGB, result.GPU.Summary)
			} else {
				fmt.Fprintf(out, "Hardware: RAM unavailable, %s\n", result.GPU.Summary)
			}
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Recommended coding model: %s\n", recommendation.Recommended)
			fmt.Fprintf(out, "Fallback model: %s\n", recommendation.Fallback)
			fmt.Fprintf(out, "Reason: %s\n", recommendation.Reason)
			if recommendation.OptionalLarger != "" {
				fmt.Fprintf(out, "Optional larger tier: %s\n", recommendation.OptionalLarger)
			}

			return nil
		},
	}
}

func newModelPullCmd() *cobra.Command {
	var assumeYes bool

	cmd := &cobra.Command{
		Use:   "pull <name>",
		Short: "Pull a local model into Ollama",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return nil
			}
			cmd.PrintErrln("Error: expected exactly 1 model name argument")
			return flag.ErrHelp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			modelName := args[0]
			globalConfigPath := config.GlobalConfigPath()
			baseURL := config.ResolveGlobalRuntimeBaseURL(globalConfigPath, ollamaplugin.DefaultBaseURL)
			apiStatus := ollamaplugin.Probe(cmd.Context(), baseURL)
			out := cmd.OutOrStdout()

			if !apiStatus.APIAvailable {
				fmt.Fprintln(out, "Ollama API is not responding.")
				if apiStatus.Error != "" {
					fmt.Fprintf(out, "Details: %s\n", apiStatus.Error)
				}
				fmt.Fprintln(out, "Try starting Ollama and rerun this command.")
				fmt.Fprintln(out, "If needed, the official CLI command to start the local server is `ollama serve`.")
				return nil
			}

			if !assumeYes {
				ok, err := confirmAction(
					cmd.InOrStdin(),
					out,
					fmt.Sprintf("Pull model %q from %s? This may download many GB. [y/N]: ", modelName, apiStatus.BaseURL),
				)
				if err != nil {
					return err
				}
				if !ok {
					fmt.Fprintln(out)
					fmt.Fprintln(out, "Cancelled.")
					return nil
				}
				fmt.Fprintln(out)
			}

			fmt.Fprintf(out, "Pulling %s from %s\n", modelName, apiStatus.BaseURL)
			fmt.Fprintln(out, "This may take a while.")

			response, err := ollamaplugin.PullModel(cmd.Context(), baseURL, ollamaplugin.PullRequest{
				Model: modelName,
			})
			if err != nil {
				fmt.Fprintf(out, "Pull failed: %s\n", err)
				return nil
			}

			if response.Status != "" {
				fmt.Fprintf(out, "Pull complete: %s (%s)\n", modelName, response.Status)
			} else {
				fmt.Fprintf(out, "Pull complete: %s\n", modelName)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Pull without confirmation")
	return cmd
}
