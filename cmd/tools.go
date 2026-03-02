package cmd

import (
    "encoding/json"
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
    "ffmpeg",
    "tree",
    "gh",
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
    "ffmpeg":  {"--version"},
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

func ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
    commandPath, err = CommandPath(name)
    if err == nil {
        return true, commandPath, isHomebrewPath(commandPath), nil
    }

    formula, ok := brewFormulaForTool(name)
    if !ok {
        return false, "", false, nil
    }

    installedByBrew, lookupErr := IsBrewFormulaInstalled(formula)
    if lookupErr != nil {
        return false, "", false, lookupErr
    }

    if !installedByBrew {
        return false, "", false, nil
    }

    return true, fmt.Sprintf("/opt/homebrew/bin/%s", name), true, nil
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
    return toolVersionFromCommand(name)
}

func ToolVersionWithPath(name, commandPath string) (string, error) {
    version, err := toolVersionFromCommand(name)
    if err == nil {
        return version, nil
    }

    if commandPath == "" || commandPath == name {
        return "", err
    }

    return toolVersionFromCommandPath(commandPath, name)
}

func ToolVersionForOutdated(name string) (string, error) {
    if name == "npm" {
        return toolVersionFromCommand("npm")
    }

    formula, ok := brewFormulaForTool(name)
    if !ok {
        return toolVersionFromCommand(name)
    }

    return toolVersionFromBrewList(formula)
}

func toolVersionFromCommand(name string) (string, error) {
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

func toolVersionFromCommandPath(commandPath, name string) (string, error) {
    cmdArgs, ok := toolVersionCommands[name]
    if !ok {
        return "", fmt.Errorf("unsupported tool: %s", name)
    }

    output, err := commandRunner(commandPath, cmdArgs...)
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

    type formulaInfo struct {
        Name      string `json:"name"`
        Versions  struct {
            Stable string `json:"stable"`
        } `json:"versions"`
        Revision  int `json:"revision"`
        Installed []struct {
            Version string `json:"version"`
        } `json:"installed"`
    }

    var payload struct {
        Formulae []formulaInfo `json:"formulae"`
    }

    if err := json.Unmarshal(output, &payload); err == nil {
        if len(payload.Formulae) > 0 {
            installed := payload.Formulae[0].Installed
            for i := len(installed) - 1; i >= 0; i-- {
                if strings.TrimSpace(installed[i].Version) != "" {
                    return installed[i].Version, nil
                }
            }

            stable := payload.Formulae[0].Versions.Stable
            if payload.Formulae[0].Revision > 0 {
                return fmt.Sprintf("%s_%d", stable, payload.Formulae[0].Revision), nil
            }
            if stable != "" {
                return stable, nil
            }
        }
    }

    return "", fmt.Errorf("failed to parse brew installed version")
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
    case "ffmpeg":
        return InstallFFmpeg()
    case "tree":
        return InstallTree()
    case "gh":
        return InstallGH()
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

func InstallFFmpeg() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("ffmpeg") {
        return fmt.Errorf("ffmpeg is already installed")
    }

    if _, err := commandRunner("brew", "install", "ffmpeg"); err != nil {
        return fmt.Errorf("brew install ffmpeg failed: %w", err)
    }

    return nil
}

func InstallFFmpegWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("ffmpeg") {
        return fmt.Errorf("ffmpeg is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "ffmpeg"); err != nil {
        return fmt.Errorf("brew install ffmpeg failed: %w", err)
    }

    return nil
}

func InstallTree() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("tree") {
        return fmt.Errorf("tree is already installed")
    }

    if _, err := commandRunner("brew", "install", "tree"); err != nil {
        return fmt.Errorf("brew install tree failed: %w", err)
    }

    return nil
}

func InstallTreeWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("tree") {
        return fmt.Errorf("tree is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "tree"); err != nil {
        return fmt.Errorf("brew install tree failed: %w", err)
    }

    return nil
}

func InstallGH() error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if IsCommandAvailable("gh") {
        return fmt.Errorf("gh is already installed")
    }

    if _, err := commandRunner("brew", "install", "gh"); err != nil {
        return fmt.Errorf("brew install gh failed: %w", err)
    }

    return nil
}

func InstallGHWithOutput(out io.Writer, force bool) error {
    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
    }

    if !force && IsCommandAvailable("gh") {
        return fmt.Errorf("gh is already installed")
    }

    if err := commandRunnerWithOutput(out, "brew", "install", "gh"); err != nil {
        return fmt.Errorf("brew install gh failed: %w", err)
    }

    return nil
}

func UpdateToolWithOutput(out io.Writer, name string) error {
    installed, _, _, err := ToolInstallState(name)
    if err != nil {
        return err
    }
    if !installed {
        return fmt.Errorf("tool %s is not installed", name)
    }

    if name == "npm" {
        if err := commandRunnerWithOutput(out, "npm", "install", "-g", "npm@latest"); err != nil {
            return fmt.Errorf("npm update failed: %w", err)
        }
        return nil
    }

    if !IsBrewInstalled() {
        return fmt.Errorf("homebrew is not installed")
   }

    formula, ok := brewFormulaForTool(name)
    if !ok {
        return fmt.Errorf("unsupported tool: %s", name)
    }

    if err := commandRunnerWithOutput(out, "brew", "upgrade", formula); err != nil {
        return fmt.Errorf("brew upgrade %s failed: %w", formula, err)
    }
    return nil
}
