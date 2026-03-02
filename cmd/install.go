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
	switch toolName {
	case "node":
		return buildNodeInstallOperations(force)
	case "php":
		return buildPHPInstallOperations(force)
	case "python3":
		return buildPython3InstallOperations(force)
	case "go":
		return buildGoInstallOperations(force)
	case "curl":
		return buildCurlInstallOperations(force)
	case "git":
		return buildGitInstallOperations(force)
	case "ffmpeg":
		return buildFFmpegInstallOperations(force)
	case "tree":
		return buildTreeInstallOperations(force)
	case "gh":
		return buildGHInstallOperations(force)
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
}

func buildNodeInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && (IsCommandAvailable("node") || IsCommandAvailable("npm")) {
		return nil, nil
	}
	return []string{"brew install node"}, nil
}

func buildPHPInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("php") {
		return nil, nil
	}
	return []string{"brew install php"}, nil
}

func buildPython3InstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}

	if !force && IsCommandAvailable("python3") {
		return nil, nil
	}

	if !force {
		installed, err := IsBrewFormulaInstalled("python3")
		if err != nil {
			return nil, fmt.Errorf("check python3 install status failed: %w", err)
		}
		if installed {
			return nil, nil
		}
	}

	return []string{"brew install python3"}, nil
}

func buildGoInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("go") {
		return nil, nil
	}
	return []string{"brew install go"}, nil
}

func buildCurlInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("curl") {
		return nil, nil
	}
	return []string{"brew install curl", "brew link curl --force"}, nil
}

func buildGitInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("git") {
		return nil, nil
	}
	return []string{"brew install git"}, nil
}

func buildFFmpegInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("ffmpeg") {
		return nil, nil
	}
	return []string{"brew install ffmpeg"}, nil
}

func buildTreeInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("tree") {
		return nil, nil
	}
	return []string{"brew install tree"}, nil
}

func buildGHInstallOperations(force bool) ([]string, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("homebrew is not installed")
	}
	if !force && IsCommandAvailable("gh") {
		return nil, nil
	}
	return []string{"brew install gh"}, nil
}

func runInstallOperation(out io.Writer, op string) error {
	args := strings.Fields(op)
	if len(args) == 0 {
		return nil
	}
	return commandRunnerWithOutput(out, args[0], args[1:]...)
}
