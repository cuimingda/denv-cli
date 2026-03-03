package cmd

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/denv"
)

func denvService() *denv.Service {
	return testCommandService()
}

func denvConcreteService() *denv.Service {
	return denvService()
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
	return denvConcreteService().ExtractVersion(out)
}

func splitVersionParts(version string) []int {
	return denvConcreteService().SplitVersionParts(version)
}

func IsBrewInstalled() bool {
	return denvService().IsBrewInstalled()
}

func IsBrewFormulaInstalled(formula string) (bool, error) {
	return denvService().IsBrewFormulaInstalled(formula)
}

func InstallNode() error { return denvService().InstallNode() }
func InstallNodeWithOutput(out io.Writer, force bool) error {
	return denvService().InstallNodeWithOutput(out, force)
}

func InstallPHP() error { return denvService().InstallPHP() }
func InstallPHPWithOutput(out io.Writer, force bool) error {
	return denvService().InstallPHPWithOutput(out, force)
}

func InstallPython3() error { return denvService().InstallPython3() }
func InstallPython3WithOutput(out io.Writer, force bool) error {
	return denvService().InstallPython3WithOutput(out, force)
}

func InstallGo() error { return denvService().InstallGo() }
func InstallGoWithOutput(out io.Writer, force bool) error {
	return denvService().InstallGoWithOutput(out, force)
}

func InstallCurl() error { return denvService().InstallCurl() }
func InstallCurlWithOutput(out io.Writer, force bool) error {
	return denvService().InstallCurlWithOutput(out, force)
}

func InstallGit() error { return denvService().InstallGit() }
func InstallGitWithOutput(out io.Writer, force bool) error {
	return denvService().InstallGitWithOutput(out, force)
}

func InstallFFmpeg() error { return denvService().InstallFFmpeg() }
func InstallFFmpegWithOutput(out io.Writer, force bool) error {
	return denvService().InstallFFmpegWithOutput(out, force)
}

func InstallTree() error { return denvService().InstallTree() }
func InstallTreeWithOutput(out io.Writer, force bool) error {
	return denvService().InstallTreeWithOutput(out, force)
}

func InstallGH() error { return denvService().InstallGH() }
func InstallGHWithOutput(out io.Writer, force bool) error {
	return denvService().InstallGHWithOutput(out, force)
}

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
	return denvConcreteService().ParseBrewStableVersion(output)
}

func resolvedBrewBinaryPath(name, formula string) (string, error) {
	return denvService().ResolvedBrewBinaryPath(name, formula)
}

func buildInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildInstallOperationsForTool(toolName string, force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool(toolName, force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildInstallQueueForTool(toolName string, force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool(toolName, force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildNodeInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildNodeInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildNodeInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildNodeInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildPHPInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildPHPInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildPHPInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildPHPInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildPython3InstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildPython3InstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildPython3InstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildPython3InstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildGoInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildGoInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildGoInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildGoInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildCurlInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildCurlInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildCurlInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildCurlInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildGitInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildGitInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildGitInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildGitInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildFFmpegInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildFFmpegInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildFFmpegInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildFFmpegInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildTreeInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildTreeInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildTreeInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildTreeInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func buildGHInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildGHInstallOperations(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildGHInstallQueue(force bool) ([]string, error) {
	ops, err := denvService().BuildGHInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops.ToOperations()), nil
}

func runInstallOperation(out io.Writer, op string) error {
	command, err := denv.ParseCommandSpec(op)
	if err != nil {
		return err
	}
	return denvService().RunInstallOperation(out, denv.InstallOperation{Spec: command})
}

func denvOperationsToStrings(ops []denv.InstallOperation) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		out = append(out, op.String())
	}
	return out
}
