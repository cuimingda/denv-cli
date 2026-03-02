package cmd

import (
    "fmt"
    "io"
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
    "curl",
    "gh",
    "git",
    "ffmpeg",
    "tree",
}

var installableTools = []string{
    "php",
    "python3",
    "node",
    "go",
    "curl",
    "git",
}

var (
    executableLookup = exec.LookPath
    commandRunner   = func(name string, args ...string) ([]byte, error) {
        return exec.Command(name, args...).CombinedOutput()
    }
    commandRunnerWithOutput = func(out io.Writer, name string, args ...string) error {
        cmd := exec.Command(name, args...)
        cmd.Stdout = out
        cmd.Stderr = out
        return cmd.Run()
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
    "curl":    {"--version"},
    "gh":      {"--version"},
    "git":     {"--version"},
    "tree":    {"--version"},
}

var toolDisplayNames = map[string]string{
    "php":     "php",
    "python3": "python3",
    "node":    "node",
    "go":      "go",
    "npm":     "npm",
    "curl":    "curl",
    "gh":      "gh",
    "git":     "git",
    "ffmpeg":  "ffmpeg",
    "tree":    "tree",
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
    if name == "ffmpeg" {
        return toolVersionFromBrewList("ffmpeg")
    }

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

func toolVersionFromBrewList(formula string) (string, error) {
    output, err := commandRunner("brew", "info", formula)
    if err == nil {
        outputText := strings.TrimSpace(string(output))
        if outputText != "" {
            var fallbackVersion string
            for _, line := range strings.Split(outputText, "\n") {
                trimmed := strings.TrimSpace(line)
                if !strings.Contains(trimmed, "/opt/homebrew/Cellar/"+formula+"/") {
                    continue
                }

                fields := strings.Fields(trimmed)
                if len(fields) == 0 {
                    continue
                }
                if !strings.Contains(fields[0], "/opt/homebrew/Cellar/"+formula+"/") {
                    continue
                }

                path := fields[0]
                pathParts := strings.Split(path, "/")
                if len(pathParts) == 0 {
                    continue
                }
                candidate := pathParts[len(pathParts)-1]
                if candidate == "" {
                    continue
                }
                fallbackVersion = candidate

                if len(fields) >= 2 && fields[len(fields)-1] == "*" {
                    return candidate, nil
                }
            }
            if fallbackVersion != "" {
                return fallbackVersion, nil
            }
        }
    }

    output, err = commandRunner("brew", "info", "--json=v2", formula)
    if err != nil {
        if formulaVersion, versionErr := extractVersion(string(output)); versionErr == nil {
            return formulaVersion, nil
        }
        return "", fmt.Errorf("brew info failed: %w", err)
    }

    return parseBrewStableVersion(output)
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

    if _, err := commandRunner("brew", "install", "node"); err != nil {
        return fmt.Errorf("brew install node failed: %w", err)
    }

    return nil
}

func InstallNodeWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && (IsCommandAvailable("node") || IsCommandAvailable("npm")) {
        return fmt.Errorf("node is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "node"); err != nil {
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

    if _, err := commandRunner("brew", "install", "php"); err != nil {
        return fmt.Errorf("brew install php failed: %w", err)
    }

    return nil
}

func InstallPHPWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("php") {
        return fmt.Errorf("php is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "php"); err != nil {
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

func InstallPython3WithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force {
        installed, err := IsBrewFormulaInstalled("python3")
        if err != nil {
            return fmt.Errorf("check python3 install status failed: %w", err)
        }
        if installed {
            return fmt.Errorf("python3 is already installed by homebrew")
        }
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "python3"); err != nil {
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

func InstallGoWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("go") {
        return fmt.Errorf("go is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "go"); err != nil {
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
    case "curl":
        return InstallCurl()
    case "git":
        return InstallGit()
    default:
        return fmt.Errorf("unsupported tool: %s", name)
    }
}

func InstallCurl() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("curl") {
        return fmt.Errorf("curl is already installed")
    }

    if _, err := commandRunner("brew", "install", "curl"); err != nil {
        return fmt.Errorf("brew install curl failed: %w", err)
    }

    if _, err := commandRunner("brew", "link", "curl", "--force"); err != nil {
        return fmt.Errorf("brew link curl failed: %w", err)
    }

    return nil
}

func InstallCurlWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("curl") {
        return fmt.Errorf("curl is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "curl"); err != nil {
        return fmt.Errorf("brew install curl failed: %w", err)
    }

    if err := commandRunnerWithOutput(out, "brew", "link", "curl", "--force"); err != nil {
        return fmt.Errorf("brew link curl failed: %w", err)
    }

    return nil
}

func InstallGit() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("git") {
        return fmt.Errorf("git is already installed")
    }

    if _, err := commandRunner("brew", "install", "git"); err != nil {
        return fmt.Errorf("brew install git failed: %w", err)
    }

    return nil
}

func InstallGitWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("git") {
        return fmt.Errorf("git is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "git"); err != nil {
        return fmt.Errorf("brew install git failed: %w", err)
    }

    return nil
}
