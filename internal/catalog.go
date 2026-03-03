package denv

import (
	"fmt"
	"sort"
	"strings"
)

type Command struct {
	Name        string
	DisplayName string
}

// ToolDefinition 描述一个可管理工具的元数据定义。
type ToolDefinition struct {
	ToolID              string
	Formula             string
	Commands            []Command
	ListCommands        []string
	Installable         bool
	InstallPolicy       InstallPolicy
	AlreadyInstalled    string
	ExtraInstallOps     []CommandSpec
	DisplayNameOverride string
	ListOrder           int
	InstallOrder        int
}

// listedCommands 取 catalog 展示层要显示的命令名列表。
func (d ToolDefinition) listedCommands() []string {
	if len(d.ListCommands) > 0 {
		// 优先使用显式配置的 list 列表，支持同一 tool 显示多个命令
		commandNames := make([]string, 0, len(d.ListCommands))
		for _, command := range d.ListCommands {
			commandNames = append(commandNames, command)
		}
		return commandNames
	}

	commandNames := make([]string, 0, len(d.Commands))
	for _, command := range d.Commands {
		commandNames = append(commandNames, command.Name)
	}
	return commandNames
}

// ToolDefinitions 返回内置默认定义的副本。
func ToolDefinitions() []ToolDefinition {
	definitions := defaultManagedToolDefinitions()
	out := make([]ToolDefinition, len(definitions))
	copy(out, definitions)
	return out
}

// DefaultInstallableToolIDs 返回默认配置中可安装工具的 ID 列表。
func DefaultInstallableToolIDs() []string {
	definitions := ToolDefinitions()
	ids := make([]string, 0, len(definitions))
	for _, def := range definitions {
		if def.Installable {
			ids = append(ids, def.ToolID)
		}
	}
	return ids
}

// CheckPolicy 声明安装时是否检查命令或 formula。
type CheckPolicy interface {
	CheckCommand() bool
	CheckFormula() bool
}

// VersionPolicy 声明查询版本时使用的参数。
type VersionPolicy interface {
	VersionArgs() []string
}

// InstallPolicy 汇总计划时检查、执行时检查与版本参数策略。
type InstallPolicy interface {
	PlanCheckPolicy() CheckPolicy
	RunCheckPolicy() CheckPolicy
	VersionPolicy() VersionPolicy
}

type packageCheckPolicy struct {
	checkCommand bool
	checkFormula bool
}

// CheckCommand 返回是否通过命令存在性判断是否已安装。
func (p packageCheckPolicy) CheckCommand() bool {
	return p.checkCommand
}

// CheckFormula 返回是否通过 brew formula 判断是否已安装。
func (p packageCheckPolicy) CheckFormula() bool {
	return p.checkFormula
}

// packageVersionPolicy 定义版本参数策略。
type packageVersionPolicy struct {
	versionArgs []string
}

// VersionArgs 返回一份副本，防止外部修改策略数组。
func (p packageVersionPolicy) VersionArgs() []string {
	return append([]string{}, p.versionArgs...)
}

// packageInstallPolicy 是工具默认策略组合实现。
type packageInstallPolicy struct {
	planCheck CheckPolicy
	runCheck  CheckPolicy
	version   VersionPolicy
}

func (p packageInstallPolicy) PlanCheckPolicy() CheckPolicy {
	return p.planCheck
}

func (p packageInstallPolicy) RunCheckPolicy() CheckPolicy {
	return p.runCheck
}

func (p packageInstallPolicy) VersionPolicy() VersionPolicy {
	return p.version
}

// NewPackageCheckPolicy 构造检查策略。
func NewPackageCheckPolicy(checkCommand bool, checkFormula bool) CheckPolicy {
	return packageCheckPolicy{checkCommand: checkCommand, checkFormula: checkFormula}
}

// NewPackageVersionPolicy 构造版本参数策略。
func NewPackageVersionPolicy(versionArgs []string) VersionPolicy {
	return packageVersionPolicy{versionArgs: versionArgs}
}

// NewPackageInstallPolicy 构造完整安装策略。
func NewPackageInstallPolicy(planCheck CheckPolicy, runCheck CheckPolicy, versionPolicy VersionPolicy) InstallPolicy {
	return packageInstallPolicy{
		planCheck: planCheck,
		runCheck:  runCheck,
		version:   versionPolicy,
	}
}

type Package interface {
	ID() string
	Formula() string
	InstallPolicy() InstallPolicy
	DisplayName(command string) string
	Commands() []Command
	IsInstallable() bool
	InstallOperations() []CommandSpec
	AlreadyInstalledLabel() string
	HasCommand(name string) bool
}

// HomebrewFormulaPackage 是基于 Homebrew formula 的包实现。
type HomebrewFormulaPackage struct {
	spec ToolDefinition
}

// ID 返回工具唯一标识。
func (p HomebrewFormulaPackage) ID() string {
	return p.spec.ToolID
}

// Formula 返回对应 brew formula 名称。
func (p HomebrewFormulaPackage) Formula() string {
	return p.spec.Formula
}

// Commands 返回该工具支持的命令列表副本。
func (p HomebrewFormulaPackage) Commands() []Command {
	return append([]Command{}, p.spec.Commands...)
}

// IsInstallable 返回配置级可安装能力。
func (p HomebrewFormulaPackage) IsInstallable() bool {
	return p.spec.Installable
}

// InstallPolicy 返回安装策略，不存在时回退默认策略。
func (p HomebrewFormulaPackage) InstallPolicy() InstallPolicy {
	if p.spec.InstallPolicy == nil {
		return NewPackageInstallPolicy(NewPackageCheckPolicy(false, false), NewPackageCheckPolicy(false, false), NewPackageVersionPolicy(nil))
	}
	return p.spec.InstallPolicy
}

// InstallOperations 返回安装附加动作。
func (p HomebrewFormulaPackage) InstallOperations() []CommandSpec {
	return append([]CommandSpec{}, p.spec.ExtraInstallOps...)
}

// AlreadyInstalledLabel 返回“已安装”提示文本。
func (p HomebrewFormulaPackage) AlreadyInstalledLabel() string {
	return p.spec.AlreadyInstalled
}

// DisplayName 按命令名返回展示名，默认回退工具级别或原始命令。
func (p HomebrewFormulaPackage) DisplayName(command string) string {
	for _, cmd := range p.spec.Commands {
		if cmd.Name == command {
			return cmd.DisplayName
		}
	}
	if p.spec.DisplayNameOverride != "" {
		return p.spec.DisplayNameOverride
	}
	return command
}

// HasCommand 判断包是否声明过该命令。
func (p HomebrewFormulaPackage) HasCommand(name string) bool {
	for _, cmd := range p.spec.Commands {
		if cmd.Name == name {
			return true
		}
	}
	return false
}

// PackageManager 供目录层兼容的核心操作查询能力。
type PackageManager interface {
	SupportedTools() []string
	InstallableTools() []string
	ResolveCommandPackages(name string) []Package
}

type Tool struct {
	spec          Package
	requestedName string
}

// ToolLifecycle keeps the existing public API name for callers that still use it.
type ToolLifecycle = Tool

// ToolOrchestrator remains a compatibility alias for historical callers.
type ToolOrchestrator = Tool

// ID 返回工具ID，nil 时返回空串。
func (t Tool) ID() string {
	if t.spec == nil {
		return ""
	}
	return t.spec.ID()
}

// Formula 返回对应 formula，nil 时返回空串。
func (t Tool) Formula() string {
	if t.spec == nil {
		return ""
	}
	return t.spec.Formula()
}

// InstallPolicy 获取安装策略。
func (t Tool) InstallPolicy() InstallPolicy {
	if t.spec == nil {
		return nil
	}
	return t.spec.InstallPolicy()
}

// DisplayName 按工具上下文获取显示名。
func (t Tool) DisplayName(command string) string {
	if t.spec == nil {
		return command
	}
	return t.spec.DisplayName(command)
}

// Commands 返回命令列表。
func (t Tool) Commands() []Command {
	if t.spec == nil {
		return nil
	}
	return t.spec.Commands()
}

// InstallOperations 返回安装附加动作。
func (t Tool) InstallOperations() []CommandSpec {
	if t.spec == nil {
		return nil
	}
	return t.spec.InstallOperations()
}

// IsInstallable 判断是否允许安装。
func (t Tool) IsInstallable() bool {
	if t.spec == nil {
		return false
	}
	return t.spec.IsInstallable()
}

// AlreadyInstalledLabel 返回用户提示文案。
func (t Tool) AlreadyInstalledLabel() string {
	if t.spec == nil {
		return ""
	}
	return t.spec.AlreadyInstalledLabel()
}

// HasCommand 检查请求命令是否属于当前工具。
func (t Tool) HasCommand(name string) bool {
	if t.spec == nil {
		return false
	}
	return t.spec.HasCommand(name)
}

// RequestedName 解析真实请求名，优先使用 requestedName。
func (t Tool) RequestedName() string {
	if t.requestedName != "" {
		return t.requestedName
	}
	return t.ID()
}

// IsInstalled 查询工具安装状态与路径归属。
func (t Tool) IsInstalled(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return ToolInstallStateWithCatalog(rt, t.catalogRef(catalog), pathPolicy, t.RequestedName())
}

// PlanInstall 按策略参数构建安装队列。
func (t Tool) PlanInstall(rt Runtime, catalog *toolCatalog, force bool, checkCommand bool, checkFormula bool) (InstallQueue, error) {
	if !t.IsInstallable() {
		return nil, fmt.Errorf("unsupported tool: %s", t.ID())
	}
	if t.ID() == "" {
		return nil, fmt.Errorf("unsupported package")
	}
	return t.BuildInstallQueueWithPolicy(rt, t.catalogRef(catalog), force, checkCommand, checkFormula)
}

// PlanInstallByPolicy 使用工具声明的 plan 检查策略。
func (t Tool) PlanInstallByPolicy(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
	policy := resolveCheckPolicy(resolveInstallPolicy(t.InstallPolicy()).PlanCheckPolicy())
	return t.PlanInstall(rt, catalog, force, policy.CheckCommand(), policy.CheckFormula())
}

// PlanInstallByRunPolicy 使用工具声明的运行期检查策略。
func (t Tool) PlanInstallByRunPolicy(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
	policy := resolveCheckPolicy(resolveInstallPolicy(t.InstallPolicy()).RunCheckPolicy())
	return t.PlanInstall(rt, catalog, force, policy.CheckCommand(), policy.CheckFormula())
}

// BuildInstallQueueWithPolicy 根据传入策略执行存在性和 brew 状态检查。
func (t Tool) BuildInstallQueueWithPolicy(rt Runtime, catalog *toolCatalog, force bool, checkCommand bool, checkFormula bool) (InstallQueue, error) {
	catalog = t.catalogRef(catalog)
	rt = NormalizeRuntime(rt)
	if !IsBrewInstalled(rt) {
		return nil, fmt.Errorf("homebrew is not installed")
	}

	if !force && checkCommand && t.spec != nil {
		if packageHasCommand(rt, t.spec) {
			return nil, nil
		}
	}

	if !force && checkFormula {
		formula, ok := catalog.brewFormulaForTool(t.ID())
		if !ok {
			return nil, fmt.Errorf("unsupported tool: %s", t.ID())
		}

		installedByFormula, err := IsBrewFormulaInstalled(rt, formula)
		if err != nil {
			return nil, fmt.Errorf("check %s install status failed: %w", t.ID(), err)
		}
		if installedByFormula {
			return nil, nil
		}
	}

	// 未命中任何“已安装”保护后，返回标准安装动作序列
	specs, ok := catalog.installOperationSequence(t.ID())
	if !ok {
		return nil, fmt.Errorf("unsupported tool: %s", t.ID())
	}

	return InstallQueueFromSpecs(specs), nil
}

// ResolveVersion 获取当前版本或最新版本。
func (t Tool) ResolveVersion(rt Runtime, catalog *toolCatalog, commandPath string, useLatest bool) (string, error) {
	catalog = t.catalogRef(catalog)
	if useLatest {
		return ToolLatestVersionWithCatalog(rt, catalog, t.RequestedName())
	}
	return ToolVersionWithPathWithCatalog(rt, catalog, t.RequestedName(), commandPath)
}

// ResolveOutdatedCurrentVersion 获取用于过期比较的当前版本。
func (t Tool) ResolveOutdatedCurrentVersion(rt Runtime, catalog *toolCatalog) (string, error) {
	return ToolVersionForOutdatedWithCatalog(rt, t.catalogRef(catalog), t.RequestedName())
}

// ResolveLatestVersion 获取用于过期比较的最新版本。
func (t Tool) ResolveLatestVersion(rt Runtime, catalog *toolCatalog) (string, error) {
	return ToolLatestVersionWithCatalog(rt, t.catalogRef(catalog), t.RequestedName())
}

// catalogRef 为 nil 时返回默认 catalog。
func (t Tool) catalogRef(catalog *toolCatalog) *toolCatalog {
	if catalog == nil {
		return NewToolCatalog()
	}
	return catalog
}

// BuildUpdateToolName 返回更新动作使用的工具名（npm 例外保持 npm 本体）。
func (t Tool) BuildUpdateToolName() string {
	if t.RequestedName() == "npm" {
		return "npm"
	}
	return t.Formula()
}

// toolCatalog 聚合工具查找、查询和命令反查索引。
type toolCatalog struct {
	// packageLookup 按工具ID->package
	packageLookup       map[string]Package
	// commandToPackages 按命令名->package 列表
	commandToPackages   map[string][]Package
	// listedTools 供 list 命令展示顺序
	listedTools         []string
	// installablePackages 供 install 相关命令顺序
	installablePackages []string
}

// NewToolCatalog 使用默认定义构建目录。
func NewToolCatalog() *toolCatalog {
	return NewToolCatalogWithDefinitions(defaultManagedToolDefinitions())
}

// InstallLongHelp 生成 install 命令的帮助文案。
func InstallLongHelp() string {
	lines := []string{
		"Install all supported developer tools.",
		"Supported tools:",
	}
	for _, line := range installableToolHelpLines() {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// installableToolHelpLines 输出 install 命令帮助项（按 InstallOrder 排序）。
func installableToolHelpLines() []string {
	helpDefinitions := []ToolDefinition{}
	for _, definition := range defaultManagedToolDefinitions() {
		if !definition.Installable {
			continue
		}
		helpDefinitions = append(helpDefinitions, definition)
	}
	sort.SliceStable(helpDefinitions, func(i, j int) bool {
		return helpDefinitions[i].InstallOrder < helpDefinitions[j].InstallOrder
	})
	lines := make([]string, 0, len(helpDefinitions))
	for _, definition := range helpDefinitions {
		lines = append(lines, fmt.Sprintf("- %s -> brew install %s", definition.ToolID, definition.Formula))
	}
	return lines
}

// cmd 构造 command spec 的便捷函数。
func cmd(name string, args ...string) CommandSpec {
	return NewCommandSpec(name, args...)
}

// command 构造 display name 与真实名映射。
func command(name, displayName string) Command {
	if displayName == "" {
		displayName = name
	}
	return Command{Name: name, DisplayName: displayName}
}

// NewToolCatalogWithDefinitions 根据传入定义构建目录。
func NewToolCatalogWithDefinitions(definitions []ToolDefinition) *toolCatalog {
	copied := make([]ToolDefinition, len(definitions))
	copy(copied, definitions)
	return newToolCatalog(copied)
}

// defaultManagedTools 返回默认工具对象切片。
func defaultManagedTools() []Package {
	pkgs := make([]Package, 0, len(defaultManagedToolDefinitions()))
	for _, definition := range defaultManagedToolDefinitions() {
		pkgs = append(pkgs, HomebrewFormulaPackage{spec: definition})
	}
	return pkgs
}

// defaultManagedToolDefinitions 定义内置工具与安装元数据。
func defaultManagedToolDefinitions() []ToolDefinition {
	defaultInstallPolicy := func(planCheckCommand, planCheckFormula, runCheckCommand, runCheckFormula bool, versionArgs []string) InstallPolicy {
		return NewPackageInstallPolicy(
			NewPackageCheckPolicy(planCheckCommand, planCheckFormula),
			NewPackageCheckPolicy(runCheckCommand, runCheckFormula),
			NewPackageVersionPolicy(versionArgs),
		)
	}
	return []ToolDefinition{
		{
			ToolID:           "php",
			Formula:          "php",
			Commands:         []Command{command("php", "php")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "php is already installed",
			ListOrder:        1,
			InstallOrder:     1,
		},
		{
			ToolID:           "python3",
			Formula:          "python3",
			Commands:         []Command{command("python3", "python3")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, true, false, true, []string{"--version"}),
			AlreadyInstalled: "python3 is already installed by homebrew",
			ListOrder:        2,
			InstallOrder:     2,
		},
		{
			ToolID:           "node",
			Formula:          "node",
			Commands:         []Command{command("node", "node"), command("npm", "npm")},
			ListCommands:     []string{"node"},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "node is already installed",
			ListOrder:        3,
			InstallOrder:     3,
		},
		{
			ToolID:           "go",
			Formula:          "go",
			Commands:         []Command{command("go", "go")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"version"}),
			AlreadyInstalled: "go is already installed",
			ListOrder:        4,
			InstallOrder:     4,
		},
		{
			ToolID:           "npm",
			Formula:          "node",
			Commands:         []Command{command("npm", "npm")},
			Installable:      false,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "npm is already provided by node",
			ListOrder:        5,
		},
		{
			ToolID:           "curl",
			Formula:          "curl",
			Commands:         []Command{command("curl", "curl")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "curl is already installed",
			ExtraInstallOps:  []CommandSpec{cmd("brew", "link", "curl", "--force")},
			ListOrder:        6,
			InstallOrder:     5,
		},
		{
			ToolID:           "gh",
			Formula:          "gh",
			Commands:         []Command{command("gh", "gh")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "gh is already installed",
			ListOrder:        7,
			InstallOrder:     9,
		},
		{
			ToolID:           "git",
			Formula:          "git",
			Commands:         []Command{command("git", "git")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "git is already installed",
			ListOrder:        8,
			InstallOrder:     6,
		},
		{
			ToolID:           "ffmpeg",
			Formula:          "ffmpeg",
			Commands:         []Command{command("ffmpeg", "ffmpeg")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"-version"}),
			AlreadyInstalled: "ffmpeg is already installed",
			ListOrder:        9,
			InstallOrder:     7,
		},
		{
			ToolID:           "tree",
			Formula:          "tree",
			Commands:         []Command{command("tree", "tree")},
			Installable:      true,
			InstallPolicy:    defaultInstallPolicy(true, false, true, false, []string{"--version"}),
			AlreadyInstalled: "tree is already installed",
			ListOrder:        10,
			InstallOrder:     8,
		},
	}
}

// newToolCatalog 创建包含有序索引的内部目录模型。
func newToolCatalog(definitions []ToolDefinition) *toolCatalog {
	listedDefinitions := orderedToolDefinitions(definitions, func(d ToolDefinition) int {
		return d.ListOrder
	})
	installDefinitions := orderedToolDefinitions(definitions, func(d ToolDefinition) int {
		return d.InstallOrder
	})

	packageLookup := make(map[string]Package, len(definitions))
	commandLookup := make(map[string][]Package, len(definitions))
	listed := make([]string, 0, len(definitions))
	installable := make([]string, 0, len(definitions))
	listedSet := make(map[string]bool, len(definitions))
	installableSet := make(map[string]bool, len(definitions))

	packages := make([]Package, 0, len(definitions))
	for _, definition := range definitions {
		pkg := HomebrewFormulaPackage{spec: definition}
		packageLookup[pkg.ID()] = pkg
		packages = append(packages, pkg)
		for _, c := range pkg.Commands() {
			commandLookup[c.Name] = append(commandLookup[c.Name], pkg)
		}
	}

	for _, definition := range listedDefinitions {
		for _, name := range definition.listedCommands() {
			if _, seen := listedSet[name]; seen {
				continue
			}
			listed = append(listed, name)
			listedSet[name] = true
		}
	}

	for _, definition := range installDefinitions {
		item := HomebrewFormulaPackage{spec: definition}
		if !item.IsInstallable() || installableSet[item.ID()] {
			continue
		}
		installable = append(installable, item.ID())
		installableSet[item.ID()] = true
	}

	return &toolCatalog{
		packageLookup:       packageLookup,
		commandToPackages:   commandLookup,
		listedTools:         listed,
		installablePackages: installable,
	}
}

// orderedToolDefinitions 按指定排序字段返回副本（稳定排序）。
func orderedToolDefinitions(definitions []ToolDefinition, orderBy func(ToolDefinition) int) []ToolDefinition {
	ordered := append([]ToolDefinition(nil), definitions...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		return orderBy(ordered[i]) < orderBy(ordered[j])
	})
	return ordered
}

// managedToolsFor 先按命令映射找，再按工具ID回退。
func (c *toolCatalog) managedToolsFor(name string) ([]Package, bool) {
	if item, ok := c.commandToPackages[name]; ok {
		return append([]Package{}, item...), true
	}
	if item, ok := c.packageLookup[name]; ok {
		return []Package{item}, true
	}
	return nil, false
}

// managedToolFor 取某命令/工具的首个匹配 package。
func (c *toolCatalog) managedToolFor(name string) (Package, bool) {
	items, ok := c.managedToolsFor(name)
	if !ok || len(items) == 0 {
		return nil, false
	}
	return items[0], true
}

// managedToolsForCommand 返回命令对应 package 列表。
func (c *toolCatalog) managedToolsForCommand(name string) []Package {
	items, ok := c.managedToolsFor(name)
	if !ok {
		return nil
	}
	return append([]Package{}, items...)
}

func (c *toolCatalog) ResolveCommandPackages(name string) []Package {
	return c.managedToolsForCommand(name)
}

// managedToolByCommand 获取第一个命令映射对象，供多数流程复用。
func (c *toolCatalog) managedToolByCommand(name string) (Package, bool) {
	return c.managedToolFor(name)
}

// installOperationSequence 将工具安装动作组装为命令列表（先安装 formula，再执行附加动作）。
func (c *toolCatalog) installOperationSequence(name string) ([]CommandSpec, bool) {
	tool, ok := c.managedToolByCommand(name)
	if !ok {
		return nil, false
	}
	ops := make([]CommandSpec, 0, 1+len(tool.InstallOperations()))
	ops = append(ops, cmd("brew", "install", tool.Formula()))
	ops = append(ops, tool.InstallOperations()...)
	return ops, true
}

// brewFormulaForTool 获取工具关联的 formula。
func (c *toolCatalog) brewFormulaForTool(name string) (string, bool) {
	tool, ok := c.managedToolByCommand(name)
	if !ok {
		return "", false
	}
	return tool.Formula(), true
}

// versionArgsForTool 获取版本参数，缺省返回 false。
func (c *toolCatalog) versionArgsForTool(name string) ([]string, bool) {
	tool, ok := c.managedToolByCommand(name)
	if !ok {
		return nil, false
	}
	policy := tool.InstallPolicy()
	if policy == nil {
		return nil, false
	}

	versionPolicy := policy.VersionPolicy()
	if versionPolicy == nil {
		return nil, false
	}
	return versionPolicy.VersionArgs(), true
}

// toolDisplayName 返回工具展示名。
func (c *toolCatalog) toolDisplayName(name string) (string, bool) {
	tool, ok := c.managedToolByCommand(name)
	if !ok {
		return "", false
	}
	return tool.DisplayName(name), true
}

// managedToolIsInstallable 检查是否存在匹配且可安装的工具。
func (c *toolCatalog) managedToolIsInstallable(name string) bool {
	tools, ok := c.managedToolsFor(name)
	if !ok {
		return false
	}
	for _, tool := range tools {
		if tool.IsInstallable() && tool.ID() == name {
			return true
		}
	}
	return false
}

// toolSpec 获取工具规格。
func (c *toolCatalog) toolSpec(name string) (Package, bool) {
	return c.managedToolByCommand(name)
}

// toolLifecycle 返回 tool 生命周期对象。
func (c *toolCatalog) toolLifecycle(name string) (ToolLifecycle, bool) {
	if c == nil {
		c = NewToolCatalog()
	}
	spec, ok := c.toolSpec(name)
	return ToolLifecycle{spec: spec, requestedName: name}, ok
}

// listedToolsCatalog 返回 listedTools 副本。
func (c *toolCatalog) listedToolsCatalog() []string {
	if c == nil {
		return nil
	}
	out := make([]string, len(c.listedTools))
	copy(out, c.listedTools)
	return out
}

// installableToolsCatalog 返回 installablePackages 副本。
func (c *toolCatalog) installableToolsCatalog() []string {
	if c == nil {
		return nil
	}
	out := make([]string, len(c.installablePackages))
	copy(out, c.installablePackages)
	return out
}
