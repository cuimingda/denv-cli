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
	OutdatedChecks() ([]denv.ToolCheckResult, error)
}

type UpdateCommandService interface {
	SupportedTools() []string
	OutdatedUpdatePlan() ([]denv.OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}

type CLIContext struct {
	service         *denv.Service
	RuntimeContext  denv.RuntimeContext
	CatalogContext  denv.CatalogContext
	InstallContext  denv.InstallContext
	UpdateContext   denv.UpdateContext
}

func NewCLIContext() *CLIContext {
	return NewCLIContextWithRuntime(denv.Runtime{})
}

func NewCLIContextWithRuntime(rt denv.Runtime) *CLIContext {
	service := denv.NewService(rt)
	return &CLIContext{
		service:         service,
		RuntimeContext:  service,
		CatalogContext:  service,
		InstallContext:  service,
		UpdateContext:   service,
	}
}

func ensureCLIContext(ctx *CLIContext) *CLIContext {
	if ctx == nil {
		return NewCLIContext()
	}
	return ctx
}
