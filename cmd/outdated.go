// cmd/outdated.go 实现 outdated 命令：扫描工具过期状态并以可选格式展示检查结果。
package cmd

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/cuimingda/denv-cli/internal"
	"github.com/spf13/cobra"
)

// NewOutdatedCmd 使用默认上下文创建 outdated 命令。
func NewOutdatedCmd() *cobra.Command {
	ctx := NewCLIContext()
	return NewOutdatedCmdWithService(outdatedCommandService{
		supportedTools:          ctx.RuntimeContext.SupportedTools,
		outdatedChecks:          ctx.CatalogContext.OutdatedChecks,
		outdatedCheckWithOutput: ctx.CatalogContext.OutdatedCheckWithOutput,
		runBrewUpdate:           ctx.CatalogContext.RunBrewUpdate,
	})
}

// NewOutdatedCmdWithService 使用已注入服务构建命令（便于测试替身）。
func NewOutdatedCmdWithService(svc OutdatedCommandService) *cobra.Command {
	if svc == nil {
		panic("outdated command requires a non-nil service implementation")
	}

	var parallel int
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

			workers, err := validateOutdatedParallel(cmd, parallel)
			if err != nil {
				return err
			}

			doingf(cmd, "brew update starting")
			brewConsole := newOutdatedProgressConsole(cmd.ErrOrStderr(), []string{"brew update - pending"})
			brewWriter := newOutdatedLineWriter(brewConsole, 0, "brew update")
			if err := svc.RunBrewUpdate(brewWriter); err != nil {
				brewWriter.Finish("brew update failed: " + err.Error())
				return err
			}
			brewWriter.Finish("brew update completed")
			doingf(cmd, "brew update finished")

			targets := svc.SupportedTools()
			doingf(cmd, "queued %d commands for outdated checks", len(targets))
			verbosef(cmd, "outdated targets: %s", strings.Join(targets, ", "))

			if workers > 1 {
				doingf(cmd, "running outdated checks with parallel=%d", workers)
			} else {
				doingf(cmd, "running outdated checks sequentially")
			}

			rows, err := runOutdatedChecks(cmd.ErrOrStderr(), svc, targets, workers)
			if err != nil {
				return err
			}
			doingf(cmd, "outdated check completed in %s", time.Since(start))

			return NewOutdatedPresenter(mode, rows, colorOutput).Render(out)
		},
	}

	cmd.Flags().String("output", string(listOutputPlain), "output format: plain|json|table|no-color")
	cmd.Flags().IntVar(&parallel, "parallel", 1, "run outdated checks in parallel (optional value defaults to 4; allowed: 2-8)")
	if flag := cmd.Flags().Lookup("parallel"); flag != nil {
		flag.NoOptDefVal = "4"
	}
	return cmd
}

type outdatedCommandService struct {
	supportedTools          func() []string
	outdatedChecks          func() ([]denv.ToolCheckResult, error)
	outdatedCheckWithOutput func(io.Writer, string) (denv.ToolCheckResult, error)
	runBrewUpdate           func(io.Writer) error
}

// SupportedTools 读取当前 runtime 下支持的工具清单。
func (s outdatedCommandService) SupportedTools() []string {
	return s.supportedTools()
}

// OutdatedChecks 委托到服务层返回过期检测结果。
func (s outdatedCommandService) OutdatedChecks() ([]denv.ToolCheckResult, error) {
	return s.outdatedChecks()
}

// OutdatedCheckWithOutput 委托到服务层返回单工具过期检测结果，并输出过程日志。
func (s outdatedCommandService) OutdatedCheckWithOutput(out io.Writer, name string) (denv.ToolCheckResult, error) {
	return s.outdatedCheckWithOutput(out, name)
}

// RunBrewUpdate 先刷新 brew 索引，再继续过期检测流程。
func (s outdatedCommandService) RunBrewUpdate(out io.Writer) error {
	return s.runBrewUpdate(out)
}

func validateOutdatedParallel(cmd *cobra.Command, value int) (int, error) {
	if cmd == nil || !cmd.Flags().Changed("parallel") {
		return 1, nil
	}
	if value < 2 || value > 8 {
		return 0, fmt.Errorf("--parallel must be between 2 and 8")
	}
	return value, nil
}

func runOutdatedChecks(errOut io.Writer, svc OutdatedCommandService, targets []string, workers int) ([]denv.ToolCheckResult, error) {
	if len(targets) == 0 {
		return nil, nil
	}

	if workers < 1 {
		workers = 1
	}
	if workers > len(targets) {
		workers = len(targets)
	}

	initialLines := make([]string, 0, len(targets))
	for _, target := range targets {
		initialLines = append(initialLines, target+" - pending")
	}
	console := newOutdatedProgressConsole(errOut, initialLines)

	type taskResult struct {
		index int
		row   denv.ToolCheckResult
		err   error
	}

	jobs := make(chan int)
	results := make(chan taskResult, len(targets))
	var wg sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				name := targets[index]
				lineWriter := newOutdatedLineWriter(console, index, name)
				lineWriter.Add("start")

				row, err := svc.OutdatedCheckWithOutput(lineWriter, name)
				if err != nil {
					lineWriter.Finish("failed: " + err.Error())
					results <- taskResult{index: index, err: err}
					continue
				}

				lineWriter.Finish(renderOutdatedLine(row, false))
				results <- taskResult{index: index, row: row}
			}
		}()
	}

	go func() {
		for index := range targets {
			jobs <- index
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	rows := make([]denv.ToolCheckResult, len(targets))
	var firstErr error
	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		rows[result.index] = result.row
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return rows, nil
}
