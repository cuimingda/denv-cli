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
                if !IsCommandAvailable(name) {
                    _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s not found\n", name)
                    if err != nil {
                        return err
                    }
                    continue
                }

                version, err := ToolVersion(name)
                if err != nil {
                    _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s not found\n", name)
                    if err != nil {
                        return err
                    }
                    continue
                }

                _, err = fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ToolDisplayName(name), version)
                if err != nil {
                    return err
                }
            }
            return nil
        },
    }
}
