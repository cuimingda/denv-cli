package cmd

import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/spf13/cobra"
)

func NewOutdatedCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "outdated",
        Short: "Show outdated status for supported developer tools",
        RunE: func(cmd *cobra.Command, _ []string) error {
            for _, name := range SupportedTools() {
                if !IsCommandAvailable(name) {
                    if _, err := fmt.Fprintln(cmd.OutOrStdout(), ToolDisplayName(name)+" not found"); err != nil {
                        return err
                    }
                    continue
                }

                current, err := ToolVersion(name)
                if err != nil {
                    if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ToolDisplayName(name), "invalid current version"); err != nil {
                        return err
                    }
                    continue
                }

                latest, err := ToolLatestVersion(name)
                if err != nil {
                    if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ToolDisplayName(name), "invalid latest version"); err != nil {
                        return err
                    }
                    continue
                }

                if cmpVersions(current, latest) < 0 {
                    currentVersion := current
                    if useColorOutput(cmd.OutOrStdout()) {
                        currentVersion = colorize(colorRed, current)
                    }
                    if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s < %s\n", ToolDisplayName(name), currentVersion, latest); err != nil {
                        return err
                    }
                    continue
                }

                currentVersion := current
                if useColorOutput(cmd.OutOrStdout()) {
                    currentVersion = colorize(colorGreen, current)
                }
                if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", ToolDisplayName(name), currentVersion); err != nil {
                    return err
                }
            }
            return nil
        },
    }
}

func ToolLatestVersion(name string) (string, error) {
    if name == "npm" {
        return toolLatestVersionByNpm()
    }

    formula, ok := brewFormulaForTool(name)
    if !ok {
        return "", fmt.Errorf("unsupported tool: %s", name)
    }
    return toolLatestVersionByBrew(formula)
}

func toolLatestVersionByBrew(formula string) (string, error) {
    output, err := commandRunner("brew", "info", "--json=v2", formula)
    if err != nil {
        return "", fmt.Errorf("brew info failed: %w", err)
    }

    var payload struct {
        Formulae []struct {
            Versions struct {
                Stable string `json:"stable"`
            } `json:"versions"`
        } `json:"formulae"`
    }

    if err := json.Unmarshal(output, &payload); err == nil {
        if len(payload.Formulae) > 0 && payload.Formulae[0].Versions.Stable != "" {
            return payload.Formulae[0].Versions.Stable, nil
        }
    }

    var list []struct {
        Versions struct {
            Stable string `json:"stable"`
        } `json:"versions"`
    }

    if err := json.Unmarshal(output, &list); err == nil {
        if len(list) > 0 && list[0].Versions.Stable != "" {
            return list[0].Versions.Stable, nil
        }
    }

    return "", fmt.Errorf("failed to parse latest version")
}

func toolLatestVersionByNpm() (string, error) {
    output, err := commandRunner("npm", "view", "npm", "version")
    if err != nil {
        return "", fmt.Errorf("npm latest version failed: %w", err)
    }

    return extractVersion(string(output))
}

func brewFormulaForTool(name string) (string, bool) {
    formulas := map[string]string{
        "php":     "php",
        "python3": "python3",
        "node":    "node",
        "go":      "go",
        "npm":     "node",
        "curl":    "curl",
        "gh":      "gh",
        "git":     "git",
        "ffmpeg":  "ffmpeg",
        "tree":    "tree",
    }

    formula, ok := formulas[name]
    return formula, ok
}

func cmpVersions(current string, latest string) int {
    currentParts := splitVersionParts(current)
    latestParts := splitVersionParts(latest)

    maxLen := len(currentParts)
    if len(latestParts) > maxLen {
        maxLen = len(latestParts)
    }

    for len(currentParts) < maxLen {
        currentParts = append(currentParts, 0)
    }
    for len(latestParts) < maxLen {
        latestParts = append(latestParts, 0)
    }

    for i := 0; i < maxLen; i++ {
        if currentParts[i] < latestParts[i] {
            return -1
        }
        if currentParts[i] > latestParts[i] {
            return 1
        }
    }

    return 0
}

func splitVersionParts(version string) []int {
    fields := strings.FieldsFunc(version, func(r rune) bool {
        return r == '.' || r == '-'
    })

    parts := make([]int, 0, len(fields))
    for _, field := range fields {
        part := 0
        for i := 0; i < len(field); i++ {
            if field[i] < '0' || field[i] > '9' {
                part = -1
                break
            }
            part = part*10 + int(field[i]-'0')
        }
        if part >= 0 {
            parts = append(parts, part)
        }
    }
    return parts
}
