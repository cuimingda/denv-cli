package domain

import (
	"io"

	"github.com/cuimingda/denv-cli/internal/infra"
)

type ServiceInstallPlannerDeps[TRuntime any, TCatalog any, TRegistry any, TInstallQueue any, TInstallPlan any, TInstallPlanOptions any, TInstallOperation any] struct {
	BuildInstallQueue             func(TRuntime, TCatalog, TRegistry, bool) (TInstallQueue, error)
	BuildInstallPlan              func(TRuntime, TCatalog, string, TInstallPlanOptions) (TInstallPlan, error)
	BuildInstallOperationsForTool func(TRuntime, TCatalog, TRegistry, string, bool) ([]TInstallOperation, error)
	BuildInstallQueueForTool      func(TRuntime, TCatalog, TRegistry, string, bool) (TInstallQueue, error)
	BuildQueueToOperations        func(TInstallQueue) []TInstallOperation
}

type ServiceInstallPlanner[TRuntime any, TCatalog any, TPathPolicy any, TRegistry any, TInstallQueue any, TInstallPlan any, TInstallPlanOptions any, TInstallOperation any] struct {
	runtimeAdapter *infra.RuntimeAdapter[TRuntime]
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy]
	registry      TRegistry
	deps          ServiceInstallPlannerDeps[TRuntime, TCatalog, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]
}

func NewServiceInstallPlanner[TRuntime any, TCatalog any, TPathPolicy any, TRegistry any, TInstallQueue any, TInstallPlan any, TInstallPlanOptions any, TInstallOperation any](
	runtimeAdapter *infra.RuntimeAdapter[TRuntime],
	catalogManager *infra.CatalogManager[TCatalog, TPathPolicy],
	registry TRegistry,
	deps ServiceInstallPlannerDeps[TRuntime, TCatalog, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation],
) *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation] {
	return &ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]{
		runtimeAdapter: runtimeAdapter,
		catalogManager: catalogManager,
		registry:      registry,
		deps:          deps,
	}
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) runtimeRef() TRuntime {
	if p == nil || p.runtimeAdapter == nil {
		var zero TRuntime
		return zero
	}
	return p.runtimeAdapter.Runtime()
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) catalogRef() TCatalog {
	if p == nil || p.catalogManager == nil {
		var zero TCatalog
		return zero
	}
	return p.catalogManager.Catalog()
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) registryRef() TRegistry {
	if p == nil {
		var zero TRegistry
		return zero
	}
	return p.registry
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) BuildInstallQueue(force bool) (TInstallQueue, error) {
	return p.deps.BuildInstallQueue(p.runtimeRef(), p.catalogRef(), p.registryRef(), force)
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) BuildInstallPlan(toolName string, options TInstallPlanOptions) (TInstallPlan, error) {
	return p.deps.BuildInstallPlan(p.runtimeRef(), p.catalogRef(), toolName, options)
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) BuildInstallOperationsForTool(toolName string, force bool) ([]TInstallOperation, error) {
	if p.deps.BuildInstallOperationsForTool != nil {
		return p.deps.BuildInstallOperationsForTool(p.runtimeRef(), p.catalogRef(), p.registryRef(), toolName, force)
	}

	if p.deps.BuildInstallQueueForTool == nil || p.deps.BuildQueueToOperations == nil {
		var zero []TInstallOperation
		return zero, nil
	}

	queue, err := p.deps.BuildInstallQueueForTool(p.runtimeRef(), p.catalogRef(), p.registryRef(), toolName, force)
	if err != nil {
		return nil, err
	}
	return p.deps.BuildQueueToOperations(queue), nil
}

func (p *ServiceInstallPlanner[TRuntime, TCatalog, TPathPolicy, TRegistry, TInstallQueue, TInstallPlan, TInstallPlanOptions, TInstallOperation]) BuildInstallQueueForTool(toolName string, force bool) (TInstallQueue, error) {
	return p.deps.BuildInstallQueueForTool(p.runtimeRef(), p.catalogRef(), p.registryRef(), toolName, force)
}

type ServiceInstallExecutorDeps[TRuntime any, TInstallOperation any, TInstallQueue any] struct {
	RunInstallOperation  func(TRuntime, io.Writer, TInstallOperation) error
	InstallQueueToOps    func(TInstallQueue) []TInstallOperation
}

type ServiceInstallExecutor[TRuntime any, TInstallOperation any, TInstallQueue any] struct {
	runtimeAdapter *infra.RuntimeAdapter[TRuntime]
	deps           ServiceInstallExecutorDeps[TRuntime, TInstallOperation, TInstallQueue]
}

func NewServiceInstallExecutor[TRuntime any, TInstallOperation any, TInstallQueue any](
	runtimeAdapter *infra.RuntimeAdapter[TRuntime],
	deps ServiceInstallExecutorDeps[TRuntime, TInstallOperation, TInstallQueue],
) *ServiceInstallExecutor[TRuntime, TInstallOperation, TInstallQueue] {
	return &ServiceInstallExecutor[TRuntime, TInstallOperation, TInstallQueue]{
		runtimeAdapter: runtimeAdapter,
		deps:           deps,
	}
}

func (e *ServiceInstallExecutor[TRuntime, TInstallOperation, TInstallQueue]) runtimeRef() TRuntime {
	if e == nil || e.runtimeAdapter == nil {
		var zero TRuntime
		return zero
	}
	return e.runtimeAdapter.Runtime()
}

func (e *ServiceInstallExecutor[TRuntime, TInstallOperation, TInstallQueue]) RunInstallOperation(out io.Writer, op TInstallOperation) error {
	return e.deps.RunInstallOperation(e.runtimeRef(), out, op)
}

func (e *ServiceInstallExecutor[TRuntime, TInstallOperation, TInstallQueue]) ExecuteInstallOperations(out io.Writer, operations []TInstallOperation) error {
	for _, operation := range operations {
		if err := e.RunInstallOperation(out, operation); err != nil {
			return err
		}
	}
	return nil
}

func (e *ServiceInstallExecutor[TRuntime, TInstallOperation, TInstallQueue]) ExecuteInstallQueue(out io.Writer, queue TInstallQueue) error {
	if e.deps.InstallQueueToOps == nil {
		return nil
	}
	return e.ExecuteInstallOperations(out, e.deps.InstallQueueToOps(queue))
}

