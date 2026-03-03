package cmd

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/denv"
)

// Command-layer contracts are intentionally narrow, aligned with service ports.
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

type CLIContext struct {
	service         *denv.Service
	Discovery       denv.Discovery
	InstallPlanner  denv.InstallPlanner
	InstallExecutor denv.InstallExecutor
	VersionResolver denv.VersionResolver
	OutdatedManager denv.OutdatedManager
	UpdateManager   denv.UpdateManager
}

func NewCLIContext() *CLIContext {
	return NewCLIContextWithRuntime(denv.Runtime{})
}

func NewCLIContextWithRuntime(rt denv.Runtime) *CLIContext {
	service := denv.NewService(rt)
	return &CLIContext{
		service:         service,
		Discovery:       service,
		InstallPlanner:  service,
		InstallExecutor: service,
		VersionResolver: service,
		OutdatedManager: service,
		UpdateManager:   service,
	}
}

func ensureCLIContext(ctx *CLIContext) *CLIContext {
	if ctx == nil {
		return NewCLIContext()
	}
	return ctx
}
