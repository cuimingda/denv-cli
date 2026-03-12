package denv

import (
	"io"

	domain "github.com/cuimingda/denv-cli/internal/domain"
	infra "github.com/cuimingda/denv-cli/internal/infra"
)

// 以下断言确保 Service 仍保持各层接口兼容。
var _ RuntimeContext = (*Service)(nil)
var _ CatalogContext = (*Service)(nil)
var _ InstallContext = (*Service)(nil)
var _ UpdateContext = (*Service)(nil)
var _ ServiceContext = (*Service)(nil)
var _ ListContext = (*Service)(nil)

// Service 是 CLI 的统一服务入口，封装运行时、目录、安装和更新能力的组装与编排能力。
type Service struct {
	runtimeAdapter   *infra.RuntimeAdapter[Runtime]
	catalogManager   *infra.CatalogManager[*toolCatalog, PathPolicy]
	discoveryService *domain.ServiceDiscovery[Runtime, *toolCatalog, PathPolicy, Package]
	installPlanner   *domain.ServiceInstallPlanner[Runtime, *toolCatalog, PathPolicy, *InstallToolDriverRegistry, InstallQueue, InstallPlan, BuildInstallPlanOptions, InstallOperation]
	installExecutor  *domain.ServiceInstallExecutor[Runtime, InstallOperation, InstallQueue]
	versionResolver  *domain.ServiceVersionResolver[Runtime, *toolCatalog, PathPolicy, ListOptions, ToolListItem]
	outdatedService  *domain.ServiceOutdatedService[Runtime, *toolCatalog, PathPolicy, ToolCheckResult, OutdatedItem]
	updateService    *domain.ServiceUpdateService[Runtime, *toolCatalog, PathPolicy]
}

// NewService 创建默认 catalog 与 pathPolicy 的服务实例。
func NewService(rt Runtime) *Service {
	return NewServiceWithCatalog(rt, NewToolCatalog())
}

// NewServiceWithCatalog 允许注入 catalog，仍使用默认 pathPolicy。
func NewServiceWithCatalog(rt Runtime, catalog *toolCatalog) *Service {
	return NewServiceWithCatalogAndPolicy(rt, catalog, DefaultPathPolicy())
}

// NewServiceWithCatalogAndPolicy 允许自定义 pathPolicy，registry 使用默认实现。
func NewServiceWithCatalogAndPolicy(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) *Service {
	return NewServiceWithCatalogAndPolicyAndInstallDrivers(rt, catalog, pathPolicy, nil)
}

// NewServiceWithCatalogAndPolicyAndInstallDrivers 允许注入 registry 的完整入口。
func NewServiceWithCatalogAndPolicyAndInstallDrivers(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, registry *InstallToolDriverRegistry) *Service {
	runtimeAdapter := newServiceRuntimeAdapter(rt)
	catalogManager := newServiceCatalogManager(catalog, pathPolicy)
	registrySnapshot := normalizeInstallToolDriverRegistry(registry)
	return &Service{
		runtimeAdapter:   runtimeAdapter,
		catalogManager:   catalogManager,
		discoveryService: newServiceDiscovery(runtimeAdapter, catalogManager),
		installPlanner:   newServiceInstallPlanner(runtimeAdapter, catalogManager, registrySnapshot),
		installExecutor:  newServiceInstallExecutor(runtimeAdapter),
		versionResolver:  newServiceVersionResolver(runtimeAdapter, catalogManager),
		outdatedService:  newServiceOutdatedService(runtimeAdapter, catalogManager),
		updateService:    newServiceUpdateService(runtimeAdapter, catalogManager),
	}
}

func newServiceRuntimeAdapter(rt Runtime) *infra.RuntimeAdapter[Runtime] {
	return infra.NewRuntimeAdapter(NormalizeRuntime(rt), func() Runtime {
		return NormalizeRuntime(Runtime{})
	})
}

func newServiceCatalogManager(catalog *toolCatalog, pathPolicy PathPolicy) *infra.CatalogManager[*toolCatalog, PathPolicy] {
	return infra.NewCatalogManager(catalog, pathPolicy, NewToolCatalog, DefaultPathPolicy)
}

func newServiceDiscovery(rt *infra.RuntimeAdapter[Runtime], catalog *infra.CatalogManager[*toolCatalog, PathPolicy]) *domain.ServiceDiscovery[Runtime, *toolCatalog, PathPolicy, Package] {
	return domain.NewServiceDiscovery(rt, catalog, domain.ServiceDiscoveryDeps[Runtime, *toolCatalog, PathPolicy, Package]{
		IsCommandAvailable: func(rt Runtime, name string) bool {
			return IsCommandAvailable(rt, name)
		},
		ToolInstallState: func(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, name string) (bool, string, bool, error) {
			return ToolInstallStateWithCatalog(rt, catalog, pathPolicy, name)
		},
		ResolvedBrewBinaryPath: func(rt Runtime, name, formula string) (string, error) {
			return ResolvedBrewBinaryPath(rt, name, formula)
		},
		IsManagedByHomebrew: func(pathPolicy PathPolicy, path string) bool {
			return pathPolicy.IsManagedByHomebrew(path)
		},
		ToolDisplayName: func(catalog *toolCatalog, name string) string {
			return ToolDisplayNameWithCatalog(catalog, name)
		},
		IsInstallableTool: func(catalog *toolCatalog, name string) bool {
			return IsInstallableToolWithCatalog(catalog, name)
		},
		ResolveCommandPackages: func(catalog *toolCatalog, name string) []Package {
			return catalog.ResolveCommandPackages(name)
		},
		SupportedTools: func(catalog *toolCatalog) []string {
			return SupportedToolsWithCatalog(catalog)
		},
		InstallableTools: func(catalog *toolCatalog) []string {
			return InstallableToolsWithCatalog(catalog)
		},
		CommandPath: func(rt Runtime, name string) (string, error) {
			return CommandPath(rt, name)
		},
		IsBrewInstalled: func(rt Runtime) bool {
			return IsBrewInstalled(rt)
		},
		IsBrewFormulaInstalled: func(rt Runtime, formula string) (bool, error) {
			return IsBrewFormulaInstalled(rt, formula)
		},
	})
}

func newServiceInstallPlanner(rt *infra.RuntimeAdapter[Runtime], catalog *infra.CatalogManager[*toolCatalog, PathPolicy], registry *InstallToolDriverRegistry) *domain.ServiceInstallPlanner[Runtime, *toolCatalog, PathPolicy, *InstallToolDriverRegistry, InstallQueue, InstallPlan, BuildInstallPlanOptions, InstallOperation] {
	return domain.NewServiceInstallPlanner(
		rt,
		catalog,
		registry,
		domain.ServiceInstallPlannerDeps[Runtime, *toolCatalog, *InstallToolDriverRegistry, InstallQueue, InstallPlan, BuildInstallPlanOptions, InstallOperation]{
			BuildInstallQueue: func(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry, force bool) (InstallQueue, error) {
				return BuildInstallQueueWithCatalogAndRegistry(rt, catalog, registry, force)
			},
			BuildInstallPlan: func(rt Runtime, catalog *toolCatalog, toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
				return BuildInstallPlanModel(rt, catalog, toolName, options)
			},
			BuildInstallQueueForTool: func(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry, toolName string, force bool) (InstallQueue, error) {
				return BuildInstallQueueForToolWithCatalogAndRegistry(rt, catalog, registry, toolName, force)
			},
			BuildQueueToOperations: func(queue InstallQueue) []InstallOperation {
				return queue.ToOperations()
			},
		},
	)
}

func newServiceInstallExecutor(rt *infra.RuntimeAdapter[Runtime]) *domain.ServiceInstallExecutor[Runtime, InstallOperation, InstallQueue] {
	return domain.NewServiceInstallExecutor(rt, domain.ServiceInstallExecutorDeps[Runtime, InstallOperation, InstallQueue]{
		RunInstallOperation: func(rt Runtime, out io.Writer, op InstallOperation) error {
			return RunInstallOperation(out, rt, op)
		},
		InstallQueueToOps: func(queue InstallQueue) []InstallOperation {
			return queue.ToOperations()
		},
	})
}

func newServiceVersionResolver(rt *infra.RuntimeAdapter[Runtime], catalog *infra.CatalogManager[*toolCatalog, PathPolicy]) *domain.ServiceVersionResolver[Runtime, *toolCatalog, PathPolicy, ListOptions, ToolListItem] {
	return domain.NewServiceVersionResolver(rt, catalog, domain.ServiceVersionResolverDeps[Runtime, *toolCatalog, PathPolicy, ListOptions, ToolListItem]{
		ListToolItems: func(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, opts ListOptions) ([]ToolListItem, error) {
			return listToolItems(rt, catalog, pathPolicy, opts)
		},
		ToolVersion: func(rt Runtime, catalog *toolCatalog, name string) (string, error) {
			return ToolVersionWithCatalog(rt, catalog, name)
		},
		ToolVersionWithPath: func(rt Runtime, catalog *toolCatalog, name, commandPath string) (string, error) {
			return ToolVersionWithPathWithCatalog(rt, catalog, name, commandPath)
		},
		ToolVersionForOutdated: func(rt Runtime, catalog *toolCatalog, name string) (string, error) {
			return ToolVersionForOutdatedWithCatalog(rt, catalog, name)
		},
		ToolLatestVersion: func(rt Runtime, catalog *toolCatalog, name string) (string, error) {
			return ToolLatestVersionWithCatalog(rt, catalog, name)
		},
	})
}

func newServiceOutdatedService(rt *infra.RuntimeAdapter[Runtime], catalog *infra.CatalogManager[*toolCatalog, PathPolicy]) *domain.ServiceOutdatedService[Runtime, *toolCatalog, PathPolicy, ToolCheckResult, OutdatedItem] {
	return domain.NewServiceOutdatedService(rt, catalog, domain.ServiceOutdatedServiceDeps[Runtime, *toolCatalog, PathPolicy, ToolCheckResult, OutdatedItem]{
		OutdatedItems: func(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedItem, error) {
			return outdatedItems(rt, catalog, pathPolicy)
		},
		OutdatedChecks: func(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]ToolCheckResult, error) {
			return outdatedChecks(rt, catalog, pathPolicy)
		},
		OutdatedCheckWithOutput: func(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, out io.Writer, name string) (ToolCheckResult, error) {
			return outdatedCheckWithOutput(rt, catalog, pathPolicy, out, name)
		},
		OutdatedUpdatePlan: func(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) ([]OutdatedItem, error) {
			return outdatedUpdatePlan(rt, catalog, pathPolicy)
		},
		RunBrewUpdate: func(rt Runtime, out io.Writer) error {
			return runBrewUpdate(rt, out)
		},
	})
}

func newServiceUpdateService(rt *infra.RuntimeAdapter[Runtime], catalog *infra.CatalogManager[*toolCatalog, PathPolicy]) *domain.ServiceUpdateService[Runtime, *toolCatalog, PathPolicy] {
	return domain.NewServiceUpdateService(rt, catalog, domain.ServiceUpdateServiceDeps[Runtime, *toolCatalog]{
		UpdateToolWithOutput: func(rt Runtime, catalog *toolCatalog, out io.Writer, name string) error {
			return UpdateToolWithOutputWithCatalog(rt, catalog, out, name)
		},
	})
}

func (s *Service) runtimeAdapterRef() *infra.RuntimeAdapter[Runtime] {
	if s == nil || s.runtimeAdapter == nil {
		return newServiceRuntimeAdapter(Runtime{})
	}
	return s.runtimeAdapter
}

func (s *Service) catalogManagerRef() *infra.CatalogManager[*toolCatalog, PathPolicy] {
	if s == nil || s.catalogManager == nil {
		return newServiceCatalogManager(NewToolCatalog(), DefaultPathPolicy())
	}
	return s.catalogManager
}

func (s *Service) discoveryRef() *domain.ServiceDiscovery[Runtime, *toolCatalog, PathPolicy, Package] {
	if s == nil || s.discoveryService == nil {
		return newServiceDiscovery(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.discoveryService
}

func (s *Service) installPlannerRef() *domain.ServiceInstallPlanner[Runtime, *toolCatalog, PathPolicy, *InstallToolDriverRegistry, InstallQueue, InstallPlan, BuildInstallPlanOptions, InstallOperation] {
	if s == nil || s.installPlanner == nil {
		return newServiceInstallPlanner(s.runtimeAdapterRef(), s.catalogManagerRef(), nil)
	}
	return s.installPlanner
}

func (s *Service) installExecutorRef() *domain.ServiceInstallExecutor[Runtime, InstallOperation, InstallQueue] {
	if s == nil || s.installExecutor == nil {
		return newServiceInstallExecutor(s.runtimeAdapterRef())
	}
	return s.installExecutor
}

func (s *Service) versionResolverRef() *domain.ServiceVersionResolver[Runtime, *toolCatalog, PathPolicy, ListOptions, ToolListItem] {
	if s == nil || s.versionResolver == nil {
		return newServiceVersionResolver(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.versionResolver
}

func (s *Service) outdatedServiceRef() *domain.ServiceOutdatedService[Runtime, *toolCatalog, PathPolicy, ToolCheckResult, OutdatedItem] {
	if s == nil || s.outdatedService == nil {
		return newServiceOutdatedService(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.outdatedService
}

func (s *Service) updateManagerRef() *domain.ServiceUpdateService[Runtime, *toolCatalog, PathPolicy] {
	if s == nil || s.updateService == nil {
		return newServiceUpdateService(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.updateService
}

func (s *Service) runtime() Runtime {
	return s.runtimeAdapterRef().Runtime()
}

func (s *Service) catalogRef() *toolCatalog {
	return s.catalogManagerRef().Catalog()
}

func (s *Service) pathPolicyRef() PathPolicy {
	return s.catalogManagerRef().PathPolicy()
}

func (s *Service) IsCommandAvailable(name string) bool {
	return s.discoveryRef().IsCommandAvailable(name)
}

func (s *Service) ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return s.discoveryRef().ToolInstallState(name)
}

func (s *Service) ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return s.discoveryRef().ResolvedBrewBinaryPath(name, formula)
}

func (s *Service) IsManagedByHomebrew(path string) bool {
	return s.discoveryRef().IsManagedByHomebrew(path)
}

func (s *Service) ToolDisplayName(name string) string {
	return s.discoveryRef().ToolDisplayName(name)
}

func (s *Service) IsInstallableTool(name string) bool {
	return s.discoveryRef().IsInstallableTool(name)
}

func (s *Service) ResolveCommandPackages(name string) []Package {
	return s.discoveryRef().ResolveCommandPackages(name)
}

func (s *Service) SupportedTools() []string {
	return s.discoveryRef().SupportedTools()
}

func (s *Service) InstallableTools() []string {
	return s.discoveryRef().InstallableTools()
}

func (s *Service) CommandPath(name string) (string, error) {
	return s.discoveryRef().CommandPath(name)
}

func (s *Service) ToolVersion(name string) (string, error) {
	return s.versionResolverRef().ToolVersion(name)
}

func (s *Service) ToolVersionWithPath(name, commandPath string) (string, error) {
	return s.versionResolverRef().ToolVersionWithPath(name, commandPath)
}

func (s *Service) ToolVersionForOutdated(name string) (string, error) {
	return s.versionResolverRef().ToolVersionForOutdated(name)
}

func (s *Service) ExtractVersion(out string) (string, error) {
	return ExtractVersion(out)
}

func (s *Service) SplitVersionParts(version string) []int {
	return SplitVersionParts(version)
}

func (s *Service) IsBrewInstalled() bool {
	return s.discoveryRef().IsBrewInstalled()
}

func (s *Service) IsBrewFormulaInstalled(formula string) (bool, error) {
	return s.discoveryRef().IsBrewFormulaInstalled(formula)
}

func (s *Service) ListToolItems(opts ListOptions) ([]ToolListItem, error) {
	return s.versionResolverRef().ListToolItems(opts)
}

func (s *Service) OutdatedChecks() ([]ToolCheckResult, error) {
	return s.outdatedServiceRef().OutdatedChecks()
}

func (s *Service) OutdatedItems() ([]OutdatedItem, error) {
	return s.outdatedServiceRef().OutdatedItems()
}

func (s *Service) OutdatedCheckWithOutput(out io.Writer, name string) (ToolCheckResult, error) {
	return s.outdatedServiceRef().OutdatedCheckWithOutput(out, name)
}

func (s *Service) OutdatedUpdatePlan() ([]OutdatedItem, error) {
	return s.outdatedServiceRef().OutdatedUpdatePlan()
}

func (s *Service) RunBrewUpdate(out io.Writer) error {
	return s.outdatedServiceRef().RunBrewUpdate(out)
}

func (s *Service) BuildInstallOperations(force bool) ([]InstallOperation, error) {
	queue, err := s.BuildInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

func (s *Service) BuildInstallQueue(force bool) (InstallQueue, error) {
	return s.installPlannerRef().BuildInstallQueue(force)
}

func (s *Service) BuildInstallOperationsForTool(toolName string, force bool) ([]InstallOperation, error) {
	return s.installPlannerRef().BuildInstallOperationsForTool(toolName, force)
}

func (s *Service) BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	return s.installPlannerRef().BuildInstallPlan(toolName, options)
}

func (s *Service) BuildInstallQueueForTool(toolName string, force bool) (InstallQueue, error) {
	return s.installPlannerRef().BuildInstallQueueForTool(toolName, force)
}

func (s *Service) RunInstallOperation(out io.Writer, op InstallOperation) error {
	return s.installExecutorRef().RunInstallOperation(out, op)
}

func (s *Service) ExecuteInstallOperations(out io.Writer, operations []InstallOperation) error {
	return s.installExecutorRef().ExecuteInstallOperations(out, operations)
}

func (s *Service) ExecuteInstallQueue(out io.Writer, queue InstallQueue) error {
	return s.installExecutorRef().ExecuteInstallQueue(out, queue)
}

func (s *Service) InstallTool(name string) error {
	return InstallToolWithCatalog(s.runtime(), s.catalogRef(), name)
}

func (s *Service) InstallToolWithOptions(name string, options InstallOptions) error {
	return InstallWithCatalog(s.runtime(), s.catalogRef(), name, options)
}

func (s *Service) UpdateToolWithOutput(out io.Writer, name string) error {
	return s.updateManagerRef().UpdateToolWithOutput(out, name)
}

func (s *Service) ToolLatestVersion(name string) (string, error) {
	return s.versionResolverRef().ToolLatestVersion(name)
}

func (s *Service) CompareVersions(current string, latest string) int {
	return CompareVersions(current, latest)
}

func (s *Service) ParseBrewStableVersion(output []byte) (string, error) {
	return ParseBrewStableVersion(output)
}
