package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/cuimingda/denv-cli/internal"
	"github.com/spf13/cobra"
)

// NewUpdateCmd 使用默认上下文创建 update 命令。
func NewUpdateCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewUpdateCmdWithService(updateCommandService{
		supportedTools:       ctx.RuntimeContext.SupportedTools,
		outdatedUpdatePlan:   ctx.UpdateContext.OutdatedUpdatePlan,
		updateToolWithOutput: ctx.UpdateContext.UpdateToolWithOutput,
	})
}

// NewUpdateCmdWithService 组装 update 命令并绑定服务实现，支持外部测试桩注入。
func NewUpdateCmdWithService(svc UpdateCommandService) *cobra.Command {
	if svc == nil {
		panic("update command requires a non-nil service implementation")
	}

	return &cobra.Command{
		Use:   "update",
		Short: "Update outdated supported developer tools to latest versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			start := time.Now()
			doingf(cmd, "scan %d tools for updates", len(svc.SupportedTools()))
			updated := false
			candidates, err := svc.OutdatedUpdatePlan()
			if err != nil {
				return err
			}
			for _, item := range candidates {
				doingf(cmd, "updating %s", item.Name)
				updateStart := time.Now()
				if err := svc.UpdateToolWithOutput(cmd.OutOrStdout(), item.Name); err != nil {
					return err
				}
				verbosef(cmd, "%s update completed in %s", item.Name, time.Since(updateStart))
				updated = true
			}

			if !updated {
				verbosef(cmd, "no outdated tools found after %s", time.Since(start))
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "no updates available")
				return err
			}

			verbosef(cmd, "update completed in %s", time.Since(start))
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "done")
			return err
		},
	}
}

// updateCommandService 是 update 命令的服务抽象。
type updateCommandService struct {
	supportedTools       func() []string
	outdatedUpdatePlan   func() ([]denv.OutdatedItem, error)
	updateToolWithOutput func(io.Writer, string) error
}

// SupportedTools 返回可用工具名称列表。
func (s updateCommandService) SupportedTools() []string {
	return s.supportedTools()
}

// OutdatedUpdatePlan 返回待更新工具列表。
func (s updateCommandService) OutdatedUpdatePlan() ([]denv.OutdatedItem, error) {
	return s.outdatedUpdatePlan()
}

// UpdateToolWithOutput 调用服务层执行单工具更新。
func (s updateCommandService) UpdateToolWithOutput(out io.Writer, name string) error {
	return s.updateToolWithOutput(out, name)
}
