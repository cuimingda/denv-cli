package cmd

import (
    "fmt"
    "os/exec"
    "regexp"
    "strings"
)

var (
    supportedTools = []string{
        "php",
        "python",
        "node",
        "go",
    }

    executableLookup = exec.LookPath
    commandRunner   = func(name string, args ...string) ([]byte, error) {
        return exec.Command(name, args...).CombinedOutput()
    }

    versionPatterns = []*regexp.Regexp{
        regexp.MustCompile(`\d+\.\d+\.\d+`),
        regexp.MustCompile(`\d+\.\d+`),
    }
)

var toolVersionCommands = map[string][]string{
    "php":    {"--version"},
    "python": {"--version"},
    "node":   {"--version"},
    "go":     {"version"},
}

var toolDisplayNames = map[string]string{
    "php":    "php",
    "python": "python",
    "node":   "node",
    "go":     "Go",
}

func IsCommandAvailable(name string) bool {
    _, err := executableLookup(name)
    return err == nil
}

func ToolDisplayName(name string) string {
    if display, ok := toolDisplayNames[name]; ok {
        return display
    }
    return name
}

func IsSupportedTool(name string) bool {
    for _, item := range supportedTools {
        if item == name {
            return true
        }
    }
    return false
}

func SupportedTools() []string {
    out := make([]string, len(supportedTools))
    copy(out, supportedTools)
    return out
}

func ToolVersion(name string) (string, error) {
    cmdArgs, ok := toolVersionCommands[name]
    if !ok {
        return "", fmt.Errorf("unsupported tool: %s", name)
    }

    output, err := commandRunner(name, cmdArgs...)
    if err != nil {
        return "", fmt.Errorf("get version failed: %w", err)
    }

    return extractVersion(string(output))
}

func extractVersion(out string) (string, error) {
    text := strings.TrimSpace(out)
    for _, re := range versionPatterns {
        if match := re.FindString(text); match != "" {
            return match, nil
        }
    }
    return "", fmt.Errorf("no version found")
}
