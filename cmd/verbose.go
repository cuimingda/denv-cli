// cmd/verbose.go 提供统一的 verbose 日志打印与开关判定，统一命令执行过程中的信息输出行为。
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// isVerbose 读取当前命令链路上的 verbose 标记，支持：
// 1) 当前命令本地 flags（用于测试/子命令场景下的显式覆盖）
// 2) 根命令 PersistentFlags（推荐入口）
// 3) 根命令普通 Flags（兼容历史入口）
// 任何一层读取失败都降级到 false，保证默认静默。
func isVerbose(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}

	if flag := cmd.Flags().Lookup("verbose"); flag != nil {
		if value, err := cmd.Flags().GetBool("verbose"); err == nil {
			return value
		}
	}

	root := cmd.Root()
	if root == nil {
		return false
	}

	if flag := root.PersistentFlags().Lookup("verbose"); flag != nil {
		if value, err := root.PersistentFlags().GetBool("verbose"); err == nil {
			return value
		}
	}

	if flag := root.Flags().Lookup("verbose"); flag != nil {
		if value, err := root.Flags().GetBool("verbose"); err == nil {
			return value
		}
	}

	return false
}

// verbosef 只在 verbose 模式下输出细粒度日志，默认写 stderr，避免污染标准输出。
func verbosef(cmd *cobra.Command, format string, args ...any) {
	if !isVerbose(cmd) {
		return
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[verbose] "+format+"\n", args...)
}

// doingf 始终输出执行进度到标准错误流，适合展示命令主流程状态。
func doingf(cmd *cobra.Command, format string, args ...any) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "[INFO] "+format+"\n", args...)
}
