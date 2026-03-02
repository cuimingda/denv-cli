package cmd

import (
	"fmt"
	"io"

	"github.com/cuimingda/denv-cli/internal/denv"
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
- gh -> brew install gh
- git -> brew install git
- ffmpeg -> brew install ffmpeg
- tree -> brew install tree`

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
	return denv.BuildInstallOperations(toolRuntime(), force)
}

func buildInstallOperationsForTool(toolName string, force bool) ([]string, error) {
	return denv.BuildInstallOperationsForTool(toolRuntime(), toolName, force)
}

func buildNodeInstallOperations(force bool) ([]string, error) {
	return denv.BuildNodeInstallOperations(toolRuntime(), force)
}

func buildPHPInstallOperations(force bool) ([]string, error) {
	return denv.BuildPHPInstallOperations(toolRuntime(), force)
}

func buildPython3InstallOperations(force bool) ([]string, error) {
	return denv.BuildPython3InstallOperations(toolRuntime(), force)
}

func buildGoInstallOperations(force bool) ([]string, error) {
	return denv.BuildGoInstallOperations(toolRuntime(), force)
}

func buildCurlInstallOperations(force bool) ([]string, error) {
	return denv.BuildCurlInstallOperations(toolRuntime(), force)
}

func buildGitInstallOperations(force bool) ([]string, error) {
	return denv.BuildGitInstallOperations(toolRuntime(), force)
}

func buildFFmpegInstallOperations(force bool) ([]string, error) {
	return denv.BuildFFmpegInstallOperations(toolRuntime(), force)
}

func buildTreeInstallOperations(force bool) ([]string, error) {
	return denv.BuildTreeInstallOperations(toolRuntime(), force)
}

func buildGHInstallOperations(force bool) ([]string, error) {
	return denv.BuildGHInstallOperations(toolRuntime(), force)
}

func runInstallOperation(out io.Writer, op string) error {
	return denv.RunInstallOperation(out, toolRuntime(), op)
}
