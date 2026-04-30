package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/danielsampar12/odin/internal/setupplan"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	var dryRun bool
	var assumeYes bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Plan or execute global Odin setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := workingDir()
			if err != nil {
				return err
			}

			plan, err := setupplan.Build(cmd.Context(), cwd)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if dryRun && assumeYes {
				return errors.New("--dry-run and --yes cannot be used together")
			}

			renderSetupPlan(out, plan)
			fmt.Fprintln(out)
			if dryRun {
				fmt.Fprintln(out, "No changes were made.")
				return nil
			}

			if !assumeYes {
				proceed, err := confirmAction(cmd.InOrStdin(), out, "Proceed with setup? [y/N]: ")
				if err != nil {
					return err
				}
				if !proceed {
					fmt.Fprintln(out)
					fmt.Fprintln(out, "Cancelled. No changes were made.")
					return nil
				}
				fmt.Fprintln(out)
			}

			fmt.Fprintln(out, "Executing setup plan.")
			fmt.Fprintln(out)

			execution, err := setupplan.Execute(cmd.Context(), cwd, plan, setupplan.ExecuteOptions{
				Stdin:  cmd.InOrStdin(),
				Stdout: out,
				Stderr: cmd.ErrOrStderr(),
			})
			if err != nil {
				return err
			}

			fmt.Fprintln(out)
			renderSetupExecutionSummary(out, execution, plan)
			if execution.HasCriticalFailure() {
				return errors.New("setup did not complete successfully")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Inspect the machine and print the setup plan without making changes")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "Execute the supported setup steps without prompting")
	return cmd
}

func renderSetupPlan(out io.Writer, plan setupplan.Plan) {
	fmt.Fprintln(out, "Odin setup plan")
	fmt.Fprintln(out)
	renderMachineSummary(out, plan)
	fmt.Fprintln(out)
	renderRecommendedStack(out, plan)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Plan:")
	for _, step := range plan.Steps {
		fmt.Fprintf(out, "%s %s: %s\n", setupStepIndicator(step.Status), step.Name, step.Reason)
		if step.Method != "" {
			fmt.Fprintf(out, "  Method: %s\n", step.Method)
		}
		if step.Command != "" {
			fmt.Fprintf(out, "  Command: %s\n", step.Command)
		}
		if step.Notes != "" {
			fmt.Fprintf(out, "  Notes: %s\n", step.Notes)
		}
		fmt.Fprintf(out, "  Mode: %s\n", setupExecutionLabel(step))
	}
}

func renderMachineSummary(out io.Writer, plan setupplan.Plan) {
	fmt.Fprintln(out, "Machine:")
	fmt.Fprintf(out, "- OS: %s\n", plan.Result.OS.Name)
	fmt.Fprintf(out, "- Architecture: %s\n", plan.Result.OS.Arch)
	if plan.Result.RAMGB > 0 {
		fmt.Fprintf(out, "- RAM: %dGB\n", plan.Result.RAMGB)
	} else {
		fmt.Fprintln(out, "- RAM: unavailable")
	}
	fmt.Fprintf(out, "- GPU: %s\n", plan.Result.GPU.Summary)
}

func renderRecommendedStack(out io.Writer, plan setupplan.Plan) {
	fmt.Fprintln(out, "Recommended stack:")
	fmt.Fprintf(out, "- Profile: %s\n", plan.Recommendation.Profile)
	fmt.Fprintln(out, "- Runtime: Ollama")
	fmt.Fprintf(out, "- Agent: %s\n", plan.Recommendation.Agent)
	fmt.Fprintf(out, "- Memory: %s\n", plan.Recommendation.Memory)
	fmt.Fprintf(out, "- Model: %s\n", plan.Recommendation.Model)
	fmt.Fprintf(out, "- Companion: %s\n", plan.Recommendation.CompanionName)
	fmt.Fprintf(out, "- Why: %s\n", plan.Recommendation.Reason)
}

func setupStepIndicator(status setupplan.StepStatus) string {
	switch status {
	case setupplan.StatusDone:
		return "✓"
	case setupplan.StatusNeeded:
		return "!"
	case setupplan.StatusManual:
		return "?"
	case setupplan.StatusUnsupported:
		return "x"
	default:
		return "-"
	}
}

func setupExecutionLabel(step setupplan.Step) string {
	switch step.Status {
	case setupplan.StatusDone:
		return "already done"
	case setupplan.StatusSkipped:
		return "deferred"
	case setupplan.StatusUnsupported:
		return "unsupported for automatic setup currently"
	}

	switch {
	case step.Automatic && step.RequiresConfirmation:
		return "future automatic after confirmation"
	case step.Automatic:
		return "automatic"
	case step.RequiresConfirmation:
		return "manual or confirmation-required"
	default:
		return "manual"
	}
}

func renderSetupExecutionSummary(out io.Writer, result setupplan.ExecutionResult, plan setupplan.Plan) {
	fmt.Fprintln(out, "Setup summary:")
	for _, step := range result.Steps {
		fmt.Fprintf(out, "%s %s: %s\n", executionIndicator(step.Outcome), step.Step.Name, step.Message)
		if step.Err != nil {
			fmt.Fprintf(out, "  Error: %s\n", step.Err)
		}
	}

	completed, skipped, failed, manual := result.Counts()
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Completed: %d\n", completed)
	fmt.Fprintf(out, "Skipped: %d\n", skipped)
	fmt.Fprintf(out, "Failed: %d\n", failed)
	fmt.Fprintf(out, "Manual steps remaining: %d\n", manual)

	if plan.Result.InGitRepo {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "You are inside a Git repository. Run `odin init` to initialize this project.")
	}
}

func executionIndicator(outcome setupplan.Outcome) string {
	switch outcome {
	case setupplan.OutcomeCompleted:
		return "✓"
	case setupplan.OutcomeFailed:
		return "x"
	case setupplan.OutcomeManual:
		return "?"
	default:
		return "-"
	}
}
