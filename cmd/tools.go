package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
)

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

    brewPath, pathErr := resolvedBrewBinaryPath(name, formula)
    if pathErr != nil {
        return true, fmt.Sprintf("/opt/homebrew/bin/%s", name), true, nil
    }

    if brewPath != "" {
        return true, brewPath, true, nil
    }

    return true, fmt.Sprintf("/opt/homebrew/bin/%s", name), true, nil
}

func resolvedBrewBinaryPath(name, formula string) (string, error) {
    output, err := commandRunner("brew", "--prefix", formula)
    if err != nil {
        return "", err
    }

    prefix := strings.TrimSpace(string(output))
    if prefix == "" {
        return "", nil
    }

    return fmt.Sprintf("%s/bin/%s", prefix, name), nil
}

func ToolDisplayName(name string) string {
	if display, ok := toolDisplayName(name); ok {
		return display
	}
	return name
}

func IsInstallableTool(name string) bool {
	return managedToolIsInstallable(name)
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
	cmdArgs, ok := versionArgsForTool(name)
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
	cmdArgs, ok := versionArgsForTool(name)
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
    return installToolWithoutOutput("node", false)
}

func InstallNodeWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("node", force, out)
}

func InstallPHP() error {
    return installToolWithoutOutput("php", false)
}

func InstallPHPWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("php", force, out)
}

func InstallPython3() error {
    return installToolWithoutOutput("python3", false)
}

func InstallPython3WithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("python3", force, out)
}

func InstallGo() error {
    return installToolWithoutOutput("go", false)
}

func InstallGoWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("go", force, out)
}

func InstallCurl() error {
    return installToolWithoutOutput("curl", false)
}

func InstallCurlWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("curl", force, out)
}

func InstallGit() error {
    return installToolWithoutOutput("git", false)
}

func InstallGitWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("git", force, out)
}

func InstallFFmpeg() error {
    return installToolWithoutOutput("ffmpeg", false)
}

func InstallFFmpegWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("ffmpeg", force, out)
}

func InstallTree() error {
    return installToolWithoutOutput("tree", false)
}

func InstallTreeWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("tree", force, out)
}

func InstallGH() error {
    return installToolWithoutOutput("gh", false)
}

func InstallGHWithOutput(out io.Writer, force bool) error {
    return installToolWithOutput("gh", force, out)
}

func InstallTool(name string) error {
    if !managedToolIsInstallable(name) {
        return fmt.Errorf("unsupported tool: %s", name)
    }
    return installToolWithoutOutput(name, false)
}

func buildInstallPlan(toolName string, force bool, checkCommand bool, checkFormula bool) ([]string, error) {
    if !IsBrewInstalled() {
        return nil, fmt.Errorf("homebrew is not installed")
    }

    if !force && checkCommand && isToolPresent(toolName) {
        return nil, nil
    }

    if !force && checkFormula {
        installedByFormula, err := IsBrewFormulaInstalled(toolName)
        if err != nil {
            return nil, fmt.Errorf("check %s install status failed: %w", toolName, err)
        }
        if installedByFormula {
            return nil, nil
        }
    }

    operations, ok := installOperationSequence(toolName)
    if !ok {
        return nil, fmt.Errorf("unsupported tool: %s", toolName)
    }

	return operations, nil
}

func isToolPresent(name string) bool {
	if name == "node" {
		return IsCommandAvailable("node") || IsCommandAvailable("npm")
	}
	return IsCommandAvailable(name)
}

func installToolWithoutOutput(name string, force bool) error {
    return installTool(name, force, nil, false)
}

func installToolWithOutput(name string, force bool, out io.Writer) error {
    return installTool(name, force, out, true)
}

func installTool(name string, force bool, out io.Writer, withOutput bool) error {
    spec, ok := managedToolFor(name)
    if !ok || !spec.installable {
        return fmt.Errorf("unsupported tool: %s", name)
    }

    operations, err := buildInstallPlan(name, force, spec.runCheckCommand, spec.runCheckFormula)
    if err != nil {
        return err
    }

    if len(operations) == 0 {
        return fmt.Errorf(spec.alreadyInstalledLabel)
    }

    for _, operation := range operations {
        if withOutput {
            if err := runInstallOperation(out, operation); err != nil {
                return fmt.Errorf("%s failed: %w", operation, err)
            }
            continue
        }

        if err := runInstallOperationCommand(operation); err != nil {
            return fmt.Errorf("%s failed: %w", operation, err)
        }
    }

    return nil
}

func runInstallOperationCommand(op string) error {
    args := strings.Fields(op)
    if len(args) == 0 {
        return nil
    }
    _, err := commandRunner(args[0], args[1:]...)
    return err
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
