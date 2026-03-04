// internal defines the package-level contract used by command and service layers.
package denv

import "github.com/cuimingda/denv-cli/internal/infra"

// Runtime 复用 infra.Runtime，作为对命令层暴露的运行时契约。
type Runtime = infra.Runtime

// NormalizeRuntime 交由 infra 层提供默认运行时行为，保持兼容调用入口不变。
func NormalizeRuntime(rt Runtime) Runtime {
	return infra.NormalizeRuntime(rt)
}
