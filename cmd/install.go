package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	longHelp := `Install all supported developer tools.
Supported tools:
- php  -> brew install php
- python3 -> brew install python3
- node -> brew install node
- go -> brew install go
- curl -> brew install curl
- git -> brew install git
- ffmpeg -> brew install ffmpeg
- tree -> brew install tree
- gh -> brew install gh`

	cmd := &cobra.Command{
		Use:     "install",
		Args:    cobra.NoArgs,
		Short:   "Install supported developer tools",
		Long:    longHelp,
		Example: "  denv install",
		RunE: func(cmd *cobra.Command, _ []string) error {
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			operations, err := buildInstallOperations(force)
			if err != nil {
				return err
			}

			if dryRun {
				for _, operation := range operations {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Would run: %s\n", operation); err != nil {
						return err
					}
				}
				return nil
			}

			for _, operation := range operations {
				if err := runInstallOperation(cmd.OutOrStdout(), operation); err != nil {
					return err
				}
			}

			_, outErr := fmt.Fprintln(cmd.OutOrStdout(), "install done")
			return outErr
		},
	}

	cmd.Flags().Bool("force", false, "install even if the tool already exists")
	cmd.Flags().Bool("dry-run", false, "show planned install operations only")
	return cmd
}

func buildInstallOperations(force bool) ([]string, error) {
	operations := make([]string, 0)
	for _, toolName := range InstallableTools() {
		toolOps, err := buildInstallOperationsForTool(toolName, force)
		if err != nil {
			return nil, err
		}
		operations = append(operations, toolOps...)
	}
	return operations, nil
}

func buildInstallOperationsForTool(toolName string, force bool) ([]string, error) {
	tool, ok := managedToolFor(toolName)
	if !ok || !tool.installable {
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}

	return buildInstallPlan(toolName, force, tool.planCheckCommand, tool.planCheckFormula)
}

func buildNodeInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("node", force, true, false)
}

func buildPHPInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("php", force, true, false)
}

func buildPython3InstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("python3", force, true, true)
}

func buildGoInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("go", force, true, false)
}

func buildCurlInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("curl", force, true, false)
}

func buildGitInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("git", force, true, false)
}

func buildFFmpegInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("ffmpeg", force, true, false)
}

func buildTreeInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("tree", force, true, false)
}

func buildGHInstallOperations(force bool) ([]string, error) {
	return buildInstallPlan("gh", force, true, false)
}

func runInstallOperation(out io.Writer, op string) error {
	args := strings.Fields(op)
	if len(args) == 0 {
		return nil
	}
	return commandRunnerWithOutput(out, args[0], args[1:]...)
}
