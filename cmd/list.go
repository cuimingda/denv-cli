package cmd

import (
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/spf13/cobra"
)

const (
    colorGreen = "\033[32m"
    colorRed   = "\033[31m"
    colorReset = "\033[0m"
)

func NewListCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List supported developer tools",
        RunE: func(cmd *cobra.Command, _ []string) error {
            showVersion, _ := cmd.Flags().GetBool("version")
            showPath, _ := cmd.Flags().GetBool("path")

            format := func(name string, version string, toolPath string, missing bool) (string, string) {
                if missing {
                    return name, "not found"
                }

                suffixParts := make([]string, 0, 2)
                if showVersion {
                    suffixParts = append(suffixParts, version)
                }
                if showPath {
                    if showVersion {
                        suffixParts = append(suffixParts, fmt.Sprintf("(%s)", toolPath))
                    } else {
                        suffixParts = append(suffixParts, toolPath)
                    }
                }

                return name, strings.Join(suffixParts, " ")
            }

            for _, name := range SupportedTools() {
                toolPath := ""
                version := ""
                missing := false

                if path, err := CommandPath(name); err == nil {
                    toolPath = path
                } else {
                    missing = true
                }

                if showVersion && !missing {
                    if toolVersion, err := ToolVersion(name); err == nil {
                        version = toolVersion
                    } else {
                        missing = true
                    }
                }

                if !showVersion && !showPath {
                    displayName := ToolDisplayName(name)
                    if useColorOutput(cmd.OutOrStdout()) {
                        if missing {
                            displayName = colorize(colorRed, displayName)
                        } else {
                            displayName = colorize(colorGreen, displayName)
                        }
                    }
                    if _, err := fmt.Fprintln(cmd.OutOrStdout(), displayName); err != nil {
                        return err
                    }
                    continue
                }

                toolName, suffix := format(ToolDisplayName(name), version, toolPath, missing)
                if useColorOutput(cmd.OutOrStdout()) {
                    if missing {
                        toolName = colorize(colorRed, toolName)
                    } else {
                        toolName = colorize(colorGreen, toolName)
                    }
                }
                line := toolName
                if suffix != "" {
                    line = fmt.Sprintf("%s %s", toolName, suffix)
                }

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

func useColorOutput(out io.Writer) bool {
    _, ok := out.(*os.File)
    return ok
}

func colorize(color string, text string) string {
    return color + text + colorReset
}
