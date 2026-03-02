package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "install <tool_name>",
        Args:  cobra.ExactArgs(1),
        Short: "Install a supported developer tool",
        RunE: func(cmd *cobra.Command, args []string) error {
            toolName := args[0]

            if !IsInstallableTool(toolName) {
                return fmt.Errorf("unsupported tool: %s", toolName)
            }

            if err := InstallTool(toolName); err != nil {
                return err
            }

            _, err := fmt.Fprintln(cmd.OutOrStdout(), "done")
            return err
        },
    }
}
