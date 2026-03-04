// cmd/list.go 实现 list 命令，处理参数与输出模式，并将结果委派给 presenter 渲染。
package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cuimingda/denv-cli/internal"
	"github.com/spf13/cobra"
)

const (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

type listOutputMode string

const (
	listOutputPlain   listOutputMode = "plain"
	listOutputJSON    listOutputMode = "json"
	listOutputTable   listOutputMode = "table"
	listOutputNoColor listOutputMode = "no-color"
)

// NewListCmd 使用默认服务构建 list 命令。
func NewListCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewListCmdWithService(ctx.CatalogContext)
}

// NewListCmdWithService 对 list 命令进行依赖注入，便于测试。
func NewListCmdWithService(svc ListCommandService) *cobra.Command {
	if svc == nil {
		panic("list command requires a non-nil service implementation")
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List supported developer tools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			showVersion, _ := cmd.Flags().GetBool("version")
			showPath, _ := cmd.Flags().GetBool("path")
			outputMode, _ := cmd.Flags().GetString("output")

			mode, err := parseListOutput(outputMode)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			colorOutput := useColorOutput(out) && mode != listOutputNoColor

			start := time.Now()
			doingf(cmd, "scan supported tools...")
			items, err := svc.ListToolItems(denv.ListOptions{
				ShowVersion: showVersion,
				ShowPath:    showPath,
			})
			if err != nil {
				return err
			}
			doingf(cmd, "list scan completed in %s", time.Since(start))

			presenter := NewListPresenter(mode, listRenderOptions{
				colorOutput: colorOutput,
				showVersion: showVersion,
				showPath:    showPath,
			}, items)
			return presenter.Render(cmd.OutOrStdout())
		},
	}

	cmd.Flags().Bool("version", false, "show versions for discovered tools")
	cmd.Flags().Bool("path", false, "show executable paths for discovered tools")
	cmd.Flags().String("output", string(listOutputPlain), "output format: plain|json|table|no-color")
	return cmd
}

// parseListOutput 校验并返回合法的输出模式。
func parseListOutput(raw string) (listOutputMode, error) {
	mode := listOutputMode(raw)
	switch mode {
	case listOutputPlain, listOutputJSON, listOutputTable, listOutputNoColor:
		return mode, nil
	}
	return "", fmt.Errorf("invalid output: %s", raw)
}

type listRenderOptions struct {
	colorOutput bool
	showVersion bool
	showPath    bool
}

// useColorOutput 当前输出是否是文件描述符（可安全上色）。
func useColorOutput(out io.Writer) bool {
	_, ok := out.(*os.File)
	return ok
}

// colorize 按 ANSI 颜色码包装文本。
func colorize(color string, text string) string {
	return color + text + colorReset
}
