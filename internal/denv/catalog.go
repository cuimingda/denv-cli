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

func (d ToolDefinition) listedCommands() []string {
	if len(d.ListCommands) > 0 {
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

func ToolDefinitions() []ToolDefinition {
	definitions := defaultManagedToolDefinitions()
	out := make([]ToolDefinition, len(definitions))
	copy(out, definitions)
	return out
}

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

type CheckPolicy interface {
	CheckCommand() bool
	CheckFormula() bool
}

type VersionPolicy interface {
	VersionArgs() []string
}

type InstallPolicy interface {
	PlanCheckPolicy() CheckPolicy
	RunCheckPolicy() CheckPolicy
	VersionPolicy() VersionPolicy
}

type packageCheckPolicy struct {
	checkCommand bool
	checkFormula bool
}

func (p packageCheckPolicy) CheckCommand() bool {
	return p.checkCommand
}

func (p packageCheckPolicy) CheckFormula() bool {
	return p.checkFormula
}

type packageVersionPolicy struct {
	versionArgs []string
}

func (p packageVersionPolicy) VersionArgs() []string {
	return append([]string{}, p.versionArgs...)
}

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

func NewPackageCheckPolicy(checkCommand bool, checkFormula bool) CheckPolicy {
	return packageCheckPolicy{checkCommand: checkCommand, checkFormula: checkFormula}
}

func NewPackageVersionPolicy(versionArgs []string) VersionPolicy {
	return packageVersionPolicy{versionArgs: versionArgs}
}

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

type HomebrewFormulaPackage struct {
	spec ToolDefinition
}

func (p HomebrewFormulaPackage) ID() string {
	return p.spec.ToolID
}

func (p HomebrewFormulaPackage) Formula() string {
	return p.spec.Formula
}

func (p HomebrewFormulaPackage) Commands() []Command {
	return append([]Command{}, p.spec.Commands...)
}

func (p HomebrewFormulaPackage) IsInstallable() bool {
	return p.spec.Installable
}

func (p HomebrewFormulaPackage) InstallPolicy() InstallPolicy {
	if p.spec.InstallPolicy == nil {
		return NewPackageInstallPolicy(NewPackageCheckPolicy(false, false), NewPackageCheckPolicy(false, false), NewPackageVersionPolicy(nil))
	}
	return p.spec.InstallPolicy
}

func (p HomebrewFormulaPackage) InstallOperations() []CommandSpec {
	return append([]CommandSpec{}, p.spec.ExtraInstallOps...)
}

func (p HomebrewFormulaPackage) AlreadyInstalledLabel() string {
	return p.spec.AlreadyInstalled
}

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

func (p HomebrewFormulaPackage) HasCommand(name string) bool {
	for _, cmd := range p.spec.Commands {
		if cmd.Name == name {
			return true
		}
	}
	return false
}

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

func (t Tool) ID() string {
	if t.spec == nil {
		return ""
	}
	return t.spec.ID()
}

func (t Tool) Formula() string {
	if t.spec == nil {
		return ""
	}
	return t.spec.Formula()
}

func (t Tool) InstallPolicy() InstallPolicy {
	if t.spec == nil {
		return nil
	}
	return t.spec.InstallPolicy()
}

func (t Tool) DisplayName(command string) string {
	if t.spec == nil {
		return command
	}
	return t.spec.DisplayName(command)
}

func (t Tool) Commands() []Command {
	if t.spec == nil {
		return nil
	}
	return t.spec.Commands()
}

func (t Tool) InstallOperations() []CommandSpec {
	if t.spec == nil {
		return nil
	}
	return t.spec.InstallOperations()
}

func (t Tool) IsInstallable() bool {
	if t.spec == nil {
		return false
	}
	return t.spec.IsInstallable()
}

func (t Tool) AlreadyInstalledLabel() string {
	if t.spec == nil {
		return ""
	}
	return t.spec.AlreadyInstalledLabel()
}

func (t Tool) HasCommand(name string) bool {
	if t.spec == nil {
		return false
	}
	return t.spec.HasCommand(name)
}

func (t Tool) RequestedName() string {
	if t.requestedName != "" {
		return t.requestedName
	}
	return t.ID()
}

func (t Tool) IsInstalled(rt Runtime, catalog *toolCatalog, pathPolicy PathPolicy) (installed bool, commandPath string, installedByHomebrew bool, err error) {
	return ToolInstallStateWithCatalog(rt, t.catalogRef(catalog), pathPolicy, t.RequestedName())
}

func (t Tool) PlanInstall(rt Runtime, catalog *toolCatalog, force bool, checkCommand bool, checkFormula bool) (InstallQueue, error) {
	if !t.IsInstallable() {
		return nil, fmt.Errorf("unsupported tool: %s", t.ID())
	}
	if t.ID() == "" {
		return nil, fmt.Errorf("unsupported package")
	}
	return t.BuildInstallQueueWithPolicy(rt, t.catalogRef(catalog), force, checkCommand, checkFormula)
}

func (t Tool) PlanInstallByPolicy(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
	policy := resolveCheckPolicy(resolveInstallPolicy(t.InstallPolicy()).PlanCheckPolicy())
	return t.PlanInstall(rt, catalog, force, policy.CheckCommand(), policy.CheckFormula())
}

func (t Tool) PlanInstallByRunPolicy(rt Runtime, catalog *toolCatalog, force bool) (InstallQueue, error) {
	policy := resolveCheckPolicy(resolveInstallPolicy(t.InstallPolicy()).RunCheckPolicy())
	return t.PlanInstall(rt, catalog, force, policy.CheckCommand(), policy.CheckFormula())
}

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

	specs, ok := catalog.installOperationSequence(t.ID())
	if !ok {
		return nil, fmt.Errorf("unsupported tool: %s", t.ID())
	}

	return InstallQueueFromSpecs(specs), nil
}

func (t Tool) ResolveVersion(rt Runtime, catalog *toolCatalog, commandPath string, useLatest bool) (string, error) {
	catalog = t.catalogRef(catalog)
	if useLatest {
		return ToolLatestVersionWithCatalog(rt, catalog, t.RequestedName())
	}
	return ToolVersionWithPathWithCatalog(rt, catalog, t.RequestedName(), commandPath)
}

func (t Tool) ResolveOutdatedCurrentVersion(rt Runtime, catalog *toolCatalog) (string, error) {
	return ToolVersionForOutdatedWithCatalog(rt, t.catalogRef(catalog), t.RequestedName())
}

func (t Tool) ResolveLatestVersion(rt Runtime, catalog *toolCatalog) (string, error) {
	return ToolLatestVersionWithCatalog(rt, t.catalogRef(catalog), t.RequestedName())
}

func (t Tool) catalogRef(catalog *toolCatalog) *toolCatalog {
	if catalog == nil {
		return NewToolCatalog()
	}
	return catalog
}

func (t Tool) BuildUpdateToolName() string {
	if t.RequestedName() == "npm" {
		return "npm"
	}
	return t.Formula()
}

type toolCatalog struct {
	packageLookup       map[string]Package
	commandToPackages   map[string][]Package
	listedTools         []string
	installablePackages []string
}

func NewToolCatalog() *toolCatalog {
	return NewToolCatalogWithDefinitions(defaultManagedToolDefinitions())
}

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

func cmd(name string, args ...string) CommandSpec {
	return NewCommandSpec(name, args...)
}

func command(name, displayName string) Command {
	if displayName == "" {
		displayName = name
	}
	return Command{Name: name, DisplayName: displayName}
}

func NewToolCatalogWithDefinitions(definitions []ToolDefinition) *toolCatalog {
	copied := make([]ToolDefinition, len(definitions))
	copy(copied, definitions)
	return newToolCatalog(copied)
}

func defaultManagedTools() []Package {
	pkgs := make([]Package, 0, len(defaultManagedToolDefinitions()))
	for _, definition := range defaultManagedToolDefinitions() {
		pkgs = append(pkgs, HomebrewFormulaPackage{spec: definition})
	}
	return pkgs
}

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

func orderedToolDefinitions(definitions []ToolDefinition, orderBy func(ToolDefinition) int) []ToolDefinition {
	ordered := append([]ToolDefinition(nil), definitions...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		return orderBy(ordered[i]) < orderBy(ordered[j])
	})
	return ordered
}

func (c *toolCatalog) managedToolsFor(name string) ([]Package, bool) {
	if item, ok := c.commandToPackages[name]; ok {
		return append([]Package{}, item...), true
	}
	if item, ok := c.packageLookup[name]; ok {
		return []Package{item}, true
	}
	return nil, false
}

func (c *toolCatalog) managedToolFor(name string) (Package, bool) {
	items, ok := c.managedToolsFor(name)
	if !ok || len(items) == 0 {
		return nil, false
	}
	return items[0], true
}

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

func (c *toolCatalog) managedToolByCommand(name string) (Package, bool) {
	return c.managedToolFor(name)
}

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

func (c *toolCatalog) brewFormulaForTool(name string) (string, bool) {
	tool, ok := c.managedToolByCommand(name)
	if !ok {
		return "", false
	}
	return tool.Formula(), true
}

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

func (c *toolCatalog) toolDisplayName(name string) (string, bool) {
	tool, ok := c.managedToolByCommand(name)
	if !ok {
		return "", false
	}
	return tool.DisplayName(name), true
}

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

func (c *toolCatalog) toolSpec(name string) (Package, bool) {
	return c.managedToolByCommand(name)
}

func (c *toolCatalog) toolLifecycle(name string) (ToolLifecycle, bool) {
	if c == nil {
		c = NewToolCatalog()
	}
	spec, ok := c.toolSpec(name)
	return ToolLifecycle{spec: spec, requestedName: name}, ok
}

func (c *toolCatalog) listedToolsCatalog() []string {
	if c == nil {
		return nil
	}
	out := make([]string, len(c.listedTools))
	copy(out, c.listedTools)
	return out
}

func (c *toolCatalog) installableToolsCatalog() []string {
	if c == nil {
		return nil
	}
	out := make([]string, len(c.installablePackages))
	copy(out, c.installablePackages)
	return out
}
