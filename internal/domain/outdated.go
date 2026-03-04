package domain

import (
	"github.com/cuimingda/denv-cli/internal/infra"
)

type ServiceOutdatedServiceDeps[TRuntime any, TCatalog any, TPathPolicy any, TToolCheckResult any, TOutdatedItem any] struct {
	OutdatedItems      func(TRuntime, TCatalog, TPathPolicy) ([]TOutdatedItem, error)
	OutdatedChecks     func(TRuntime, TCatalog, TPathPolicy) ([]TToolCheckResult, error)
	OutdatedUpdatePlan func(TRuntime, TCatalog, TPathPolicy) ([]TOutdatedItem, error)
}

type ServiceOutdatedService[TRuntime any, TCatalog any, TPathPolicy any, TToolCheckResult any, TOutdatedItem any] struct {
	runtimeAdapter *infra.RuntimeAdapter[TRuntime]
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy]
	deps           ServiceOutdatedServiceDeps[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]
}

func NewServiceOutdatedService[TRuntime any, TCatalog any, TPathPolicy any, TToolCheckResult any, TOutdatedItem any](
	runtimeAdapter *infra.RuntimeAdapter[TRuntime],
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy],
	deps ServiceOutdatedServiceDeps[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem],
) *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem] {
	return &ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]{
		runtimeAdapter: runtimeAdapter,
		catalogManager: catalogManager,
		deps:           deps,
	}
}

func (o *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]) runtimeRef() TRuntime {
	if o == nil || o.runtimeAdapter == nil {
		var zero TRuntime
		return zero
	}
	return o.runtimeAdapter.Runtime()
}

func (o *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]) catalogRef() TCatalog {
	if o == nil || o.catalogManager == nil {
		var zero TCatalog
		return zero
	}
	return o.catalogManager.Catalog()
}

func (o *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]) pathPolicyRef() TPathPolicy {
	if o == nil || o.catalogManager == nil {
		var zero TPathPolicy
		return zero
	}
	return o.catalogManager.PathPolicy()
}

func (o *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]) OutdatedItems() ([]TOutdatedItem, error) {
	return o.deps.OutdatedItems(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

func (o *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]) OutdatedChecks() ([]TToolCheckResult, error) {
	return o.deps.OutdatedChecks(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

func (o *ServiceOutdatedService[TRuntime, TCatalog, TPathPolicy, TToolCheckResult, TOutdatedItem]) OutdatedUpdatePlan() ([]TOutdatedItem, error) {
	return o.deps.OutdatedUpdatePlan(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}
