package domain

import (
	"github.com/cuimingda/denv-cli/internal/infra"
)

type ServiceVersionResolverDeps[TRuntime any, TCatalog any, TPathPolicy any, TListOptions any, TToolListItem any] struct {
	ListToolItems          func(TRuntime, TCatalog, TPathPolicy, TListOptions) ([]TToolListItem, error)
	ToolVersion            func(TRuntime, TCatalog, string) (string, error)
	ToolVersionWithPath    func(TRuntime, TCatalog, string, string) (string, error)
	ToolVersionForOutdated func(TRuntime, TCatalog, string) (string, error)
	ToolLatestVersion      func(TRuntime, TCatalog, string) (string, error)
}

type ServiceVersionResolver[TRuntime any, TCatalog any, TPathPolicy any, TListOptions any, TToolListItem any] struct {
	runtimeAdapter *infra.RuntimeAdapter[TRuntime]
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy]
	deps           ServiceVersionResolverDeps[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]
}

func NewServiceVersionResolver[TRuntime any, TCatalog any, TPathPolicy any, TListOptions any, TToolListItem any](
	runtimeAdapter *infra.RuntimeAdapter[TRuntime],
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy],
	deps ServiceVersionResolverDeps[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem],
) *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem] {
	return &ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]{
		runtimeAdapter: runtimeAdapter,
		catalogManager: catalogManager,
		deps:           deps,
	}
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) runtimeRef() TRuntime {
	if v == nil || v.runtimeAdapter == nil {
		var zero TRuntime
		return zero
	}
	return v.runtimeAdapter.Runtime()
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) catalogRef() TCatalog {
	if v == nil || v.catalogManager == nil {
		var zero TCatalog
		return zero
	}
	return v.catalogManager.Catalog()
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) pathPolicyRef() TPathPolicy {
	if v == nil || v.catalogManager == nil {
		var zero TPathPolicy
		return zero
	}
	return v.catalogManager.PathPolicy()
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) ListToolItems(opts TListOptions) ([]TToolListItem, error) {
	return v.deps.ListToolItems(v.runtimeRef(), v.catalogRef(), v.pathPolicyRef(), opts)
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) ToolVersion(name string) (string, error) {
	return v.deps.ToolVersion(v.runtimeRef(), v.catalogRef(), name)
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) ToolVersionWithPath(name, commandPath string) (string, error) {
	return v.deps.ToolVersionWithPath(v.runtimeRef(), v.catalogRef(), name, commandPath)
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) ToolVersionForOutdated(name string) (string, error) {
	return v.deps.ToolVersionForOutdated(v.runtimeRef(), v.catalogRef(), name)
}

func (v *ServiceVersionResolver[TRuntime, TCatalog, TPathPolicy, TListOptions, TToolListItem]) ToolLatestVersion(name string) (string, error) {
	return v.deps.ToolLatestVersion(v.runtimeRef(), v.catalogRef(), name)
}
