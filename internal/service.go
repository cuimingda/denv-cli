package denv

import "io"

var _ RuntimeContext = (*Service)(nil)
var _ CatalogContext = (*Service)(nil)
var _ InstallContext = (*Service)(nil)
var _ UpdateContext = (*Service)(nil)
var _ ServiceContext = (*Service)(nil)
var _ ListContext = (*Service)(nil)

type Service struct {
	runtimeAdapter   *serviceRuntimeAdapter
	catalogManager   *serviceCatalogManager
	discoveryService *serviceDiscovery
	installPlanner   *serviceInstallPlanner
	installExecutor  *serviceInstallExecutor
	versionResolver  *serviceVersionResolver
	outdatedService  *serviceOutdatedService
	updateService    *serviceUpdateService
}

type serviceRuntimeAdapter struct {
	rt Runtime
}

func newServiceRuntimeAdapter(rt Runtime) *serviceRuntimeAdapter {
	return &serviceRuntimeAdapter{rt: NormalizeRuntime(rt)}
}

func (a *serviceRuntimeAdapter) runtime() Runtime {
	if a == nil {
		return NormalizeRuntime(Runtime{})
	}
	return a.rt
}

type serviceCatalogManager struct {
	catalog    *toolCatalog
	pathPolicy PathPolicy
}

func newServiceCatalogManager(catalog *toolCatalog, pathPolicy PathPolicy) *serviceCatalogManager {
	if catalog == nil {
		catalog = NewToolCatalog()
	}
	if pathPolicy == nil {
		pathPolicy = DefaultPathPolicy()
	}
	return &serviceCatalogManager{catalog: catalog, pathPolicy: pathPolicy}
}

type serviceDiscovery struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
}

func newServiceDiscovery(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceDiscovery {
	return &serviceDiscovery{
		runtime: rt,
		catalog: catalog,
	}
}

func (d *serviceDiscovery) runtimeRef() Runtime {
	if d == nil || d.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return d.runtime.runtime()
}

func (d *serviceDiscovery) catalogRef() *toolCatalog {
	if d == nil || d.catalog == nil {
		return NewToolCatalog()
	}
	return d.catalog.catalogRef()
}

func (d *serviceDiscovery) pathPolicyRef() PathPolicy {
	if d == nil || d.catalog == nil {
		return DefaultPathPolicy()
	}
	return d.catalog.pathPolicyRef()
}

func (d *serviceDiscovery) IsCommandAvailable(name string) bool {
	return IsCommandAvailable(d.runtimeRef(), name)
}

func (d *serviceDiscovery) ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return ToolInstallStateWithCatalog(d.runtimeRef(), d.catalogRef(), d.pathPolicyRef(), name)
}

func (d *serviceDiscovery) ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return ResolvedBrewBinaryPath(d.runtimeRef(), name, formula)
}

func (d *serviceDiscovery) IsManagedByHomebrew(path string) bool {
	return d.pathPolicyRef().IsManagedByHomebrew(path)
}

func (d *serviceDiscovery) ToolDisplayName(name string) string {
	return ToolDisplayNameWithCatalog(d.catalogRef(), name)
}

func (d *serviceDiscovery) IsInstallableTool(name string) bool {
	return IsInstallableToolWithCatalog(d.catalogRef(), name)
}

func (d *serviceDiscovery) ResolveCommandPackages(name string) []Package {
	return d.catalogRef().ResolveCommandPackages(name)
}

func (d *serviceDiscovery) SupportedTools() []string {
	return SupportedToolsWithCatalog(d.catalogRef())
}

func (d *serviceDiscovery) InstallableTools() []string {
	return InstallableToolsWithCatalog(d.catalogRef())
}

func (d *serviceDiscovery) CommandPath(name string) (string, error) {
	return CommandPath(d.runtimeRef(), name)
}

func (d *serviceDiscovery) IsBrewInstalled() bool {
	return IsBrewInstalled(d.runtimeRef())
}

func (d *serviceDiscovery) IsBrewFormulaInstalled(formula string) (bool, error) {
	return IsBrewFormulaInstalled(d.runtimeRef(), formula)
}

func (m *serviceCatalogManager) catalogRef() *toolCatalog {
	if m == nil || m.catalog == nil {
		return NewToolCatalog()
	}
	return m.catalog
}

func (m *serviceCatalogManager) pathPolicyRef() PathPolicy {
	if m == nil || m.pathPolicy == nil {
		return DefaultPathPolicy()
	}
	return m.pathPolicy
}

type serviceInstallPlanner struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
	registry *InstallToolDriverRegistry
}

func newServiceInstallPlanner(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager, registry *InstallToolDriverRegistry) *serviceInstallPlanner {
	return &serviceInstallPlanner{
		runtime: rt,
		catalog: catalog,
		registry: registry,
	}
}

func (p *serviceInstallPlanner) runtimeRef() Runtime {
	if p == nil || p.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return p.runtime.runtime()
}

func (p *serviceInstallPlanner) catalogRef() *toolCatalog {
	if p == nil || p.catalog == nil {
		return NewToolCatalog()
	}
	return p.catalog.catalogRef()
}

func (p *serviceInstallPlanner) registryRef() *InstallToolDriverRegistry {
	if p == nil || p.registry == nil {
		return nil
	}
	return p.registry
}

type serviceInstallExecutor struct {
	runtime *serviceRuntimeAdapter
}

func newServiceInstallExecutor(rt *serviceRuntimeAdapter) *serviceInstallExecutor {
	return &serviceInstallExecutor{
		runtime: rt,
	}
}

func (e *serviceInstallExecutor) runtimeRef() Runtime {
	if e == nil || e.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return e.runtime.runtime()
}

type serviceVersionResolver struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
}

func newServiceVersionResolver(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceVersionResolver {
	return &serviceVersionResolver{
		runtime: rt,
		catalog: catalog,
	}
}

func (r *serviceVersionResolver) runtimeRef() Runtime {
	if r == nil || r.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return r.runtime.runtime()
}

func (r *serviceVersionResolver) catalogRef() *toolCatalog {
	if r == nil || r.catalog == nil {
		return NewToolCatalog()
	}
	return r.catalog.catalogRef()
}

func (r *serviceVersionResolver) pathPolicyRef() PathPolicy {
	if r == nil || r.catalog == nil {
		return DefaultPathPolicy()
	}
	return r.catalog.pathPolicyRef()
}

func (r *serviceVersionResolver) ListToolItems(opts ListOptions) ([]ToolListItem, error) {
	return listToolItems(r.runtimeRef(), r.catalogRef(), r.pathPolicyRef(), opts)
}

func (r *serviceVersionResolver) ToolVersion(name string) (string, error) {
	return ToolVersionWithCatalog(r.runtimeRef(), r.catalogRef(), name)
}

func (r *serviceVersionResolver) ToolVersionWithPath(name, commandPath string) (string, error) {
	return ToolVersionWithPathWithCatalog(r.runtimeRef(), r.catalogRef(), name, commandPath)
}

func (r *serviceVersionResolver) ToolVersionForOutdated(name string) (string, error) {
	return ToolVersionForOutdatedWithCatalog(r.runtimeRef(), r.catalogRef(), name)
}

func (r *serviceVersionResolver) ToolLatestVersion(name string) (string, error) {
	return ToolLatestVersionWithCatalog(r.runtimeRef(), r.catalogRef(), name)
}

type serviceOutdatedService struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
}

func newServiceOutdatedService(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceOutdatedService {
	return &serviceOutdatedService{
		runtime: rt,
		catalog: catalog,
	}
}

func (o *serviceOutdatedService) runtimeRef() Runtime {
	if o == nil || o.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return o.runtime.runtime()
}

func (o *serviceOutdatedService) catalogRef() *toolCatalog {
	if o == nil || o.catalog == nil {
		return NewToolCatalog()
	}
	return o.catalog.catalogRef()
}

func (o *serviceOutdatedService) pathPolicyRef() PathPolicy {
	if o == nil || o.catalog == nil {
		return DefaultPathPolicy()
	}
	return o.catalog.pathPolicyRef()
}

func (o *serviceOutdatedService) OutdatedItems() ([]OutdatedItem, error) {
	return outdatedItems(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

func (o *serviceOutdatedService) OutdatedChecks() ([]ToolCheckResult, error) {
	return outdatedChecks(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

func (o *serviceOutdatedService) OutdatedUpdatePlan() ([]OutdatedItem, error) {
	return outdatedUpdatePlan(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

type serviceUpdateService struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
}

func newServiceUpdateService(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceUpdateService {
	return &serviceUpdateService{
		runtime: rt,
		catalog: catalog,
	}
}

func (o *serviceUpdateService) runtimeRef() Runtime {
	if o == nil || o.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return o.runtime.runtime()
}

func (o *serviceUpdateService) catalogRef() *toolCatalog {
	if o == nil || o.catalog == nil {
		return NewToolCatalog()
	}
	return o.catalog.catalogRef()
}

func (o *serviceUpdateService) UpdateToolWithOutput(out io.Writer, name string) error {
	return UpdateToolWithOutputWithCatalog(o.runtimeRef(), o.catalogRef(), out, name)
}

func NewService(rt Runtime) *Service {
	return NewServiceWithCatalog(rt, NewToolCatalog())
}

func NewServiceWithCatalog(rt Runtime, catalog *toolCatalog) *Service {
	return NewServiceWithCatalogAndPolicy(rt, catalog, DefaultPathPolicy())
}

func NewServiceWithCatalogAndPolicy(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) *Service {
	return NewServiceWithCatalogAndPolicyAndInstallDrivers(rt, catalog, pathPolicy, nil)
}

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

func (s *Service) runtimeAdapterRef() *serviceRuntimeAdapter {
	if s == nil || s.runtimeAdapter == nil {
		return newServiceRuntimeAdapter(Runtime{})
	}
	return s.runtimeAdapter
}

func (s *Service) catalogManagerRef() *serviceCatalogManager {
	if s == nil || s.catalogManager == nil {
		return newServiceCatalogManager(NewToolCatalog(), DefaultPathPolicy())
	}
	return s.catalogManager
}

func (s *Service) discoveryRef() *serviceDiscovery {
	if s == nil || s.discoveryService == nil {
		return newServiceDiscovery(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.discoveryService
}

func (s *Service) installPlannerRef() *serviceInstallPlanner {
	if s == nil || s.installPlanner == nil {
		return newServiceInstallPlanner(s.runtimeAdapterRef(), s.catalogManagerRef(), nil)
	}
	return s.installPlanner
}

func (s *Service) installExecutorRef() *serviceInstallExecutor {
	if s == nil || s.installExecutor == nil {
		return newServiceInstallExecutor(s.runtimeAdapterRef())
	}
	return s.installExecutor
}

func (s *Service) versionResolverRef() *serviceVersionResolver {
	if s == nil || s.versionResolver == nil {
		return newServiceVersionResolver(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.versionResolver
}

func (s *Service) outdatedServiceRef() *serviceOutdatedService {
	if s == nil || s.outdatedService == nil {
		return newServiceOutdatedService(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.outdatedService
}

func (s *Service) updateManagerRef() *serviceUpdateService {
	if s == nil || s.updateService == nil {
		return newServiceUpdateService(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.updateService
}

func (s *Service) runtime() Runtime {
	return s.runtimeAdapterRef().runtime()
}

func (s *Service) catalogRef() *toolCatalog {
	return s.catalogManagerRef().catalogRef()
}

func (s *Service) pathPolicyRef() PathPolicy {
	return s.catalogManagerRef().pathPolicyRef()
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

func (s *Service) OutdatedUpdatePlan() ([]OutdatedItem, error) {
	return s.outdatedServiceRef().OutdatedUpdatePlan()
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
	queue, err := s.BuildInstallQueueForTool(toolName, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
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

func (p *serviceInstallPlanner) BuildInstallQueue(force bool) (InstallQueue, error) {
	return BuildInstallQueueWithCatalogAndRegistry(p.runtimeRef(), p.catalogRef(), p.registryRef(), force)
}

func (p *serviceInstallPlanner) BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	planner := newInstallPlanServiceWithDefaults(p.runtimeRef(), p.catalogRef(), p.registryRef())
	return planner.BuildInstallPlan(toolName, options)
}

func (p *serviceInstallPlanner) BuildInstallOperationsForTool(toolName string, force bool) ([]InstallOperation, error) {
	queue, err := p.BuildInstallQueueForTool(toolName, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

func (p *serviceInstallPlanner) BuildInstallQueueForTool(toolName string, force bool) (InstallQueue, error) {
	return BuildInstallQueueForToolWithCatalogAndRegistry(p.runtimeRef(), p.catalogRef(), p.registryRef(), toolName, force)
}

func (e *serviceInstallExecutor) RunInstallOperation(out io.Writer, op InstallOperation) error {
	return RunInstallOperation(out, e.runtimeRef(), op)
}

func (e *serviceInstallExecutor) ExecuteInstallOperations(out io.Writer, operations []InstallOperation) error {
	for _, operation := range operations {
		if err := e.RunInstallOperation(out, operation); err != nil {
			return err
		}
	}
	return nil
}

func (e *serviceInstallExecutor) ExecuteInstallQueue(out io.Writer, queue InstallQueue) error {
	return e.ExecuteInstallOperations(out, queue.ToOperations())
}
