// internal/service.go 提供统一的 Service 门面，组装子服务并对外暴露命令与业务能力的稳定入口。
package denv

import "io"

// 以下断言确保 Service 仍保持各层接口兼容
var _ RuntimeContext = (*Service)(nil)
var _ CatalogContext = (*Service)(nil)
var _ InstallContext = (*Service)(nil)
var _ UpdateContext = (*Service)(nil)
var _ ServiceContext = (*Service)(nil)
var _ ListContext = (*Service)(nil)

// Service 是 CLI 的统一服务入口，封装运行时、目录、安装和更新能力。
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

// newServiceRuntimeAdapter 在创建时对 Runtime 做 Normalize 处理。
func newServiceRuntimeAdapter(rt Runtime) *serviceRuntimeAdapter {
	return &serviceRuntimeAdapter{rt: NormalizeRuntime(rt)}
}

// runtime 返回标准化的运行时对象。
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

// newServiceCatalogManager 在 catalog/pathPolicy 为空时兜底默认值。
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

// newServiceDiscovery 组装发现能力子模块。
func newServiceDiscovery(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceDiscovery {
	return &serviceDiscovery{
		runtime: rt,
		catalog: catalog,
	}
}

// runtimeRef 保证总能返回可用的 Runtime。
func (d *serviceDiscovery) runtimeRef() Runtime {
	if d == nil || d.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return d.runtime.runtime()
}

// catalogRef 保证总能返回可用 catalog。
func (d *serviceDiscovery) catalogRef() *toolCatalog {
	if d == nil || d.catalog == nil {
		return NewToolCatalog()
	}
	return d.catalog.catalogRef()
}

// pathPolicyRef 安全返回用于判断工具归属的路径策略。
func (d *serviceDiscovery) pathPolicyRef() PathPolicy {
	if d == nil || d.catalog == nil {
		return DefaultPathPolicy()
	}
	return d.catalog.pathPolicyRef()
}

// IsCommandAvailable 委托到 CommandPath 探测逻辑。
// IsCommandAvailable 检查命令是否存在于 PATH，用于判断工具是否可直接调用。
func (d *serviceDiscovery) IsCommandAvailable(name string) bool {
	return IsCommandAvailable(d.runtimeRef(), name)
}

// ToolInstallState 查询工具是否已安装、命令路径以及是否由 Homebrew 管理。
func (d *serviceDiscovery) ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return ToolInstallStateWithCatalog(d.runtimeRef(), d.catalogRef(), d.pathPolicyRef(), name)
}

// ResolvedBrewBinaryPath 在 catalog 语义下计算 brew 管理命令的绝对路径。
func (d *serviceDiscovery) ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return ResolvedBrewBinaryPath(d.runtimeRef(), name, formula)
}

// IsManagedByHomebrew 判断某路径是否落在 brew 安装目录规则下。
func (d *serviceDiscovery) IsManagedByHomebrew(path string) bool {
	return d.pathPolicyRef().IsManagedByHomebrew(path)
}

// ToolDisplayName 读取工具在展示文案中的友好名称。
func (d *serviceDiscovery) ToolDisplayName(name string) string {
	return ToolDisplayNameWithCatalog(d.catalogRef(), name)
}

// IsInstallableTool 判断工具是否在可安装集合内。
func (d *serviceDiscovery) IsInstallableTool(name string) bool {
	return IsInstallableToolWithCatalog(d.catalogRef(), name)
}

// ResolveCommandPackages 获取命令对应的候选安装包列表。
func (d *serviceDiscovery) ResolveCommandPackages(name string) []Package {
	return d.catalogRef().ResolveCommandPackages(name)
}

// SupportedTools 返回支持的全部工具列表。
func (d *serviceDiscovery) SupportedTools() []string {
	return SupportedToolsWithCatalog(d.catalogRef())
}

// InstallableTools 返回当前目录中可安装的工具集合。
func (d *serviceDiscovery) InstallableTools() []string {
	return InstallableToolsWithCatalog(d.catalogRef())
}

// CommandPath 解析工具命令的二进制路径。
func (d *serviceDiscovery) CommandPath(name string) (string, error) {
	return CommandPath(d.runtimeRef(), name)
}

// IsBrewInstalled 检查当前机器是否已安装 Homebrew。
func (d *serviceDiscovery) IsBrewInstalled() bool {
	return IsBrewInstalled(d.runtimeRef())
}

// IsBrewFormulaInstalled 查询某个 brew formula 是否已安装。
func (d *serviceDiscovery) IsBrewFormulaInstalled(formula string) (bool, error) {
	return IsBrewFormulaInstalled(d.runtimeRef(), formula)
}

// catalogRef 获取 service 的 catalog 引用。
func (m *serviceCatalogManager) catalogRef() *toolCatalog {
	if m == nil || m.catalog == nil {
		return NewToolCatalog()
	}
	return m.catalog
}

// pathPolicyRef 获取路径策略引用。
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

// newServiceInstallPlanner 创建规划子服务。
func newServiceInstallPlanner(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager, registry *InstallToolDriverRegistry) *serviceInstallPlanner {
	return &serviceInstallPlanner{
		runtime: rt,
		catalog: catalog,
		registry: registry,
	}
}

// runtimeRef 安全返回执行时使用的 Runtime。
func (p *serviceInstallPlanner) runtimeRef() Runtime {
	if p == nil || p.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return p.runtime.runtime()
}

// catalogRef 安全返回 catalog 引用，服务级兜底会回退默认实例。
func (p *serviceInstallPlanner) catalogRef() *toolCatalog {
	if p == nil || p.catalog == nil {
		return NewToolCatalog()
	}
	return p.catalog.catalogRef()
}

// registryRef 允许返回 nil，交给下游恢复默认行为。
func (p *serviceInstallPlanner) registryRef() *InstallToolDriverRegistry {
	if p == nil || p.registry == nil {
		return nil
	}
	return p.registry
}

type serviceInstallExecutor struct {
	runtime *serviceRuntimeAdapter
}

// newServiceInstallExecutor 创建执行器子服务。
func newServiceInstallExecutor(rt *serviceRuntimeAdapter) *serviceInstallExecutor {
	return &serviceInstallExecutor{
		runtime: rt,
	}
}

// runtimeRef 安全返回执行器使用的 Runtime。
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

// newServiceVersionResolver 创建版本查询子服务。
func newServiceVersionResolver(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceVersionResolver {
	return &serviceVersionResolver{
		runtime: rt,
		catalog: catalog,
	}
}

// runtimeRef 安全返回运行时。
func (r *serviceVersionResolver) runtimeRef() Runtime {
	if r == nil || r.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return r.runtime.runtime()
}

// catalogRef 安全返回 catalog。
func (r *serviceVersionResolver) catalogRef() *toolCatalog {
	if r == nil || r.catalog == nil {
		return NewToolCatalog()
	}
	return r.catalog.catalogRef()
}

// pathPolicyRef 安全返回路径策略。
func (r *serviceVersionResolver) pathPolicyRef() PathPolicy {
	if r == nil || r.catalog == nil {
		return DefaultPathPolicy()
	}
	return r.catalog.pathPolicyRef()
}

// ListToolItems 按版本/列表参数查询工具信息。
func (r *serviceVersionResolver) ListToolItems(opts ListOptions) ([]ToolListItem, error) {
	return listToolItems(r.runtimeRef(), r.catalogRef(), r.pathPolicyRef(), opts)
}

// ToolVersion 获取工具当前版本。
func (r *serviceVersionResolver) ToolVersion(name string) (string, error) {
	return ToolVersionWithCatalog(r.runtimeRef(), r.catalogRef(), name)
}

// ToolVersionWithPath 按指定命令路径解析版本。
func (r *serviceVersionResolver) ToolVersionWithPath(name, commandPath string) (string, error) {
	return ToolVersionWithPathWithCatalog(r.runtimeRef(), r.catalogRef(), name, commandPath)
}

// ToolVersionForOutdated 返回用于过期判断的版本。
func (r *serviceVersionResolver) ToolVersionForOutdated(name string) (string, error) {
	return ToolVersionForOutdatedWithCatalog(r.runtimeRef(), r.catalogRef(), name)
}

// ToolLatestVersion 查询工具可安装的最新版本。
func (r *serviceVersionResolver) ToolLatestVersion(name string) (string, error) {
	return ToolLatestVersionWithCatalog(r.runtimeRef(), r.catalogRef(), name)
}

type serviceOutdatedService struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
}

// newServiceOutdatedService 创建过期检测子服务。
func newServiceOutdatedService(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceOutdatedService {
	return &serviceOutdatedService{
		runtime: rt,
		catalog: catalog,
	}
}

// runtimeRef 安全返回 runtime。
func (o *serviceOutdatedService) runtimeRef() Runtime {
	if o == nil || o.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return o.runtime.runtime()
}

// catalogRef 安全返回 catalog。
func (o *serviceOutdatedService) catalogRef() *toolCatalog {
	if o == nil || o.catalog == nil {
		return NewToolCatalog()
	}
	return o.catalog.catalogRef()
}

// pathPolicyRef 安全返回 pathPolicy。
func (o *serviceOutdatedService) pathPolicyRef() PathPolicy {
	if o == nil || o.catalog == nil {
		return DefaultPathPolicy()
	}
	return o.catalog.pathPolicyRef()
}

// OutdatedItems 返回兼容列表格式。
func (o *serviceOutdatedService) OutdatedItems() ([]OutdatedItem, error) {
	return outdatedItems(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

// OutdatedChecks 返回详细检查结构体。
func (o *serviceOutdatedService) OutdatedChecks() ([]ToolCheckResult, error) {
	return outdatedChecks(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

// OutdatedUpdatePlan 返回可更新工具清单。
func (o *serviceOutdatedService) OutdatedUpdatePlan() ([]OutdatedItem, error) {
	return outdatedUpdatePlan(o.runtimeRef(), o.catalogRef(), o.pathPolicyRef())
}

type serviceUpdateService struct {
	runtime *serviceRuntimeAdapter
	catalog *serviceCatalogManager
}

// newServiceUpdateService 创建更新子服务。
func newServiceUpdateService(rt *serviceRuntimeAdapter, catalog *serviceCatalogManager) *serviceUpdateService {
	return &serviceUpdateService{
		runtime: rt,
		catalog: catalog,
	}
}

// runtimeRef 安全返回 runtime。
func (o *serviceUpdateService) runtimeRef() Runtime {
	if o == nil || o.runtime == nil {
		return newServiceRuntimeAdapter(Runtime{}).runtime()
	}
	return o.runtime.runtime()
}

// catalogRef 安全返回 catalog。
func (o *serviceUpdateService) catalogRef() *toolCatalog {
	if o == nil || o.catalog == nil {
		return NewToolCatalog()
	}
	return o.catalog.catalogRef()
}

// UpdateToolWithOutput 用 catalog 执行更新逻辑。
func (o *serviceUpdateService) UpdateToolWithOutput(out io.Writer, name string) error {
	return UpdateToolWithOutputWithCatalog(o.runtimeRef(), o.catalogRef(), out, name)
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

// runtimeAdapterRef 返回运行时适配器，缺失时返回默认适配器。
func (s *Service) runtimeAdapterRef() *serviceRuntimeAdapter {
	if s == nil || s.runtimeAdapter == nil {
		return newServiceRuntimeAdapter(Runtime{})
	}
	return s.runtimeAdapter
}

// catalogManagerRef 返回目录管理器，缺失时返回默认实例。
func (s *Service) catalogManagerRef() *serviceCatalogManager {
	if s == nil || s.catalogManager == nil {
		return newServiceCatalogManager(NewToolCatalog(), DefaultPathPolicy())
	}
	return s.catalogManager
}

// discoveryRef 返回发现子服务，缺失时重建。
func (s *Service) discoveryRef() *serviceDiscovery {
	if s == nil || s.discoveryService == nil {
		return newServiceDiscovery(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.discoveryService
}

// installPlannerRef 返回安装规划子服务，缺失时重建。
func (s *Service) installPlannerRef() *serviceInstallPlanner {
	if s == nil || s.installPlanner == nil {
		return newServiceInstallPlanner(s.runtimeAdapterRef(), s.catalogManagerRef(), nil)
	}
	return s.installPlanner
}

// installExecutorRef 返回安装执行子服务，缺失时重建。
func (s *Service) installExecutorRef() *serviceInstallExecutor {
	if s == nil || s.installExecutor == nil {
		return newServiceInstallExecutor(s.runtimeAdapterRef())
	}
	return s.installExecutor
}

// versionResolverRef 返回版本查询子服务，缺失时重建。
func (s *Service) versionResolverRef() *serviceVersionResolver {
	if s == nil || s.versionResolver == nil {
		return newServiceVersionResolver(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.versionResolver
}

// outdatedServiceRef 返回过期检查子服务，缺失时重建。
func (s *Service) outdatedServiceRef() *serviceOutdatedService {
	if s == nil || s.outdatedService == nil {
		return newServiceOutdatedService(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.outdatedService
}

// updateManagerRef 返回更新子服务，缺失时重建。
func (s *Service) updateManagerRef() *serviceUpdateService {
	if s == nil || s.updateService == nil {
		return newServiceUpdateService(s.runtimeAdapterRef(), s.catalogManagerRef())
	}
	return s.updateService
}

// runtime 返回服务内部统一使用的运行时实例。
func (s *Service) runtime() Runtime {
	return s.runtimeAdapterRef().runtime()
}

// catalogRef 返回当前服务使用的 catalog。
func (s *Service) catalogRef() *toolCatalog {
	return s.catalogManagerRef().catalogRef()
}

// pathPolicyRef 返回当前服务使用的路径策略。
func (s *Service) pathPolicyRef() PathPolicy {
	return s.catalogManagerRef().pathPolicyRef()
}

// IsCommandAvailable 查询命令是否存在。
func (s *Service) IsCommandAvailable(name string) bool {
	return s.discoveryRef().IsCommandAvailable(name)
}

// ToolInstallState 获取工具安装状态、命令路径和 Homebrew 管理标记。
func (s *Service) ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return s.discoveryRef().ToolInstallState(name)
}

// ResolvedBrewBinaryPath 解析 brew 命令的二进制绝对路径。
func (s *Service) ResolvedBrewBinaryPath(name, formula string) (string, error) {
	return s.discoveryRef().ResolvedBrewBinaryPath(name, formula)
}

// IsManagedByHomebrew 判断路径是否为 Homebrew 管理。
func (s *Service) IsManagedByHomebrew(path string) bool {
	return s.discoveryRef().IsManagedByHomebrew(path)
}

// ToolDisplayName 获取工具展示名。
func (s *Service) ToolDisplayName(name string) string {
	return s.discoveryRef().ToolDisplayName(name)
}

// IsInstallableTool 判断是否可安装。
func (s *Service) IsInstallableTool(name string) bool {
	return s.discoveryRef().IsInstallableTool(name)
}

// ResolveCommandPackages 按命令名返回关联包。
func (s *Service) ResolveCommandPackages(name string) []Package {
	return s.discoveryRef().ResolveCommandPackages(name)
}

// SupportedTools 返回 catalog 支持的所有工具名。
func (s *Service) SupportedTools() []string {
	return s.discoveryRef().SupportedTools()
}

// InstallableTools 返回可安装工具列表。
func (s *Service) InstallableTools() []string {
	return s.discoveryRef().InstallableTools()
}

// CommandPath 查询命令完整路径。
func (s *Service) CommandPath(name string) (string, error) {
	return s.discoveryRef().CommandPath(name)
}

// ToolVersion 获取当前工具版本。
func (s *Service) ToolVersion(name string) (string, error) {
	return s.versionResolverRef().ToolVersion(name)
}

// ToolVersionWithPath 按命令路径获取版本。
func (s *Service) ToolVersionWithPath(name, commandPath string) (string, error) {
	return s.versionResolverRef().ToolVersionWithPath(name, commandPath)
}

// ToolVersionForOutdated 获取用于过期判定的版本。
func (s *Service) ToolVersionForOutdated(name string) (string, error) {
	return s.versionResolverRef().ToolVersionForOutdated(name)
}

// ExtractVersion 抽取版本字符串。
func (s *Service) ExtractVersion(out string) (string, error) {
	return ExtractVersion(out)
}

// SplitVersionParts 按规则拆分版本片段。
func (s *Service) SplitVersionParts(version string) []int {
	return SplitVersionParts(version)
}

// IsBrewInstalled 返回当前环境是否存在 brew 命令。
func (s *Service) IsBrewInstalled() bool {
	return s.discoveryRef().IsBrewInstalled()
}

// IsBrewFormulaInstalled 查询 brew formula 是否已安装。
func (s *Service) IsBrewFormulaInstalled(formula string) (bool, error) {
	return s.discoveryRef().IsBrewFormulaInstalled(formula)
}

// ListToolItems 返回可展示列表内容。
func (s *Service) ListToolItems(opts ListOptions) ([]ToolListItem, error) {
	return s.versionResolverRef().ListToolItems(opts)
}

// OutdatedChecks 获取完整过期检查结果。
func (s *Service) OutdatedChecks() ([]ToolCheckResult, error) {
	return s.outdatedServiceRef().OutdatedChecks()
}

// OutdatedItems 返回用户友好结构的过期结果。
func (s *Service) OutdatedItems() ([]OutdatedItem, error) {
	return s.outdatedServiceRef().OutdatedItems()
}

// OutdatedUpdatePlan 返回待更新的工具清单。
func (s *Service) OutdatedUpdatePlan() ([]OutdatedItem, error) {
	return s.outdatedServiceRef().OutdatedUpdatePlan()
}

// BuildInstallOperations 先构建队列再转成动作列表。
func (s *Service) BuildInstallOperations(force bool) ([]InstallOperation, error) {
	queue, err := s.BuildInstallQueue(force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

// BuildInstallQueue 构建安装队列。
func (s *Service) BuildInstallQueue(force bool) (InstallQueue, error) {
	return s.installPlannerRef().BuildInstallQueue(force)
}

// BuildInstallOperationsForTool 先构建指定工具队列再转为执行操作列表。
func (s *Service) BuildInstallOperationsForTool(toolName string, force bool) ([]InstallOperation, error) {
	queue, err := s.BuildInstallQueueForTool(toolName, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

// BuildInstallPlan 构建单工具安装计划。
func (s *Service) BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	return s.installPlannerRef().BuildInstallPlan(toolName, options)
}

// BuildInstallQueueForTool 构建指定工具的安装队列。
func (s *Service) BuildInstallQueueForTool(toolName string, force bool) (InstallQueue, error) {
	return s.installPlannerRef().BuildInstallQueueForTool(toolName, force)
}

// RunInstallOperation 执行单条安装动作。
func (s *Service) RunInstallOperation(out io.Writer, op InstallOperation) error {
	return s.installExecutorRef().RunInstallOperation(out, op)
}

// ExecuteInstallOperations 执行动作列表。
func (s *Service) ExecuteInstallOperations(out io.Writer, operations []InstallOperation) error {
	return s.installExecutorRef().ExecuteInstallOperations(out, operations)
}

// ExecuteInstallQueue 执行完整队列。
func (s *Service) ExecuteInstallQueue(out io.Writer, queue InstallQueue) error {
	return s.installExecutorRef().ExecuteInstallQueue(out, queue)
}

// InstallTool 使用默认安装配置执行安装。
func (s *Service) InstallTool(name string) error {
	return InstallToolWithCatalog(s.runtime(), s.catalogRef(), name)
}

// InstallToolWithOptions 按 options 执行安装流程。
func (s *Service) InstallToolWithOptions(name string, options InstallOptions) error {
	return InstallWithCatalog(s.runtime(), s.catalogRef(), name, options)
}

// UpdateToolWithOutput 执行工具更新并输出日志到指定 writer。
func (s *Service) UpdateToolWithOutput(out io.Writer, name string) error {
	return s.updateManagerRef().UpdateToolWithOutput(out, name)
}

// ToolLatestVersion 获取最新版本号。
func (s *Service) ToolLatestVersion(name string) (string, error) {
	return s.versionResolverRef().ToolLatestVersion(name)
}

// CompareVersions 在服务层透传版本比较能力。
func (s *Service) CompareVersions(current string, latest string) int {
	return CompareVersions(current, latest)
}

// ParseBrewStableVersion 复用 brew JSON 解析逻辑。
func (s *Service) ParseBrewStableVersion(output []byte) (string, error) {
	return ParseBrewStableVersion(output)
}

// BuildInstallQueue 委托带 registry 的底层函数构建全量安装队列。
func (p *serviceInstallPlanner) BuildInstallQueue(force bool) (InstallQueue, error) {
	return BuildInstallQueueWithCatalogAndRegistry(p.runtimeRef(), p.catalogRef(), p.registryRef(), force)
}

// BuildInstallPlan 调用内部 planner 构建单工具计划。
func (p *serviceInstallPlanner) BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	planner := newInstallPlanServiceWithDefaults(p.runtimeRef(), p.catalogRef(), p.registryRef())
	return planner.BuildInstallPlan(toolName, options)
}

// BuildInstallOperationsForTool 将工具队列转为动作序列。
func (p *serviceInstallPlanner) BuildInstallOperationsForTool(toolName string, force bool) ([]InstallOperation, error) {
	queue, err := p.BuildInstallQueueForTool(toolName, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

// BuildInstallQueueForTool 构建指定工具队列，支持 policy registry。
func (p *serviceInstallPlanner) BuildInstallQueueForTool(toolName string, force bool) (InstallQueue, error) {
	return BuildInstallQueueForToolWithCatalogAndRegistry(p.runtimeRef(), p.catalogRef(), p.registryRef(), toolName, force)
}

// RunInstallOperation 委托底层执行。
func (e *serviceInstallExecutor) RunInstallOperation(out io.Writer, op InstallOperation) error {
	return RunInstallOperation(out, e.runtimeRef(), op)
}

// ExecuteInstallOperations 顺序执行操作列表，任意一步失败则立即返回错误。
func (e *serviceInstallExecutor) ExecuteInstallOperations(out io.Writer, operations []InstallOperation) error {
	for _, operation := range operations {
		if err := e.RunInstallOperation(out, operation); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteInstallQueue 将 InstallQueue 转为动作列表后逐条执行。
func (e *serviceInstallExecutor) ExecuteInstallQueue(out io.Writer, queue InstallQueue) error {
	return e.ExecuteInstallOperations(out, queue.ToOperations())
}
