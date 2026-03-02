package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
    longHelp := `Install a supported developer tool.
Supported tools:
- php  -> brew install php
- python3 -> brew install python3
- node -> brew install node
- go -> brew install go`

    return &cobra.Command{
        Use:   "install <tool_name>",
        Args:  cobra.ExactArgs(1),
        Short: "Install a supported developer tool: php, python3, node, go",
        Long:  longHelp,
        Example: `  denv install php
  denv install python3
  denv install node
  denv install go`,
        RunE: func(cmd *cobra.Command, args []string) error {
            toolName := args[0]

            if !IsInstallableTool(toolName) {
                return fmt.Errorf("unsupported tool: %s", toolName)
            }

            var err error
            switch toolName {
            case "node":
                err = InstallNodeWithOutput(cmd.OutOrStdout())
            case "php":
                err = InstallPHPWithOutput(cmd.OutOrStdout())
            case "python3":
                err = InstallPython3WithOutput(cmd.OutOrStdout())
            case "go":
                err = InstallGoWithOutput(cmd.OutOrStdout())
            default:
                return fmt.Errorf("unsupported tool: %s", toolName)
            }

            if err != nil {
                return err
            }

            _, outErr := fmt.Fprintln(cmd.OutOrStdout(), "done")
            return outErr
        },
    }
}
