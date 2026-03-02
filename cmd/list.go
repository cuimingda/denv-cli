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
                version := ""
                missing := false
                installedByHomebrew := false
                var toolPath string

                installed, path, homebrewInstalled, stateErr := ToolInstallState(name)
                if stateErr != nil {
                    return stateErr
                }

                if installed {
                    toolPath = path
                    installedByHomebrew = homebrewInstalled
                } else {
                    missing = true
                }

                if showVersion && !missing {
                    if toolVersion, err := ToolVersionWithPath(name, toolPath); err == nil {
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

                toolName := ToolDisplayName(name)
                if useColorOutput(cmd.OutOrStdout()) {
                    if missing {
                        toolName = colorize(colorRed, toolName)
                    } else {
                        toolName = colorize(colorGreen, toolName)
                    }
                }

                toolName, suffix := format(toolName, version, toolPath, missing)
                if showPath && !missing && suffix != "" && useColorOutput(cmd.OutOrStdout()) {
                    pathWithOptionalBraces := ""
                    if showVersion {
                        idx := strings.LastIndex(suffix, "(")
                        if idx >= 0 && strings.HasSuffix(suffix, ")") {
                            pathWithOptionalBraces = suffix[idx:]
                            prefix := strings.TrimSpace(suffix[:idx])
                            if installedByHomebrew {
                                pathWithOptionalBraces = fmt.Sprintf("%s %s", prefix, pathWithOptionalBraces)
                            } else {
                                pathWithOptionalBraces = fmt.Sprintf("%s %s", prefix, colorize(colorRed, pathWithOptionalBraces))
                            }
                            suffix = pathWithOptionalBraces
                        }
                    } else {
                        if installedByHomebrew {
                            suffix = fmt.Sprintf("%s", suffix)
                        } else {
                            suffix = colorize(colorRed, suffix)
                        }
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

func isHomebrewPath(path string) bool {
    return strings.HasPrefix(path, "/opt/homebrew/")
}

func colorize(color string, text string) string {
    return color + text + colorReset
}
