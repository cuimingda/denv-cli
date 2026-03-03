package denv

import "io"

// RuntimeContext 负责运行时发现与可用性探测能力。
type RuntimeContext interface {
	Discovery
}

// CatalogContext 负责工具清单、版本与过期查询能力。
type CatalogContext interface {
	VersionResolver
	OutdatedManager
}

// ListContext 只提供只读列表相关能力。
type ListContext interface {
	VersionResolver
}

// InstallContext 负责安装计划构建与安装执行能力。
type InstallContext interface {
	InstallPlanner
	InstallExecutor
}

// UpdateContext 负责过期发现与更新执行编排能力。
type UpdateContext interface {
	OutdatedManager
	UpdateManager
}

// ServiceContext 聚合 CLI 命令层可消费的完整能力集合。
type ServiceContext interface {
	RuntimeContext
	CatalogContext
	InstallContext
	UpdateContext
}

// Discovery 负责环境查询：支持工具清单、安装状态与brew状态检测。
type Discovery interface {
	SupportedTools() []string
	InstallableTools() []string
	IsInstallableTool(name string) bool
	ToolDisplayName(name string) string
	IsCommandAvailable(name string) bool
	ToolInstallState(name string) (installed bool, commandPath string, installedByHomebrew bool, err error)
	CommandPath(name string) (string, error)
	ResolveCommandPackages(name string) []Package
	ResolvedBrewBinaryPath(name string, formula string) (string, error)
	IsManagedByHomebrew(path string) bool
	IsBrewInstalled() bool
	IsBrewFormulaInstalled(formula string) (bool, error)
}

// InstallPlanner 按配置与运行时信息构建安装任务队列/计划。
type InstallPlanner interface {
	BuildInstallQueue(force bool) (InstallQueue, error)
	BuildInstallPlan(toolName string, options BuildInstallPlanOptions) (InstallPlan, error)
}

// InstallExecutor 执行构建好的安装队列。
type InstallExecutor interface {
	ExecuteInstallQueue(out io.Writer, queue InstallQueue) error
}

// VersionResolver 提供列表、版本与过期数据的解析能力。
type VersionResolver interface {
	ListToolItems(opts ListOptions) ([]ToolListItem, error)
	OutdatedItems() ([]OutdatedItem, error)
}

// OutdatedManager 负责过期检测流程与结果过滤。
type OutdatedManager interface {
	OutdatedChecks() ([]ToolCheckResult, error)
	OutdatedItems() ([]OutdatedItem, error)
	OutdatedUpdatePlan() ([]OutdatedItem, error)
}

// UpdateManager 负责生成更新候选并执行更新动作。
type UpdateManager interface {
	OutdatedUpdatePlan() ([]OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}
