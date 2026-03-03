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

type ToolRegistry interface {
	SupportedTools() []string
	InstallableTools() []string
	ToolDisplayName(name string) string
	CompareVersions(current string, latest string) int
	IsInstallableTool(name string) bool
}

type ToolService interface {
	ToolRegistry
	ToolStateProbe
	InstallPlanner
	ExecutionEngine
	UpdatePolicy
}

type ToolStateProbe interface {
	OutdatedItems() ([]denv.OutdatedItem, error)
	ListToolItems(opts denv.ListOptions) ([]denv.ToolListItem, error)
	ToolInstallState(name string) (bool, string, bool, error)
	ToolVersion(name string) (string, error)
	ToolVersionWithPath(name, commandPath string) (string, error)
	ToolVersionForOutdated(name string) (string, error)
	ToolLatestVersion(name string) (string, error)
	IsCommandAvailable(name string) bool
	CommandPath(name string) (string, error)
	IsManagedByHomebrew(path string) bool
	IsBrewInstalled() bool
	IsBrewFormulaInstalled(formula string) (bool, error)
	ResolvedBrewBinaryPath(name, formula string) (string, error)
}

type InstallPlanner interface {
	BuildInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildInstallOperationsForTool(toolName string, force bool) ([]denv.InstallOperation, error)
	BuildInstallQueue(force bool) (denv.InstallQueue, error)
	BuildInstallQueueForTool(toolName string, force bool) (denv.InstallQueue, error)
	BuildNodeInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildNodeInstallQueue(force bool) (denv.InstallQueue, error)
	BuildPHPInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildPHPInstallQueue(force bool) (denv.InstallQueue, error)
	BuildPython3InstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildPython3InstallQueue(force bool) (denv.InstallQueue, error)
	BuildGoInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildGoInstallQueue(force bool) (denv.InstallQueue, error)
	BuildCurlInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildCurlInstallQueue(force bool) (denv.InstallQueue, error)
	BuildGitInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildGitInstallQueue(force bool) (denv.InstallQueue, error)
	BuildFFmpegInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildFFmpegInstallQueue(force bool) (denv.InstallQueue, error)
	BuildTreeInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildTreeInstallQueue(force bool) (denv.InstallQueue, error)
	BuildGHInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildGHInstallQueue(force bool) (denv.InstallQueue, error)
}

type UpdatePolicy interface {
	OutdatedUpdatePlan() ([]denv.OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}

type ExecutionEngine interface {
	RunInstallOperation(out io.Writer, op denv.InstallOperation) error
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
	ExecuteInstallOperations(out io.Writer, operations []denv.InstallOperation) error
	ExecuteInstallQueue(out io.Writer, queue denv.InstallQueue) error
}

type CommandService interface {
	ToolRegistry
	ToolStateProbe
	InstallPlanner
	ExecutionEngine
	UpdatePolicy
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
