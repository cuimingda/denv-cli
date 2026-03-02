package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List supported developer tools",
        RunE: func(cmd *cobra.Command, _ []string) error {
            showVersion, _ := cmd.Flags().GetBool("version")
            showPath, _ := cmd.Flags().GetBool("path")

            format := func(name string, version string, toolPath string, missing bool) string {
                if missing {
                    return fmt.Sprintf("%s not found", name)
                }

                if showVersion && showPath {
                    return fmt.Sprintf("%s %s (%s)", name, version, toolPath)
                }

                if showVersion {
                    return fmt.Sprintf("%s %s", name, version)
                }

                if showPath {
                    return fmt.Sprintf("%s %s", name, toolPath)
                }

                return name
            }

            for _, name := range SupportedTools() {
                toolPath := ""
                version := ""
                missing := false

                if showVersion || showPath {
                    if path, err := CommandPath(name); err == nil {
                        toolPath = path
                    } else {
                        missing = true
                    }

                    if !missing && showVersion {
                        if toolVersion, err := ToolVersion(name); err == nil {
                            version = toolVersion
                        } else {
                            missing = true
                        }
                    }
                }

                line := format(ToolDisplayName(name), version, toolPath, missing)
                if _, err := fmt.Fprintln(cmd.OutOrStdout(), line); err != nil {
                    return err
                }
            }
            return nil
        },
    }

    cmd.Flags().Bool("version", false, "show versions for discovered tools")
    cmd.Flags().Bool("path", false, "show executable paths for discovered tools")
    return cmd
}
