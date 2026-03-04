// cmd/install.go 实现 install 命令：解析参数、构建安装队列、支持 dry-run，并委托服务层执行实际安装动作。
package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/cuimingda/denv-cli/internal"
	"github.com/spf13/cobra"
)

// NewInstallCmd 使用默认上下文创建 install 命令。
func NewInstallCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewInstallCmdWithService(ctx.InstallContext)
}

// installCommandService 是命令层对 install context 的薄封装。
type installCommandService struct {
	service denv.InstallContext
}

// BuildInstallQueue 委托给 service 构建安装队列。
func (s installCommandService) BuildInstallQueue(force bool) (denv.InstallQueue, error) {
	return s.service.BuildInstallQueue(force)
}

// ExecuteInstallQueue 委托给 service 执行安装队列。
func (s installCommandService) ExecuteInstallQueue(out io.Writer, queue denv.InstallQueue) error {
	return s.service.ExecuteInstallQueue(out, queue)
}

// NewInstallCmdWithService 组装 install 命令并绑定可替换实现，便于测试与复用。
func NewInstallCmdWithService(service denv.InstallContext) *cobra.Command {
	if service == nil {
		panic("install command requires a non-nil install planner")
	}

	longHelp := denv.InstallLongHelp()
	commandService := installCommandService{service: service}

	cmd := &cobra.Command{
		Use:     "install",
		Args:    cobra.NoArgs,
		Short:   "Install supported developer tools",
		Long:    longHelp,
		Example: "  denv install",
		RunE: func(cmd *cobra.Command, _ []string) error {
			force, _ := cmd.Flags().GetBool("force")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			operationStart := time.Now()

			doingf(cmd, "prepare install plan (force=%t, dry-run=%t)", force, dryRun)
			// 先计算动作清单，再决定是否执行或仅展示 dry-run 信息
			installQueue, err := commandService.BuildInstallQueue(force)
			if err != nil {
				return err
			}
			doingf(cmd, "prepared %d operations in %s", installQueue.Len(), time.Since(operationStart))

			if dryRun {
				doingf(cmd, "showing installation plan without execution")
				for _, operation := range installQueue.ToOperations() {
					if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Would run: %s\n", operation.String()); err != nil {
						return err
					}
				}
				doingf(cmd, "dry-run plan generation completed in %s", time.Since(operationStart))
				return nil
			}

			doingf(cmd, "start executing %d install operations", installQueue.Len())
			if err := commandService.ExecuteInstallQueue(cmd.OutOrStdout(), installQueue); err != nil {
				return err
			}

			verbosef(cmd, "install completed in %s", time.Since(operationStart))
			_, outErr := fmt.Fprintln(cmd.OutOrStdout(), "install done")
			return outErr
		},
	}

	cmd.Flags().Bool("force", false, "install even if the tool already exists")
	cmd.Flags().Bool("dry-run", false, "show planned install operations only")
	return cmd
}
