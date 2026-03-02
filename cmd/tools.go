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

func denvService() *denv.Service {
	return denv.NewService(toolRuntime())
}

func IsCommandAvailable(name string) bool {
	return denvService().IsCommandAvailable(name)
}

func ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return denvService().ToolInstallState(name)
}

func ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return denvService().ResolvedBrewBinaryPath(name, formula)
}

func ToolDisplayName(name string) string {
	return denvService().ToolDisplayName(name)
}

func IsInstallableTool(name string) bool {
	return denvService().IsInstallableTool(name)
}

func SupportedTools() []string {
	return denvService().SupportedTools()
}

func InstallableTools() []string {
	return denvService().InstallableTools()
}

func CommandPath(name string) (string, error) {
	return denvService().CommandPath(name)
}

func ToolVersion(name string) (string, error) {
	return denvService().ToolVersion(name)
}

func ToolVersionWithPath(name, commandPath string) (string, error) {
	return denvService().ToolVersionWithPath(name, commandPath)
}

func ToolVersionForOutdated(name string) (string, error) {
	return denvService().ToolVersionForOutdated(name)
}

func extractVersion(out string) (string, error) {
	return denvService().ExtractVersion(out)
}

func splitVersionParts(version string) []int {
	return denvService().SplitVersionParts(version)
}

func IsBrewInstalled() bool {
	return denvService().IsBrewInstalled()
}

func IsBrewFormulaInstalled(formula string) (bool, error) {
	return denvService().IsBrewFormulaInstalled(formula)
}

func InstallNode() error { return denvService().InstallNode() }
func InstallNodeWithOutput(out io.Writer, force bool) error { return denvService().InstallNodeWithOutput(out, force) }

func InstallPHP() error { return denvService().InstallPHP() }
func InstallPHPWithOutput(out io.Writer, force bool) error { return denvService().InstallPHPWithOutput(out, force) }

func InstallPython3() error { return denvService().InstallPython3() }
func InstallPython3WithOutput(out io.Writer, force bool) error { return denvService().InstallPython3WithOutput(out, force) }

func InstallGo() error { return denvService().InstallGo() }
func InstallGoWithOutput(out io.Writer, force bool) error { return denvService().InstallGoWithOutput(out, force) }

func InstallCurl() error { return denvService().InstallCurl() }
func InstallCurlWithOutput(out io.Writer, force bool) error { return denvService().InstallCurlWithOutput(out, force) }

func InstallGit() error { return denvService().InstallGit() }
func InstallGitWithOutput(out io.Writer, force bool) error { return denvService().InstallGitWithOutput(out, force) }

func InstallFFmpeg() error { return denvService().InstallFFmpeg() }
func InstallFFmpegWithOutput(out io.Writer, force bool) error { return denvService().InstallFFmpegWithOutput(out, force) }

func InstallTree() error { return denvService().InstallTree() }
func InstallTreeWithOutput(out io.Writer, force bool) error { return denvService().InstallTreeWithOutput(out, force) }

func InstallGH() error { return denvService().InstallGH() }
func InstallGHWithOutput(out io.Writer, force bool) error { return denvService().InstallGHWithOutput(out, force) }

func InstallTool(name string) error { return denvService().InstallTool(name) }

func UpdateToolWithOutput(out io.Writer, name string) error {
	return denvService().UpdateToolWithOutput(out, name)
}

func ToolLatestVersion(name string) (string, error) {
	return denvService().ToolLatestVersion(name)
}

func cmpVersions(current string, latest string) int {
	return denvService().CompareVersions(current, latest)
}

func parseBrewStableVersion(output []byte) (string, error) {
	return denvService().ParseBrewStableVersion(output)
}

func resolvedBrewBinaryPath(name, formula string) (string, error) {
	return denvService().ResolvedBrewBinaryPath(name, formula)
}
