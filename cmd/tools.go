package cmd

import (
	"io"
	"os/exec"

	"github.com/cuimingda/denv-cli/internal/denv"
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
)

func toolRuntime() denv.Runtime {
	return denv.Runtime{
		ExecutableLookup:        executableLookup,
		CommandRunner:           commandRunner,
		CommandRunnerWithOutput: commandRunnerWithOutput,
	}
}

func IsCommandAvailable(name string) bool {
	return denv.IsCommandAvailable(toolRuntime(), name)
}

func ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return denv.ToolInstallState(toolRuntime(), name)
}

func ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return denv.ResolvedBrewBinaryPath(toolRuntime(), name, formula)
}

func ToolDisplayName(name string) string {
	return denv.ToolDisplayName(name)
}

func IsInstallableTool(name string) bool {
	return denv.IsInstallableTool(name)
}

func SupportedTools() []string {
	return denv.SupportedTools()
}

func InstallableTools() []string {
	return denv.InstallableTools()
}

func CommandPath(name string) (string, error) {
	return denv.CommandPath(toolRuntime(), name)
}

func ToolVersion(name string) (string, error) {
	return denv.ToolVersion(toolRuntime(), name)
}

func ToolVersionWithPath(name, commandPath string) (string, error) {
	return denv.ToolVersionWithPath(toolRuntime(), name, commandPath)
}

func ToolVersionForOutdated(name string) (string, error) {
	return denv.ToolVersionForOutdated(toolRuntime(), name)
}

func extractVersion(out string) (string, error) {
	return denv.ExtractVersion(out)
}

func splitVersionParts(version string) []int {
	return denv.SplitVersionParts(version)
}

func IsBrewInstalled() bool {
	return denv.IsBrewInstalled(toolRuntime())
}

func IsBrewFormulaInstalled(formula string) (bool, error) {
	return denv.IsBrewFormulaInstalled(toolRuntime(), formula)
}

func InstallNode() error { return denv.InstallNode(toolRuntime()) }
func InstallNodeWithOutput(out io.Writer, force bool) error { return denv.InstallNodeWithOutput(toolRuntime(), out, force) }

func InstallPHP() error { return denv.InstallPHP(toolRuntime()) }
func InstallPHPWithOutput(out io.Writer, force bool) error { return denv.InstallPHPWithOutput(toolRuntime(), out, force) }

func InstallPython3() error { return denv.InstallPython3(toolRuntime()) }
func InstallPython3WithOutput(out io.Writer, force bool) error { return denv.InstallPython3WithOutput(toolRuntime(), out, force) }

func InstallGo() error { return denv.InstallGo(toolRuntime()) }
func InstallGoWithOutput(out io.Writer, force bool) error { return denv.InstallGoWithOutput(toolRuntime(), out, force) }

func InstallCurl() error { return denv.InstallCurl(toolRuntime()) }
func InstallCurlWithOutput(out io.Writer, force bool) error { return denv.InstallCurlWithOutput(toolRuntime(), out, force) }

func InstallGit() error { return denv.InstallGit(toolRuntime()) }
func InstallGitWithOutput(out io.Writer, force bool) error { return denv.InstallGitWithOutput(toolRuntime(), out, force) }

func InstallFFmpeg() error { return denv.InstallFFmpeg(toolRuntime()) }
func InstallFFmpegWithOutput(out io.Writer, force bool) error { return denv.InstallFFmpegWithOutput(toolRuntime(), out, force) }

func InstallTree() error { return denv.InstallTree(toolRuntime()) }
func InstallTreeWithOutput(out io.Writer, force bool) error { return denv.InstallTreeWithOutput(toolRuntime(), out, force) }

func InstallGH() error { return denv.InstallGH(toolRuntime()) }
func InstallGHWithOutput(out io.Writer, force bool) error { return denv.InstallGHWithOutput(toolRuntime(), out, force) }

func InstallTool(name string) error { return denv.InstallTool(toolRuntime(), name) }

func UpdateToolWithOutput(out io.Writer, name string) error {
	return denv.UpdateToolWithOutput(toolRuntime(), out, name)
}

func ToolLatestVersion(name string) (string, error) {
	return denv.ToolLatestVersion(toolRuntime(), name)
}

func cmpVersions(current string, latest string) int {
	return denv.CompareVersions(current, latest)
}

func parseBrewStableVersion(output []byte) (string, error) {
	return denv.ParseBrewStableVersion(output)
}

func resolvedBrewBinaryPath(name, formula string) (string, error) {
	return ResolvedBrewBinaryPath(name, formula)
}
