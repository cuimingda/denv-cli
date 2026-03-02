package cmd

import (
	"io"
	"os/exec"

	"github.com/cuimingda/denv-cli/internal/denv"
)

var (
	executableLookup = exec.LookPath
	commandRunner    = func(name string, args ...string) ([]byte, error) {
		return exec.Command(name, args...).CombinedOutput()
	}
	commandRunnerWithOutput = func(out io.Writer, name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Stdout = out
		cmd.Stderr = out
		return cmd.Run()
	}
)

type ToolService interface {
	SupportedTools() []string
	InstallableTools() []string
	IsCommandAvailable(name string) bool
	ToolInstallState(name string) (bool, string, bool, error)
	CommandPath(name string) (string, error)
	ToolVersion(name string) (string, error)
	ToolVersionWithPath(name, commandPath string) (string, error)
	ToolVersionForOutdated(name string) (string, error)
	ToolLatestVersion(name string) (string, error)
	ToolDisplayName(name string) string
	CompareVersions(current string, latest string) int
	IsManagedByHomebrew(path string) bool
	IsInstallableTool(name string) bool
	IsBrewInstalled() bool
	IsBrewFormulaInstalled(formula string) (bool, error)
	ResolvedBrewBinaryPath(name, formula string) (string, error)
	OutdatedItems() ([]denv.OutdatedItem, error)
	ListToolItems(opts denv.ListOptions) ([]denv.ToolListItem, error)
}

type OperationService interface {
	BuildInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildInstallOperationsForTool(toolName string, force bool) ([]denv.InstallOperation, error)
	BuildNodeInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildPHPInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildPython3InstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildGoInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildCurlInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildGitInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildFFmpegInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildTreeInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildGHInstallOperations(force bool) ([]denv.InstallOperation, error)
	RunInstallOperation(out io.Writer, op denv.InstallOperation) error
	UpdateToolWithOutput(out io.Writer, name string) error
	InstallNodeWithOutput(out io.Writer, force bool) error
	InstallPHPWithOutput(out io.Writer, force bool) error
	InstallPython3WithOutput(out io.Writer, force bool) error
	InstallGoWithOutput(out io.Writer, force bool) error
	InstallCurlWithOutput(out io.Writer, force bool) error
	InstallGitWithOutput(out io.Writer, force bool) error
	InstallFFmpegWithOutput(out io.Writer, force bool) error
	InstallTreeWithOutput(out io.Writer, force bool) error
	InstallGHWithOutput(out io.Writer, force bool) error
	InstallTool(name string) error
	InstallNode() error
	InstallPHP() error
	InstallPython3() error
	InstallGo() error
	InstallCurl() error
	InstallGit() error
	InstallFFmpeg() error
	InstallTree() error
	InstallGH() error
	OutdatedUpdatePlan() ([]denv.OutdatedItem, error)
	ExecuteInstallOperations(out io.Writer, operations []denv.InstallOperation) error
}

type CommandService interface {
	ToolService
	OperationService
}

type CLIContext struct {
	Runtime denv.Runtime
	Service CommandService
}

func NewCLIContext() *CLIContext {
	return NewCLIContextWithRuntime(commandRuntime())
}

func NewCLIContextWithRuntime(rt denv.Runtime) *CLIContext {
	normalized := denv.NormalizeRuntime(rt)
	return &CLIContext{
		Runtime: normalized,
		Service: denv.NewService(normalized),
	}
}

func commandRuntime() denv.Runtime {
	return denv.Runtime{
		ExecutableLookup:        executableLookup,
		CommandRunner:           commandRunner,
		CommandRunnerWithOutput: commandRunnerWithOutput,
	}
}
