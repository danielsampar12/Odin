package cli

import (
	"errors"
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
			baseURL, err := globalBaseURL()
			if err != nil {
				return err
			}
			commandStatus := ollamaplugin.Detect()
			apiStatus := ollamaplugin.Probe(cmd.Context(), baseURL)

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Ollama models (%s)\n", apiStatus.BaseURL)
			fmt.Fprintln(out)

			if !apiStatus.APIAvailable {
				if !commandStatus.Installed {
					return errors.New("Ollama is not installed and the local API is not reachable; install Ollama, start it, then rerun `odin model list`")
				}
				return fmt.Errorf("Ollama API is not reachable at %s (%s); start Ollama (`ollama serve`) and rerun this command", apiStatus.BaseURL, apiStatus.Error)
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
	var useCase string

	cmd := &cobra.Command{
		Use:   "recommend",
		Short: "Recommend a local coding model for this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if useCase != string(models.UseCaseAgent) && useCase != string(models.UseCaseChat) {
				return fmt.Errorf("unknown --use-case %q: use %q or %q", useCase, models.UseCaseAgent, models.UseCaseChat)
			}

			cwd, err := workingDir()
			if err != nil {
				return err
			}

			result, err := doctor.Collect(cmd.Context(), cwd)
			if err != nil {
				return err
			}

			recommendation := models.Recommend(result.RAMGB, result.GPU, models.UseCase(useCase))
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Odin model recommendation")
			fmt.Fprintln(out)
			if result.RAMGB > 0 {
				fmt.Fprintf(out, "Hardware: %dGB RAM, %s\n", result.RAMGB, result.GPU.Summary)
			} else {
				fmt.Fprintf(out, "Hardware: RAM unavailable, %s\n", result.GPU.Summary)
			}
			fmt.Fprintln(out)
			fmt.Fprintf(out, "Use case: %s\n", useCase)
			fmt.Fprintf(out, "Recommended coding model: %s\n", recommendation.Recommended)
			fmt.Fprintf(out, "Context: %d tokens\n", recommendation.Context)
			if useCase == string(models.UseCaseAgent) {
				fmt.Fprintln(out, "  Set num_ctx to this: Ollama's default context is too small for OpenCode's prompt.")
			}
			fmt.Fprintf(out, "Fallback model: %s\n", recommendation.Fallback)
			fmt.Fprintf(out, "Reason: %s\n", recommendation.Reason)
			if recommendation.OptionalLarger != "" {
				fmt.Fprintf(out, "Optional larger tier: %s\n", recommendation.OptionalLarger)
				fmt.Fprintf(out, "  %s\n", recommendation.OptionalLargerReason)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&useCase, "use-case", string(models.UseCaseAgent), "What the model is for: agent (OpenCode with tool calls) or chat (questions, no tools)")
	return cmd
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
			baseURL, err := globalBaseURL()
			if err != nil {
				return err
			}
			apiStatus := ollamaplugin.Probe(cmd.Context(), baseURL)
			out := cmd.OutOrStdout()

			if !apiStatus.APIAvailable {
				return fmt.Errorf("Ollama API is not responding at %s (%s); start Ollama (`ollama serve`) and rerun this command", apiStatus.BaseURL, apiStatus.Error)
			}
			if ollamaModelInstalled(apiStatus.Models, modelName) {
				fmt.Fprintf(out, "%s is already installed.\n", modelName)
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
				return fmt.Errorf("pull %s failed: %w", modelName, err)
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

// globalBaseURL is the Ollama URL from ~/.odin/config.toml, or Ollama's default when it is unset.
func globalBaseURL() (string, error) {
	global, err := config.Load(config.GlobalConfigPath())
	if err != nil || global.Runtime.BaseURL == "" {
		return ollamaplugin.DefaultBaseURL, err
	}
	return global.Runtime.BaseURL, nil
}
