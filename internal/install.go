// internal/install.go 定义工具安装领域的核心流程：安装计划构建、安装策略、执行入口与命令驱动注册。
package denv

import (
	"fmt"
	"io"
	"strings"
)

type BuildInstallPlanOptions struct {
	// Force 是否强制重新生成安装计划
	Force bool
}

type InstallOptions struct {
	// Force 是否跳过存在性判断并强制执行
	Force bool
	// Output 安装输出写入目标
	Output io.Writer
}

// installQueueBuilder 根据工具名和强制选项构造安装队列。
type installQueueBuilder func(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error)

// InstallToolDriverRegistry 维护工具级安装构建器注册表。
type InstallToolDriverRegistry struct {
	// drivers 记录 toolID 到构建器的映射
	drivers map[string]installQueueBuilder
}

// newInstallToolDriverRegistry 初始化空注册表。
func newInstallToolDriverRegistry() *InstallToolDriverRegistry {
	return &InstallToolDriverRegistry{
		drivers: map[string]installQueueBuilder{},
	}
}

// NewInstallToolDriverRegistry 返回空注册表实例。
func NewInstallToolDriverRegistry() *InstallToolDriverRegistry {
	return newInstallToolDriverRegistry()
}

// NewInstallToolDriverRegistryWithDefaults 读取内置规则并注册默认 driver。
func NewInstallToolDriverRegistryWithDefaults() *InstallToolDriverRegistry {
	return NewInstallToolDriverRegistryBuilder().BuildWithDefaults()
}

// InstallToolDriverRegistryBuilder 用于构建 driver 注册表。
type InstallToolDriverRegistryBuilder struct {
	// drivers 缓存待注册的 toolID->builder
	drivers map[string]installQueueBuilder
}

func NewInstallToolDriverRegistryBuilder() *InstallToolDriverRegistryBuilder {
	return &InstallToolDriverRegistryBuilder{
		drivers: map[string]installQueueBuilder{},
	}
}

// Register 记录单个工具的安装队列构建器。
func (r *InstallToolDriverRegistryBuilder) Register(toolID string, buildQueue installQueueBuilder) {
	if r == nil || toolID == "" || buildQueue == nil {
		return
	}
	if r.drivers == nil {
		r.drivers = map[string]installQueueBuilder{}
	}
	r.drivers[toolID] = buildQueue
}

// Build 返回当前 builder 的只读 registry。
func (r *InstallToolDriverRegistryBuilder) Build() *InstallToolDriverRegistry {
	return newInstallToolDriverRegistryWithDrivers(r.drivers)
}

// BuildWithDefaults 注册所有可安装工具的默认构建逻辑。
func (r *InstallToolDriverRegistryBuilder) BuildWithDefaults() *InstallToolDriverRegistry {
	// 将默认管理工具定义映射为统一的 builder：基于工具名构建安装队列
	for _, definition := range defaultManagedToolDefinitions() {
		if !definition.Installable {
			continue
		}
		toolID := definition.ToolID
		localID := toolID
		r.Register(localID, func(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
			return buildInstallQueueByTool(rt, catalog, localID, force)
		})
	}
	return r.Build()
}

// resolve 查找指定工具的构建器。
func (r *InstallToolDriverRegistry) resolve(toolID string) (installQueueBuilder, bool) {
	if r == nil || toolID == "" {
		return nil, false
	}
	buildQueue, ok := r.drivers[toolID]
	return buildQueue, ok
}

// newInstallToolDriverRegistryWithDrivers 基于 map 创建 registry，自动过滤非法项。
func newInstallToolDriverRegistryWithDrivers(drivers map[string]installQueueBuilder) *InstallToolDriverRegistry {
	registry := newInstallToolDriverRegistry()
	for toolID, buildQueue := range drivers {
		if toolID == "" || buildQueue == nil {
			continue
		}
		registry.drivers[toolID] = buildQueue
	}
	return registry
}

// normalizeInstallToolDriverRegistry 允许传入 nil，自动回退到默认注册表。
func normalizeInstallToolDriverRegistry(registry *InstallToolDriverRegistry) *InstallToolDriverRegistry {
	if registry == nil {
		return NewInstallToolDriverRegistryWithDefaults()
	}
	return newInstallToolDriverRegistryWithDrivers(registry.drivers)
}

type installPlanService struct {
	// runtime 运行时执行环境（已规范化）
	runtime Runtime
	// catalog 工具目录（已规范化）
	catalog *toolCatalog
	// registry 安装驱动注册表
	registry *InstallToolDriverRegistry
}

// newInstallPlanService 依赖注入并初始化核心安装服务，缺省 registry 直接 panic 防止空对象。
func newInstallPlanService(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry) *installPlanService {
	if registry == nil {
		panic("install plan service requires a registry")
	}
	return &installPlanService{
		runtime:  NormalizeRuntime(rt),
		catalog:  ensureToolCatalog(catalog),
		registry: registry,
	}
}

// newInstallPlanServiceWithDefaults 补齐空 registry 后创建服务。
func newInstallPlanServiceWithDefaults(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry) *installPlanService {
	return newInstallPlanService(rt, catalog, normalizeInstallToolDriverRegistry(registry))
}

// newDefaultInstallPlanService 使用默认 catalog 与默认 registry 创建服务。
func newDefaultInstallPlanService(rt Runtime, catalog *toolCatalog) *installPlanService {
	return newInstallPlanServiceWithDefaults(rt, catalog, nil)
}

// resolveInstallToolDriver 解析工具专用计划构建器。
func (s *installPlanService) resolveInstallToolDriver(toolID string) (installQueueBuilder, bool) {
	if s == nil {
		return nil, false
	}
	if s.registry == nil {
		return nil, false
	}
	return s.registry.resolve(toolID)
}

// buildQueueForTool 先查 driver registry，未命中则按默认逻辑构建。
func (s *installPlanService) buildQueueForTool(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallQueue, error) {
	if builder, ok := s.resolveInstallToolDriver(toolName); ok {
		return builder(rt, catalog, force)
	}
	return buildInstallQueueByTool(rt, catalog, toolName, force)
}

// BuildInstallQueue 为所有可安装工具构建合并后的队列。
func (s *installPlanService) BuildInstallQueue(force bool) (InstallQueue, error) {
	queue := make(InstallQueue, 0)
	catalog := s.catalog
	for _, toolName := range catalog.installableToolsCatalog() {
		toolQueue, err := s.BuildInstallQueueForTool(toolName, force)
		if err != nil {
			return nil, err
		}
		queue = append(queue, toolQueue...)
	}
	return queue, nil
}

// BuildInstallQueueForTool 按单工具构建安装队列。
func (s *installPlanService) BuildInstallQueueForTool(toolName string, force bool) (InstallQueue, error) {
	return s.buildQueueForTool(s.runtime, s.catalog, toolName, force)
}

// BuildInstallPlan 构建单工具安装计划。
func (s *installPlanService) BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	queue, err := s.BuildInstallQueueForTool(toolName, options.Force)
	if err != nil {
		return InstallPlan{}, err
	}
	return NewInstallPlan(toolName, queue.ToOperations(), options.Force), nil
}

// BuildInstallPlanWithPolicy 透传 run-time 策略参数的安装计划构建入口。
func (s *installPlanService) BuildInstallPlanWithPolicy(toolName string, force bool, checkCommand bool, checkFormula bool) (InstallPlan, error) {
	return BuildInstallPlanWithCatalog(s.runtime, s.catalog, toolName, force, checkCommand, checkFormula)
}

// IsCommandAvailable 判断命令是否在 PATH 中可用。
func IsCommandAvailable(rt Runtime, name string) bool {
	_, err := CommandPath(rt, name)
	return err == nil
}

// CommandPath 查找可执行文件全路径。
func CommandPath(rt Runtime, name string) (string, error) {
	rt = NormalizeRuntime(rt)
	return rt.ExecutableLookup(name)
}

// ToolInstallState 查询工具是否安装、可执行路径及是否由 brew 管理。
func ToolInstallState(rt Runtime, name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return ToolInstallStateWithCatalog(rt, NewToolCatalog(), DefaultPathPolicy(), name)
}

// ToolInstallStateWithCatalog 使用注入 catalog/pathPolicy 做更准确的安装状态判断。
func ToolInstallStateWithCatalog(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy, name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	rt = NormalizeRuntime(rt)
	if catalog == nil {
		catalog = NewToolCatalog()
	}
	if pathPolicy == nil {
		pathPolicy = DefaultPathPolicy()
	}
	commandPath, err = CommandPath(rt, name)
	if err == nil {
		return true, commandPath, pathPolicy.IsManagedByHomebrew(commandPath), nil
	}

	// PATH 未发现时，尝试依据 catalog 中可管理包的 formula 反查 brew 安装痕迹
	candidates := catalog.ResolveCommandPackages(name)
	if len(candidates) == 0 {
		return false, "", false, nil
	}

	for _, candidate := range candidates {
		formula := candidate.Formula()
		installedByBrew, lookupErr := IsBrewFormulaInstalled(rt, formula)
		if lookupErr != nil {
			return false, "", false, lookupErr
		}
		if !installedByBrew {
			continue
		}

		brewPath, pathErr := ResolvedBrewBinaryPath(rt, name, formula)
		if pathErr != nil {
			return true, pathPolicy.HomebrewDefaultToolPath(name), true, nil
		}

		if brewPath != "" {
			return true, brewPath, true, nil
		}
		return true, pathPolicy.HomebrewDefaultToolPath(name), true, nil
	}

	return false, "", false, nil
}

// ResolvedBrewBinaryPath 通过 brew --prefix 推导命令真实安装路径。
func ResolvedBrewBinaryPath(rt Runtime, name, formula string) (string, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("brew", "--prefix", formula)
	if err != nil {
		return "", err
	}

	prefix := strings.TrimSpace(string(output))
	if prefix == "" {
		return "", nil
	}

	return prefix + "/bin/" + name, nil
}

// ToolDisplayName 走默认 catalog 的展示名解析。
func ToolDisplayName(name string) string {
	return ToolDisplayNameWithCatalog(NewToolCatalog(), name)
}

// ToolDisplayNameWithCatalog 依赖 catalog 提供显示名映射。
func ToolDisplayNameWithCatalog(catalog *toolCatalog, name string) string {
	catalog = ensureToolCatalog(catalog)
	if display, ok := catalog.toolDisplayName(name); ok {
		return display
	}
	return name
}

// IsInstallableTool 使用默认 catalog 判断是否可安装。
func IsInstallableTool(name string) bool {
	return IsInstallableToolWithCatalog(NewToolCatalog(), name)
}

// IsInstallableToolWithCatalog 使用指定目录判断工具是否可安装。
func IsInstallableToolWithCatalog(catalog *toolCatalog, name string) bool {
	catalog = ensureToolCatalog(catalog)
	return catalog.managedToolIsInstallable(name)
}

// SupportedTools 列出可展示工具名集合。
func SupportedTools() []string {
	return SupportedToolsWithCatalog(NewToolCatalog())
}

// SupportedToolsWithCatalog 返回 catalog 中声明的可展示命令集合。
func SupportedToolsWithCatalog(catalog *toolCatalog) []string {
	catalog = ensureToolCatalog(catalog)
	return catalog.listedToolsCatalog()
}

// InstallableTools 列出可安装工具。
func InstallableTools() []string {
	return InstallableToolsWithCatalog(NewToolCatalog())
}

// InstallableToolsWithCatalog 返回 catalog 中安装清单。
func InstallableToolsWithCatalog(catalog *toolCatalog) []string {
	catalog = ensureToolCatalog(catalog)
	return catalog.installableToolsCatalog()
}

// IsBrewInstalled 仅判断 brew 命令是否存在。
func IsBrewInstalled(rt Runtime) bool {
	return IsCommandAvailable(rt, "brew")
}

// IsBrewFormulaInstalled 判断某个 formula 是否已由 brew 安装。
func IsBrewFormulaInstalled(rt Runtime, formula string) (bool, error) {
	rt = NormalizeRuntime(rt)
	output, err := rt.CommandRunner("brew", "list", "--formula", formula)
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			return false, nil
		}
		if strings.Contains(text, "No such keg") {
			return false, nil
		}
		if strings.Contains(text, "No formula") {
			return false, nil
		}
	}

	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == formula {
			return true, nil
		}
	}
	return false, nil
}

// BuildInstallOperationsForTool 按工具构建可执行动作序列（默认 catalog）。
func BuildInstallOperationsForTool(rt Runtime, toolName string, force bool) ([]InstallOperation, error) {
	return BuildInstallOperationsForToolWithCatalog(rt, NewToolCatalog(), toolName, force)
}

// BuildInstallOperationsForToolWithCatalog 使用显式 catalog 生成动作序列。
func BuildInstallOperationsForToolWithCatalog(rt Runtime, catalog *toolCatalog, toolName string, force bool) ([]InstallOperation, error) {
	// 复用队列构建，再统一转为可执行动作列表
	queue, err := BuildInstallQueueForToolWithCatalog(rt, catalog, toolName, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

// BuildInstallPlan 使用默认 catalog 构建安装计划。
func BuildInstallPlan(rt Runtime, toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	return BuildInstallPlanModel(rt, NewToolCatalog(), toolName, options)
}

// BuildInstallPlanModel 使用显式 catalog 构建安装计划。
func BuildInstallPlanModel(rt Runtime, catalog *toolCatalog, toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	planner := newDefaultInstallPlanService(rt, catalog)
	queue, err := planner.BuildInstallQueueForTool(toolName, options.Force)
	if err != nil {
		return InstallPlan{}, err
	}
	return NewInstallPlan(toolName, queue.ToOperations(), options.Force), nil
}

// BuildInstallQueue 使用默认 catalog + 默认 registry。
func BuildInstallQueue(rt Runtime, force bool) (InstallQueue, error) {
	return BuildInstallQueueWithCatalogAndRegistry(rt, NewToolCatalog(), nil, force)
}

// BuildInstallQueueWithCatalogAndRegistry 使用自定义 registry 构建全量队列。
func BuildInstallQueueWithCatalogAndRegistry(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, catalog, registry)
	return planner.BuildInstallQueue(force)
}

// BuildInstallQueueWithCatalog 使用 catalog 构建全量队列。
func BuildInstallQueueWithCatalog(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
	return BuildInstallQueueWithCatalogAndRegistry(rt, catalog, nil, force)
}

// BuildInstallQueueForToolWithRegistry 使用 registry 对单工具做构建。
func BuildInstallQueueForToolWithRegistry(rt Runtime, toolName string, registry *InstallToolDriverRegistry, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, NewToolCatalog(), registry)
	return planner.BuildInstallQueueForTool(toolName, force)
}

// BuildInstallQueueForTool 使用默认 registry 为单工具构建队列。
func BuildInstallQueueForTool(rt Runtime, toolName string, force bool) (InstallQueue, error) {
	return BuildInstallQueueForToolWithRegistry(rt, toolName, nil, force)
}

// BuildInstallQueueForToolWithCatalogAndRegistry 使用 catalog + registry 为单工具构建队列。
func BuildInstallQueueForToolWithCatalogAndRegistry(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry, toolName string, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, catalog, registry)
	return planner.BuildInstallQueueForTool(toolName, force)
}

// BuildInstallQueueForToolWithCatalog 使用默认 registry 构建单工具队列。
func BuildInstallQueueForToolWithCatalog(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, catalog, nil)
	return planner.buildQueueForTool(rt, ensureToolCatalog(catalog), toolName, force)
}

// buildInstallQueueByTool 工具级默认构建入口，兼容不存在/不可安装保护。
func buildInstallQueueByTool(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallQueue, error) {
	catalog = ensureToolCatalog(catalog)
	tool, ok := catalog.toolLifecycle(toolName)
	if !ok || !tool.IsInstallable() {
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
	return tool.PlanInstallByPolicy(rt, catalog, force)
}

// BuildInstallPlanWithCatalog 按传入策略构建单工具计划。
func BuildInstallPlanWithCatalog(rt Runtime, catalog *toolCatalog, toolName string, force bool, checkCommand bool, checkFormula bool) (InstallPlan, error) {
	catalog = ensureToolCatalog(catalog)
	tool, ok := catalog.toolLifecycle(toolName)
	if !ok || !tool.IsInstallable() {
		return InstallPlan{}, fmt.Errorf("unsupported tool: %s", toolName)
	}

	queue, err := tool.PlanInstall(rt, catalog, force, checkCommand, checkFormula)
	if err != nil {
		return InstallPlan{}, err
	}
	return NewInstallPlan(toolName, queue.ToOperations(), force), nil
}

// BuildInstallQueueForPackageWithCatalog 将 pkg 作为安装单元构建队列。
func BuildInstallQueueForPackageWithCatalog(rt Runtime, catalog *toolCatalog, pkg Package, force bool, checkCommand bool, checkFormula bool) (InstallQueue, error) {
	if pkg == nil || !pkg.IsInstallable() {
		name := ""
		if pkg != nil {
			name = pkg.ID()
		}
		return nil, fmt.Errorf("unsupported package: %s", name)
	}

	if pkg.ID() == "" {
		return nil, fmt.Errorf("unsupported package")
	}

	return Tool{spec: pkg, requestedName: pkg.ID()}.PlanInstall(rt, ensureToolCatalog(catalog), force, checkCommand, checkFormula)
}

// BuildInstallOperations 使用默认 catalog 生成安装动作列表。
func BuildInstallOperations(rt Runtime, force bool) ([]InstallOperation, error) {
	return BuildInstallOperationsWithCatalog(rt, NewToolCatalog(), force)
}

// BuildInstallOperationsWithCatalog 将队列转换为动作列表（便于逐条执行）。
func BuildInstallOperationsWithCatalog(rt Runtime, catalog *toolCatalog, force bool) ([]InstallOperation, error) {
	queue, err := BuildInstallQueueWithCatalog(rt, catalog, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

// RunInstallOperation 用带输出流的方式执行单个安装动作。
func RunInstallOperation(out io.Writer, rt Runtime, op InstallOperation) error {
	rt = NormalizeRuntime(rt)
	if op.Spec.Name == "" {
		return nil
	}
	return rt.CommandRunnerWithOutput(out, op.Spec.Name, op.Spec.Args...)
}

// runInstallOperationCommand 仅执行命令，不收集输出，用于无日志场景。
func runInstallOperationCommand(rt Runtime, op InstallOperation) error {
	rt = NormalizeRuntime(rt)
	if op.Spec.Name == "" {
		return nil
	}
	_, err := rt.CommandRunner(op.Spec.Name, op.Spec.Args...)
	return err
}

// executeInstallOperations 按序执行动作列表，失败即中断并返回上下文错误。
func executeInstallOperations(rt Runtime, out io.Writer, operations []InstallOperation) error {
	for _, operation := range operations {
		var err error
		if out != nil {
			err = RunInstallOperation(out, rt, operation)
		} else {
			err = runInstallOperationCommand(rt, operation)
		}
		if err != nil {
			return fmt.Errorf("%s failed: %w", operation, err)
		}
	}
	return nil
}

// Install 安装工具入口（默认 catalog），对外保持兼容 API。
func Install(rt Runtime, toolName string, options InstallOptions) error {
	return InstallWithCatalog(rt, NewToolCatalog(), toolName, options)
}

// InstallWithCatalog 使用显式 catalog 的安装入口。
func InstallWithCatalog(rt Runtime, catalog *toolCatalog, toolName string, options InstallOptions) error {
	if !IsInstallableToolWithCatalog(catalog, toolName) {
		return fmt.Errorf("unsupported tool: %s", toolName)
	}
	plan, err := buildInstallPlanForToolWithRunPolicy(rt, catalog, toolName, options.Force)
	if err != nil {
		return err
	}
	return ExecuteInstallPlanWithCatalog(rt, catalog, plan, options)
}

// InstallTool 以默认选项安装单工具。
func InstallTool(rt Runtime, name string) error {
	return InstallWithCatalog(rt, NewToolCatalog(), name, InstallOptions{})
}

// InstallToolWithCatalog 使用显式 catalog 的默认选项安装。
func InstallToolWithCatalog(rt Runtime, catalog *toolCatalog, name string) error {
	return InstallWithCatalog(rt, catalog, name, InstallOptions{})
}

// installTool 内部入口，允许注入输出行为。
func installTool(rt Runtime, name string, force bool, out io.Writer, withOutput bool) error {
	return installToolWithCatalog(rt, NewToolCatalog(), name, force, out, withOutput)
}

// installToolWithCatalog 将旧参数签名映射到统一 options。
func installToolWithCatalog(rt Runtime, catalog *toolCatalog, name string, force bool, out io.Writer, withOutput bool) error {
	options := InstallOptions{
		Force: force,
	}
	if withOutput {
		if out == nil {
			out = io.Discard
		}
		options.Output = out
	}
	return installToolWithCatalogWithOptions(rt, catalog, name, options)
}

// installToolWithCatalogWithOptions 先生成计划，再执行安装流程。
func installToolWithCatalogWithOptions(rt Runtime, catalog *toolCatalog, name string, options InstallOptions) error {
	catalog = ensureToolCatalog(catalog)
	plan, err := buildInstallPlanForToolWithRunPolicy(rt, catalog, name, options.Force)
	if err != nil {
		return err
	}
	return ExecuteInstallPlanWithCatalog(rt, catalog, plan, options)
}

// buildInstallPlanForToolWithRunPolicy 按运行时策略生成安装计划（用于真正执行）。
func buildInstallPlanForToolWithRunPolicy(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallPlan, error) {
	catalog = ensureToolCatalog(catalog)
	tool, ok := catalog.toolLifecycle(toolName)
	if !ok || !tool.IsInstallable() {
		return InstallPlan{}, fmt.Errorf("unsupported tool: %s", toolName)
	}

	queue, err := tool.PlanInstallByRunPolicy(rt, catalog, force)
	if err != nil {
		return InstallPlan{}, err
	}
	return NewInstallPlan(toolName, queue.ToOperations(), force), nil
}

// ExecuteInstallPlanWithCatalog 执行计划并在空动作时给出明确错误。
func ExecuteInstallPlanWithCatalog(rt Runtime, catalog *toolCatalog, plan InstallPlan, options InstallOptions) error {
	rt = NormalizeRuntime(rt)
	catalog = ensureToolCatalog(catalog)
	operations := plan.ToOperations()
	if len(operations) == 0 {
		tool, ok := catalog.toolLifecycle(plan.ToolID)
		if ok {
			return fmt.Errorf(tool.AlreadyInstalledLabel())
		}
		return fmt.Errorf("unsupported tool: %s", plan.ToolID)
	}

	return executeInstallOperations(rt, options.Output, operations)
}

// ensureToolCatalog 在 catalog 为空时创建默认 catalog。
func ensureToolCatalog(catalog *toolCatalog) *toolCatalog {
	if catalog == nil {
		return NewToolCatalog()
	}
	return catalog
}

// packageHasCommand 快速判断包是否已具备某个 command 可执行文件。
func packageHasCommand(rt Runtime, pkg Package) bool {
	if pkg == nil {
		return false
	}
	for _, command := range pkg.Commands() {
		if IsCommandAvailable(rt, command.Name) {
			return true
		}
	}
	return false
}

// resolveCheckPolicy 为空策略时返回默认 false 策略。
func resolveCheckPolicy(policy CheckPolicy) CheckPolicy {
	if policy == nil {
		return NewPackageCheckPolicy(false, false)
	}
	return policy
}

// resolveInstallPolicy 为空策略时返回默认三段组合策略。
func resolveInstallPolicy(policy InstallPolicy) InstallPolicy {
	if policy == nil {
		return NewPackageInstallPolicy(
			NewPackageCheckPolicy(false, false),
			NewPackageCheckPolicy(false, false),
			NewPackageVersionPolicy(nil),
		)
	}
	return policy
}

// UpdateToolWithOutput 对外入口，默认使用内置 catalog。
func UpdateToolWithOutput(rt Runtime, out io.Writer, name string) error {
	return UpdateToolWithOutputWithCatalog(rt, NewToolCatalog(), out, name)
}

// UpdateToolWithOutputWithCatalog 兼容 npm 与非 npm 的升级路径。
func UpdateToolWithOutputWithCatalog(rt Runtime, catalog *toolCatalog, out io.Writer, name string) error {
	rt = NormalizeRuntime(rt)
	installed, _, _, err := ToolInstallStateWithCatalog(rt, catalog, DefaultPathPolicy(), name)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("tool %s is not installed", name)
	}

	if name == "npm" {
		if err := rt.CommandRunnerWithOutput(out, "npm", "install", "-g", "npm@latest"); err != nil {
			return fmt.Errorf("npm update failed: %w", err)
		}
		return nil
	}

	if !IsBrewInstalled(rt) {
		return fmt.Errorf("homebrew is not installed")
	}

	formula, ok := catalog.brewFormulaForTool(name)
	if !ok {
		return fmt.Errorf("unsupported tool: %s", name)
	}

	if err := rt.CommandRunnerWithOutput(out, "brew", "upgrade", formula); err != nil {
		return fmt.Errorf("brew upgrade %s failed: %w", formula, err)
	}
	return nil
}

// IsHomebrewPath 兼容旧接口：判断路径是否为 brew 管理路径。
func IsHomebrewPath(path string) bool {
	return DefaultPathPolicy().IsManagedByHomebrew(path)
}
