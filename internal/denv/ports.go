package denv

import "io"

// RuntimeContext models runtime-facing capabilities for tool discovery and command/runtime probes.
type RuntimeContext interface {
	Discovery
}

// CatalogContext models tool lifecycle querying and outdated/version listing capabilities.
type CatalogContext interface {
	VersionResolver
	OutdatedManager
}

// ListContext models only listing behavior for command-style read paths.
type ListContext interface {
	VersionResolver
}

// InstallContext models install planning and execution capabilities.
type InstallContext interface {
	InstallPlanner
	InstallExecutor
}

// UpdateContext models outdated discovery and executable update orchestration.
type UpdateContext interface {
	OutdatedManager
	UpdateManager
}

// ServiceContext is the full root-capability surface the CLI commands can consume.
type ServiceContext interface {
	RuntimeContext
	CatalogContext
	InstallContext
	UpdateContext
}

// Discovery models read-only environment introspection for supported tools.
type Discovery interface {
	SupportedTools() []string
	InstallableTools() []string
	IsInstallableTool(name string) bool
	ToolDisplayName(name string) string
	IsCommandAvailable(name string) bool
	ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error)
	CommandPath(name string) (string, error)
	ResolveCommandPackages(name string) []Package
	ResolvedBrewBinaryPath(name string, formula string) (string, error)
	IsManagedByHomebrew(path string) bool
	IsBrewInstalled() bool
	IsBrewFormulaInstalled(formula string) (bool, error)
}

// InstallPlanner builds install plans from catalog and runtime state.
type InstallPlanner interface {
	BuildInstallQueue(force bool) (InstallQueue, error)
	BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error)
}

// InstallExecutor executes prepared install actions.
type InstallExecutor interface {
	ExecuteInstallQueue(out io.Writer, queue InstallQueue) error
}

// VersionResolver provides list/outdated view models.
type VersionResolver interface {
	ListToolItems(opts ListOptions) ([]ToolListItem, error)
	OutdatedItems() ([]OutdatedItem, error)
}

// OutdatedManager handles scan + filter operations for outdated checks.
type OutdatedManager interface {
	OutdatedChecks() ([]ToolCheckResult, error)
	OutdatedItems() ([]OutdatedItem, error)
	OutdatedUpdatePlan() ([]OutdatedItem, error)
}

// UpdateManager orchestrates outdated filtering and actual update actions.
type UpdateManager interface {
	OutdatedUpdatePlan() ([]OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}
