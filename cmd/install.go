package cmd

import (
	"fmt"
	"io"

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
- git -> brew install git`

	cmd := &cobra.Command{
		Use:     "install",
		Args:    cobra.NoArgs,
		Short:   "Install supported developer tools",
		Long:    longHelp,
		Example: "  denv install",
		RunE: func(cmd *cobra.Command, _ []string) error {
			force, _ := cmd.Flags().GetBool("force")

			for _, toolName := range InstallableTools() {
				if err := installToolWithOutput(cmd.OutOrStdout(), toolName, force); err != nil {
					return err
				}
			}

			_, outErr := fmt.Fprintln(cmd.OutOrStdout(), "install done")
			return outErr
		},
	}

	cmd.Flags().Bool("force", false, "install even if the tool already exists")
	return cmd
}

func installToolWithOutput(out io.Writer, toolName string, force bool) error {
	switch toolName {
	case "node":
		return InstallNodeWithOutput(out, force)
	case "php":
		return InstallPHPWithOutput(out, force)
	case "python3":
		return InstallPython3WithOutput(out, force)
	case "go":
		return InstallGoWithOutput(out, force)
	case "curl":
		return InstallCurlWithOutput(out, force)
	case "git":
		return InstallGitWithOutput(out, force)
	default:
		return fmt.Errorf("unsupported tool: %s", toolName)
	}
}
