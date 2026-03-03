package cmd

import (
	"github.com/cuimingda/denv-cli/internal/denv"
)

// Command-layer contracts are intentionally narrow, aligned with service ports.
type ListCommandService interface {
	denv.VersionResolver
}

type InstallCommandService interface {
	denv.InstallPlanner
	denv.InstallExecutor
}

type OutdatedCommandService interface {
	denv.VersionResolver
	denv.Discovery
}

type UpdateCommandService interface {
	denv.UpdateManager
	denv.Discovery
}

type CLIContext struct {
	Service         *denv.Service
	Discovery       denv.Discovery
	InstallPlanner  denv.InstallPlanner
	InstallExecutor denv.InstallExecutor
	VersionResolver denv.VersionResolver
	UpdateManager   denv.UpdateManager
}

func NewCLIContext() *CLIContext {
	return NewCLIContextWithRuntime(denv.Runtime{})
}

func NewCLIContextWithRuntime(rt denv.Runtime) *CLIContext {
	service := denv.NewService(rt)
	return &CLIContext{
		Service:         service,
		Discovery:       service,
		InstallPlanner:  service,
		InstallExecutor: service,
		VersionResolver: service,
		UpdateManager:   service,
	}
}
