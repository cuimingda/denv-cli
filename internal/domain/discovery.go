package domain

import (
	"github.com/cuimingda/denv-cli/internal/infra"
)

type ServiceDiscoveryDeps[TRuntime any, TCatalog any, TPathPolicy any, TPackage any] struct {
	IsCommandAvailable    func(TRuntime, string) bool
	ToolInstallState      func(TRuntime, TCatalog, TPathPolicy, string) (bool, string, bool, error)
	ResolvedBrewBinaryPath func(TRuntime, string, string) (string, error)
	IsManagedByHomebrew   func(TPathPolicy, string) bool
	ToolDisplayName       func(TCatalog, string) string
	IsInstallableTool     func(TCatalog, string) bool
	ResolveCommandPackages func(TCatalog, string) []TPackage
	SupportedTools        func(TCatalog) []string
	InstallableTools      func(TCatalog) []string
	CommandPath           func(TRuntime, string) (string, error)
	IsBrewInstalled       func(TRuntime) bool
	IsBrewFormulaInstalled func(TRuntime, string) (bool, error)
}

type ServiceDiscovery[TRuntime any, TCatalog any, TPathPolicy any, TPackage any] struct {
	runtimeAdapter *infra.RuntimeAdapter[TRuntime]
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy]
	deps          ServiceDiscoveryDeps[TRuntime, TCatalog, TPathPolicy, TPackage]
}

func NewServiceDiscovery[TRuntime any, TCatalog any, TPathPolicy any, TPackage any](
	runtimeAdapter *infra.RuntimeAdapter[TRuntime],
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy],
	deps ServiceDiscoveryDeps[TRuntime, TCatalog, TPathPolicy, TPackage],
) *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage] {
	return &ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]{
		runtimeAdapter: runtimeAdapter,
		catalogManager: catalogManager,
		deps:           deps,
	}
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) runtimeRef() TRuntime {
	if d == nil || d.runtimeAdapter == nil {
		var zero TRuntime
		return zero
	}
	return d.runtimeAdapter.Runtime()
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) catalogRef() TCatalog {
	if d == nil || d.catalogManager == nil {
		var zero TCatalog
		return zero
	}
	return d.catalogManager.Catalog()
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) pathPolicyRef() TPathPolicy {
	if d == nil || d.catalogManager == nil {
		var zero TPathPolicy
		return zero
	}
	return d.catalogManager.PathPolicy()
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) IsCommandAvailable(name string) bool {
	return d.deps.IsCommandAvailable(d.runtimeRef(), name)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return d.deps.ToolInstallState(d.runtimeRef(), d.catalogRef(), d.pathPolicyRef(), name)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return d.deps.ResolvedBrewBinaryPath(d.runtimeRef(), name, formula)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) IsManagedByHomebrew(path string) bool {
	return d.deps.IsManagedByHomebrew(d.pathPolicyRef(), path)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) ToolDisplayName(name string) string {
	return d.deps.ToolDisplayName(d.catalogRef(), name)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) IsInstallableTool(name string) bool {
	return d.deps.IsInstallableTool(d.catalogRef(), name)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) ResolveCommandPackages(name string) []TPackage {
	return d.deps.ResolveCommandPackages(d.catalogRef(), name)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) SupportedTools() []string {
	return d.deps.SupportedTools(d.catalogRef())
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) InstallableTools() []string {
	return d.deps.InstallableTools(d.catalogRef())
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) CommandPath(name string) (string, error) {
	return d.deps.CommandPath(d.runtimeRef(), name)
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) IsBrewInstalled() bool {
	return d.deps.IsBrewInstalled(d.runtimeRef())
}

func (d *ServiceDiscovery[TRuntime, TCatalog, TPathPolicy, TPackage]) IsBrewFormulaInstalled(formula string) (bool, error) {
	return d.deps.IsBrewFormulaInstalled(d.runtimeRef(), formula)
}

