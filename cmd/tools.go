package cmd

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/denv"
)

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
	ListCommandService
	InstallCommandService
	OutdatedCommandService
	UpdateCommandService
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
