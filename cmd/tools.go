package cmd

import (
    "fmt"
    "os/exec"
    "regexp"
    "strings"
)

var listedTools = []string{
    "php",
    "python3",
    "node",
    "go",
    "npm",
}

var installableTools = []string{
    "php",
    "python3",
    "node",
    "go",
}

var (
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
    "php":     {"--version"},
    "python3": {"--version"},
    "node":    {"--version"},
    "go":      {"version"},
    "npm":     {"--version"},
}

var toolDisplayNames = map[string]string{
    "php":     "php",
    "python3": "python3",
    "node":    "node",
    "go":      "go",
    "npm":     "npm",
}

func IsCommandAvailable(name string) bool {
    _, err := CommandPath(name)
    return err == nil
}

func ToolDisplayName(name string) string {
    if display, ok := toolDisplayNames[name]; ok {
        return display
    }
    return name
}

func IsInstallableTool(name string) bool {
    for _, item := range installableTools {
        if item == name {
            return true
        }
    }
    return false
}

func SupportedTools() []string {
    out := make([]string, len(listedTools))
    copy(out, listedTools)
    return out
}

func InstallableTools() []string {
    out := make([]string, len(installableTools))
    copy(out, installableTools)
    return out
}

func CommandPath(name string) (string, error) {
    return executableLookup(name)
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

func IsBrewInstalled() bool {
    return IsCommandAvailable("brew")
}

func IsBrewFormulaInstalled(formula string) (bool, error) {
    output, err := commandRunner("brew", "list", "--formula", formula)
    if err != nil {
        text := strings.TrimSpace(string(output))
        if text == "" {
            return false, nil
        }
        if strings.Contains(text, "No such keg") {
            return false, nil
        }
        if strings.Contains(text, "No formula") {
            return false, nil
        }
    }

    for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
        if strings.TrimSpace(line) == formula {
            return true, nil
        }
    }
    return false, nil
}

func InstallNode() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("node") || IsCommandAvailable("npm") {
        return fmt.Errorf("node is already installed")
    }

    if _, err := commandRunner("brew", "install", "node@24"); err != nil {
        return fmt.Errorf("brew install node failed: %w", err)
    }

    return nil
}

func InstallPHP() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("php") {
        return fmt.Errorf("php is already installed")
    }

    if _, err := commandRunner("brew", "install", "php@8.4"); err != nil {
        return fmt.Errorf("brew install php failed: %w", err)
    }

    return nil
}

func InstallPython3() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    installed, err := IsBrewFormulaInstalled("python3")
    if err != nil {
        return fmt.Errorf("check python3 install status failed: %w", err)
    }
    if installed {
        return fmt.Errorf("python3 is already installed by homebrew")
    }

    if _, err := commandRunner("brew", "install", "python3"); err != nil {
        return fmt.Errorf("brew install python3 failed: %w", err)
    }

    return nil
}

func InstallGo() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("go") {
        return fmt.Errorf("go is already installed")
    }

    if _, err := commandRunner("brew", "install", "go"); err != nil {
        return fmt.Errorf("brew install go failed: %w", err)
    }

    return nil
}

func InstallTool(name string) error {
    switch name {
    case "node":
        return InstallNode()
    case "php":
        return InstallPHP()
    case "python3":
        return InstallPython3()
    case "go":
        return InstallGo()
    default:
        return fmt.Errorf("unsupported tool: %s", name)
    }
}
