package denv

import (
	"fmt"
	"io"
	"strings"
)

type BuildInstallPlanOptions struct {
	Force bool
}

type InstallOptions struct {
	Force  bool
	Output io.Writer
}

type installQueueBuilder func(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error)

type InstallToolDriverRegistry struct {
	drivers map[string]installQueueBuilder
}

func newInstallToolDriverRegistry() *InstallToolDriverRegistry {
	return &InstallToolDriverRegistry{
		drivers: map[string]installQueueBuilder{},
	}
}

func NewInstallToolDriverRegistry() *InstallToolDriverRegistry {
	return newInstallToolDriverRegistry()
}

func NewInstallToolDriverRegistryWithDefaults() *InstallToolDriverRegistry {
	return NewInstallToolDriverRegistryBuilder().BuildWithDefaults()
}

type InstallToolDriverRegistryBuilder struct {
	drivers map[string]installQueueBuilder
}

func NewInstallToolDriverRegistryBuilder() *InstallToolDriverRegistryBuilder {
	return &InstallToolDriverRegistryBuilder{
		drivers: map[string]installQueueBuilder{},
	}
}

func (r *InstallToolDriverRegistryBuilder) Register(toolID string, buildQueue installQueueBuilder) {
	if r == nil || toolID == "" || buildQueue == nil {
		return
	}
	if r.drivers == nil {
		r.drivers = map[string]installQueueBuilder{}
	}
	r.drivers[toolID] = buildQueue
}

func (r *InstallToolDriverRegistryBuilder) Build() *InstallToolDriverRegistry {
	return newInstallToolDriverRegistryWithDrivers(r.drivers)
}

func (r *InstallToolDriverRegistryBuilder) BuildWithDefaults() *InstallToolDriverRegistry {
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

func (r *InstallToolDriverRegistry) resolve(toolID string) (installQueueBuilder, bool) {
	if r == nil || toolID == "" {
		return nil, false
	}
	buildQueue, ok := r.drivers[toolID]
	return buildQueue, ok
}

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

func normalizeInstallToolDriverRegistry(registry *InstallToolDriverRegistry) *InstallToolDriverRegistry {
	if registry == nil {
		return NewInstallToolDriverRegistryWithDefaults()
	}
	return newInstallToolDriverRegistryWithDrivers(registry.drivers)
}

type installPlanService struct {
	runtime  Runtime
	catalog  *toolCatalog
	registry *InstallToolDriverRegistry
}

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

func newInstallPlanServiceWithDefaults(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry) *installPlanService {
	return newInstallPlanService(rt, catalog, normalizeInstallToolDriverRegistry(registry))
}

func newDefaultInstallPlanService(rt Runtime, catalog *toolCatalog) *installPlanService {
	return newInstallPlanServiceWithDefaults(rt, catalog, nil)
}

func (s *installPlanService) resolveInstallToolDriver(toolID string) (installQueueBuilder, bool) {
	if s == nil {
		return nil, false
	}
	if s.registry == nil {
		return nil, false
	}
	return s.registry.resolve(toolID)
}

func (s *installPlanService) buildQueueForTool(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallQueue, error) {
	if builder, ok := s.resolveInstallToolDriver(toolName); ok {
		return builder(rt, catalog, force)
	}
	return buildInstallQueueByTool(rt, catalog, toolName, force)
}

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

func (s *installPlanService) BuildInstallQueueForTool(toolName string, force bool) (InstallQueue, error) {
	return s.buildQueueForTool(s.runtime, s.catalog, toolName, force)
}

func (s *installPlanService) BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	queue, err := s.BuildInstallQueueForTool(toolName, options.Force)
	if err != nil {
		return InstallPlan{}, err
	}
	return NewInstallPlan(toolName, queue.ToOperations(), options.Force), nil
}

func (s *installPlanService) BuildInstallPlanWithPolicy(toolName string, force bool, checkCommand bool, checkFormula bool) (InstallPlan, error) {
	return BuildInstallPlanWithCatalog(s.runtime, s.catalog, toolName, force, checkCommand, checkFormula)
}

func IsCommandAvailable(rt Runtime, name string) bool {
	_, err := CommandPath(rt, name)
	return err == nil
}

func CommandPath(rt Runtime, name string) (string, error) {
	rt = NormalizeRuntime(rt)
	return rt.ExecutableLookup(name)
}

func ToolInstallState(rt Runtime, name string) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return ToolInstallStateWithCatalog(rt, NewToolCatalog(), DefaultPathPolicy(), name)
}

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

func ToolDisplayName(name string) string {
	return ToolDisplayNameWithCatalog(NewToolCatalog(), name)
}

func ToolDisplayNameWithCatalog(catalog *toolCatalog, name string) string {
	catalog = ensureToolCatalog(catalog)
	if display, ok := catalog.toolDisplayName(name); ok {
		return display
	}
	return name
}

func IsInstallableTool(name string) bool {
	return IsInstallableToolWithCatalog(NewToolCatalog(), name)
}

func IsInstallableToolWithCatalog(catalog *toolCatalog, name string) bool {
	catalog = ensureToolCatalog(catalog)
	return catalog.managedToolIsInstallable(name)
}

func SupportedTools() []string {
	return SupportedToolsWithCatalog(NewToolCatalog())
}

func SupportedToolsWithCatalog(catalog *toolCatalog) []string {
	catalog = ensureToolCatalog(catalog)
	return catalog.listedToolsCatalog()
}

func InstallableTools() []string {
	return InstallableToolsWithCatalog(NewToolCatalog())
}

func InstallableToolsWithCatalog(catalog *toolCatalog) []string {
	catalog = ensureToolCatalog(catalog)
	return catalog.installableToolsCatalog()
}

func IsBrewInstalled(rt Runtime) bool {
	return IsCommandAvailable(rt, "brew")
}

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

func BuildInstallOperationsForTool(rt Runtime, toolName string, force bool) ([]InstallOperation, error) {
	return BuildInstallOperationsForToolWithCatalog(rt, NewToolCatalog(), toolName, force)
}

func BuildInstallOperationsForToolWithCatalog(rt Runtime, catalog *toolCatalog, toolName string, force bool) ([]InstallOperation, error) {
	queue, err := BuildInstallQueueForToolWithCatalog(rt, catalog, toolName, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

func BuildInstallPlan(rt Runtime, toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	return BuildInstallPlanModel(rt, NewToolCatalog(), toolName, options)
}

func BuildInstallPlanModel(rt Runtime, catalog *toolCatalog, toolName string, options BuildInstallPlanOptions) (InstallPlan, error) {
	planner := newDefaultInstallPlanService(rt, catalog)
	queue, err := planner.BuildInstallQueueForTool(toolName, options.Force)
	if err != nil {
		return InstallPlan{}, err
	}
	return NewInstallPlan(toolName, queue.ToOperations(), options.Force), nil
}

func BuildInstallQueue(rt Runtime, force bool) (InstallQueue, error) {
	return BuildInstallQueueWithCatalogAndRegistry(rt, NewToolCatalog(), nil, force)
}

func BuildInstallQueueWithCatalogAndRegistry(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, catalog, registry)
	return planner.BuildInstallQueue(force)
}

func BuildInstallQueueWithCatalog(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
	return BuildInstallQueueWithCatalogAndRegistry(rt, catalog, nil, force)
}

func BuildInstallQueueForToolWithRegistry(rt Runtime, toolName string, registry *InstallToolDriverRegistry, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, NewToolCatalog(), registry)
	return planner.BuildInstallQueueForTool(toolName, force)
}

func BuildInstallQueueForTool(rt Runtime, toolName string, force bool) (InstallQueue, error) {
	return BuildInstallQueueForToolWithRegistry(rt, toolName, nil, force)
}

func BuildInstallQueueForToolWithCatalogAndRegistry(rt Runtime, catalog *toolCatalog, registry *InstallToolDriverRegistry, toolName string, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, catalog, registry)
	return planner.BuildInstallQueueForTool(toolName, force)
}

func BuildInstallQueueForToolWithCatalog(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallQueue, error) {
	planner := newInstallPlanServiceWithDefaults(rt, catalog, nil)
	return planner.buildQueueForTool(rt, ensureToolCatalog(catalog), toolName, force)
}

func buildInstallQueueByTool(rt Runtime, catalog *toolCatalog, toolName string, force bool) (InstallQueue, error) {
	catalog = ensureToolCatalog(catalog)
	tool, ok := catalog.toolLifecycle(toolName)
	if !ok || !tool.IsInstallable() {
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
	return tool.PlanInstallByPolicy(rt, catalog, force)
}

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

func BuildInstallOperations(rt Runtime, force bool) ([]InstallOperation, error) {
	return BuildInstallOperationsWithCatalog(rt, NewToolCatalog(), force)
}

func BuildInstallOperationsWithCatalog(rt Runtime, catalog *toolCatalog, force bool) ([]InstallOperation, error) {
	queue, err := BuildInstallQueueWithCatalog(rt, catalog, force)
	if err != nil {
		return nil, err
	}
	return queue.ToOperations(), nil
}

func RunInstallOperation(out io.Writer, rt Runtime, op InstallOperation) error {
	rt = NormalizeRuntime(rt)
	if op.Spec.Name == "" {
		return nil
	}
	return rt.CommandRunnerWithOutput(out, op.Spec.Name, op.Spec.Args...)
}

func runInstallOperationCommand(rt Runtime, op InstallOperation) error {
	rt = NormalizeRuntime(rt)
	if op.Spec.Name == "" {
		return nil
	}
	_, err := rt.CommandRunner(op.Spec.Name, op.Spec.Args...)
	return err
}

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

func Install(rt Runtime, toolName string, options InstallOptions) error {
	return InstallWithCatalog(rt, NewToolCatalog(), toolName, options)
}

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

func InstallTool(rt Runtime, name string) error {
	return InstallWithCatalog(rt, NewToolCatalog(), name, InstallOptions{})
}

func InstallToolWithCatalog(rt Runtime, catalog *toolCatalog, name string) error {
	return InstallWithCatalog(rt, catalog, name, InstallOptions{})
}

func installTool(rt Runtime, name string, force bool, out io.Writer, withOutput bool) error {
	return installToolWithCatalog(rt, NewToolCatalog(), name, force, out, withOutput)
}

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

func installToolWithCatalogWithOptions(rt Runtime, catalog *toolCatalog, name string, options InstallOptions) error {
	catalog = ensureToolCatalog(catalog)
	plan, err := buildInstallPlanForToolWithRunPolicy(rt, catalog, name, options.Force)
	if err != nil {
		return err
	}
	return ExecuteInstallPlanWithCatalog(rt, catalog, plan, options)
}

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

func ensureToolCatalog(catalog *toolCatalog) *toolCatalog {
	if catalog == nil {
		return NewToolCatalog()
	}
	return catalog
}

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

func resolveCheckPolicy(policy CheckPolicy) CheckPolicy {
	if policy == nil {
		return NewPackageCheckPolicy(false, false)
	}
	return policy
}

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

func UpdateToolWithOutput(rt Runtime, out io.Writer, name string) error {
	return UpdateToolWithOutputWithCatalog(rt, NewToolCatalog(), out, name)
}

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

func IsHomebrewPath(path string) bool {
	return DefaultPathPolicy().IsManagedByHomebrew(path)
}
