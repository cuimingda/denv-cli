package cmd

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/denv"
)

type ToolRegistry interface {
	SupportedTools() []string
	InstallableTools() []string
	ResolveCommandPackages(name string) []denv.Package
	ToolDisplayName(name string) string
	IsInstallableTool(name string) bool
}

type ToolStateReader interface {
	IsCommandAvailable(name string) bool
	ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error)
	CommandPath(name string) (string, error)
	ResolvedBrewBinaryPath(name, formula string) (string, error)
	IsManagedByHomebrew(path string) bool
	IsBrewInstalled() bool
	IsBrewFormulaInstalled(formula string) (bool, error)
}

type ToolInstallPlanner interface {
	BuildInstallOperations(force bool) ([]denv.InstallOperation, error)
	BuildInstallQueue(force bool) (denv.InstallQueue, error)
	BuildInstallOperationsForTool(toolName string, force bool) ([]denv.InstallOperation, error)
	BuildInstallPlan(toolName string, options denv.BuildInstallPlanOptions) ([]denv.InstallOperation, error)
	BuildInstallQueueForTool(toolName string, force bool) (denv.InstallQueue, error)
}

type InstallPlanner interface {
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
	InstallToolWithOptions(name string, options denv.InstallOptions) error
}

type InstallExecutor interface {
	RunInstallOperation(out io.Writer, op denv.InstallOperation) error
	ExecuteInstallOperations(out io.Writer, operations []denv.InstallOperation) error
	ExecuteInstallQueue(out io.Writer, queue denv.InstallQueue) error
}

type InstallActions interface {
	InstallTool(name string) error
	InstallNode() error
	InstallNodeWithOutput(out io.Writer, force bool) error
	InstallPHP() error
	InstallPHPWithOutput(out io.Writer, force bool) error
	InstallPython3() error
	InstallPython3WithOutput(out io.Writer, force bool) error
	InstallGo() error
	InstallGoWithOutput(out io.Writer, force bool) error
	InstallCurl() error
	InstallCurlWithOutput(out io.Writer, force bool) error
	InstallGit() error
	InstallGitWithOutput(out io.Writer, force bool) error
	InstallFFmpeg() error
	InstallFFmpegWithOutput(out io.Writer, force bool) error
	InstallTree() error
	InstallTreeWithOutput(out io.Writer, force bool) error
	InstallGH() error
	InstallGHWithOutput(out io.Writer, force bool) error
}

type VersionService interface {
	ToolVersion(name string) (string, error)
	ToolVersionWithPath(name, commandPath string) (string, error)
	ToolVersionForOutdated(name string) (string, error)
	ToolLatestVersion(name string) (string, error)
	CompareVersions(current string, latest string) int
	ExtractVersion(out string) (string, error)
	SplitVersionParts(version string) []int
	ParseBrewStableVersion(output []byte) (string, error)
}

type UpdatePlanner interface {
	OutdatedItems() ([]denv.OutdatedItem, error)
	OutdatedUpdatePlan() ([]denv.OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}

type ListCommandService interface {
	ListToolItems(opts denv.ListOptions) ([]denv.ToolListItem, error)
}

type InstallCommandService interface {
	BuildInstallQueue(force bool) (denv.InstallQueue, error)
	ExecuteInstallQueue(out io.Writer, queue denv.InstallQueue) error
}

type OutdatedCommandService interface {
	SupportedTools() []string
	OutdatedItems() ([]denv.OutdatedItem, error)
}

type UpdateCommandService interface {
	SupportedTools() []string
	OutdatedUpdatePlan() ([]denv.OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}

type CommandService interface {
	ToolRegistry
	ToolStateReader
	VersionService
	InstallPlanner
	InstallExecutor
	UpdatePlanner
	ListCommandService
	InstallCommandService
	OutdatedCommandService
	UpdateCommandService
	ToolInstallPlanner
	InstallActions
}

type CLIContext struct {
	Service CommandService
}

func NewCLIContext() *CLIContext {
	return NewCLIContextWithRuntime(denv.Runtime{})
}

func NewCLIContextWithRuntime(rt denv.Runtime) *CLIContext {
	return &CLIContext{Service: denv.NewService(rt)}
}
