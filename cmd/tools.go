// cmd/tools.go 定义 CLI 层服务接口与上下文组装，屏蔽命令与底层 Service 的直接耦合并提供默认依赖注入。
package cmd

import (
	"io"

	"github.com/cuimingda/denv-cli/internal"
)

// Command-layer contracts are intentionally narrow, aligned with service ports.
// ListCommandService 仅包含列表命令所需能力。
type ListCommandService interface {
	ListToolItems(opts denv.ListOptions) ([]denv.ToolListItem, error)
}

// InstallCommandService 仅包含安装命令所需能力。
type InstallCommandService interface {
	BuildInstallQueue(force bool) (denv.InstallQueue, error)
	ExecuteInstallQueue(out io.Writer, queue denv.InstallQueue) error
}

// OutdatedCommandService 仅包含 outdated 命令所需能力。
type OutdatedCommandService interface {
	SupportedTools() []string
	OutdatedChecks() ([]denv.ToolCheckResult, error)
}

// UpdateCommandService 仅包含 update 命令所需能力。
type UpdateCommandService interface {
	SupportedTools() []string
	OutdatedUpdatePlan() ([]denv.OutdatedItem, error)
	UpdateToolWithOutput(out io.Writer, name string) error
}

type CLIContext struct {
	service        *denv.Service
	RuntimeContext denv.RuntimeContext
	CatalogContext denv.CatalogContext
	InstallContext denv.InstallContext
	UpdateContext  denv.UpdateContext
}

// NewCLIContext 使用默认 Runtime 创建上下文。
func NewCLIContext() *CLIContext {
	return NewCLIContextWithRuntime(denv.Runtime{})
}

// NewCLIContextWithRuntime 注入运行时依赖，主要用于测试。
func NewCLIContextWithRuntime(rt denv.Runtime) *CLIContext {
	service := denv.NewService(rt)
	return &CLIContext{
		service:        service,
		RuntimeContext: service,
		CatalogContext: service,
		InstallContext: service,
		UpdateContext:  service,
	}
}

// ensureCLIContext 做空保护，避免调用方未传 context 时 panic。
func ensureCLIContext(ctx *CLIContext) *CLIContext {
	if ctx == nil {
		return NewCLIContext()
	}
	return ctx
}
