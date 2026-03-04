package domain

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/infra"
)

type ServiceUpdateServiceDeps[TRuntime any, TCatalog any] struct {
	UpdateToolWithOutput func(TRuntime, TCatalog, io.Writer, string) error
}

type ServiceUpdateService[TRuntime any, TCatalog any, TPathPolicy any] struct {
	runtimeAdapter *infra.RuntimeAdapter[TRuntime]
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy]
	deps           ServiceUpdateServiceDeps[TRuntime, TCatalog]
}

func NewServiceUpdateService[TRuntime any, TCatalog any, TPathPolicy any](
	runtimeAdapter *infra.RuntimeAdapter[TRuntime],
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy],
	deps ServiceUpdateServiceDeps[TRuntime, TCatalog],
) *ServiceUpdateService[TRuntime, TCatalog, TPathPolicy] {
	return &ServiceUpdateService[TRuntime, TCatalog, TPathPolicy]{
		runtimeAdapter: runtimeAdapter,
		catalogManager: catalogManager,
		deps:           deps,
	}
}

func (u *ServiceUpdateService[TRuntime, TCatalog, TPathPolicy]) runtimeRef() TRuntime {
	if u == nil || u.runtimeAdapter == nil {
		var zero TRuntime
		return zero
	}
	return u.runtimeAdapter.Runtime()
}

func (u *ServiceUpdateService[TRuntime, TCatalog, TPathPolicy]) catalogRef() TCatalog {
	if u == nil || u.catalogManager == nil {
		var zero TCatalog
		return zero
	}
	return u.catalogManager.Catalog()
}

func (u *ServiceUpdateService[TRuntime, TCatalog, TPathPolicy]) UpdateToolWithOutput(out io.Writer, name string) error {
	return u.deps.UpdateToolWithOutput(u.runtimeRef(), u.catalogRef(), out, name)
}
