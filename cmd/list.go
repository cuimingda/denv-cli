package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "list",
        Short: "List supported developer tools",
        RunE: func(cmd *cobra.Command, _ []string) error {
            for _, name := range SupportedTools() {
                _, err := fmt.Fprintln(cmd.OutOrStdout(), name)
                if err != nil {
                    return err
                }
            }
            return nil
        },
    }
}
