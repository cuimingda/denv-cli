package cmd

import (
	"time"

	"github.com/cuimingda/denv-cli/internal"
	"github.com/spf13/cobra"
)

// NewOutdatedCmd 使用默认上下文创建 outdated 命令。
func NewOutdatedCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewOutdatedCmdWithService(outdatedCommandService{
		supportedTools: ctx.RuntimeContext.SupportedTools,
		outdatedChecks: ctx.CatalogContext.OutdatedChecks,
	})
}

// NewOutdatedCmdWithService 使用已注入服务构建命令（便于测试替身）。
func NewOutdatedCmdWithService(svc OutdatedCommandService) *cobra.Command {
	if svc == nil {
		panic("outdated command requires a non-nil service implementation")
	}

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated status for supported developer tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			outputMode, _ := cmd.Flags().GetString("output")
			mode, err := parseListOutput(outputMode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			colorOutput := useColorOutput(out) && mode != listOutputNoColor
			start := time.Now()
			// 先输出即时报错上下文信息，便于快速定位检查耗时
			doingf(cmd, "check outdated status for %d tools", len(svc.SupportedTools()))

			rows, err := svc.OutdatedChecks()
			if err != nil {
				return err
			}
			doingf(cmd, "outdated check completed in %s", time.Since(start))

			return NewOutdatedPresenter(mode, rows, colorOutput).Render(out)
		},
	}

	cmd.Flags().String("output", string(listOutputPlain), "output format: plain|json|table|no-color")
	return cmd
}

type outdatedCommandService struct {
	supportedTools func() []string
	outdatedChecks func() ([]denv.ToolCheckResult, error)
}

// SupportedTools 读取当前 runtime 下支持的工具清单。
func (s outdatedCommandService) SupportedTools() []string {
	return s.supportedTools()
}

// OutdatedChecks 委托到服务层返回过期检测结果。
func (s outdatedCommandService) OutdatedChecks() ([]denv.ToolCheckResult, error) {
	return s.outdatedChecks()
}
