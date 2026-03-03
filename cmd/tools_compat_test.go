package cmd

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/interval"
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

func InstallNode() error { return denvService().InstallTool("node") }
func InstallNodeWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("node", denv.InstallOptions{Force: force, Output: out})
}

func InstallPHP() error { return denvService().InstallTool("php") }
func InstallPHPWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("php", denv.InstallOptions{Force: force, Output: out})
}

func InstallPython3() error { return denvService().InstallTool("python3") }
func InstallPython3WithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("python3", denv.InstallOptions{Force: force, Output: out})
}

func InstallGo() error { return denvService().InstallTool("go") }
func InstallGoWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("go", denv.InstallOptions{Force: force, Output: out})
}

func InstallCurl() error { return denvService().InstallTool("curl") }
func InstallCurlWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("curl", denv.InstallOptions{Force: force, Output: out})
}

func InstallGit() error { return denvService().InstallTool("git") }
func InstallGitWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("git", denv.InstallOptions{Force: force, Output: out})
}

func InstallFFmpeg() error { return denvService().InstallTool("ffmpeg") }
func InstallFFmpegWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("ffmpeg", denv.InstallOptions{Force: force, Output: out})
}

func InstallTree() error { return denvService().InstallTool("tree") }
func InstallTreeWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("tree", denv.InstallOptions{Force: force, Output: out})
}

func InstallGH() error { return denvService().InstallTool("gh") }
func InstallGHWithOutput(out io.Writer, force bool) error {
	return denvService().InstallToolWithOptions("gh", denv.InstallOptions{Force: force, Output: out})
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
	ops, err := denvService().BuildInstallOperationsForTool("node", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildNodeInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("node", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildPHPInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("php", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildPHPInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("php", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildPython3InstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("python3", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildPython3InstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("python3", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildGoInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("go", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildGoInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("go", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildCurlInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("curl", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildCurlInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("curl", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildGitInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("git", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildGitInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("git", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildFFmpegInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("ffmpeg", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildFFmpegInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("ffmpeg", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildTreeInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("tree", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildTreeInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("tree", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
}

func buildGHInstallOperations(force bool) ([]string, error) {
	ops, err := denvService().BuildInstallOperationsForTool("gh", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(ops), nil
}

func buildGHInstallQueue(force bool) ([]string, error) {
	queue, err := denvService().BuildInstallQueueForTool("gh", force)
	if err != nil {
		return nil, err
	}
	return denvOperationsToStrings(queue.ToOperations()), nil
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
