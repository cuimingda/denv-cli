# internal/app

职责：
- 作为命令入口与领域服务之间的应用层边界。
- 协调运行时/领域输入和输出，不直接发起系统调用。

当前实现锚点：
- 入口与协调：`cmd` 通过 `cmd/tools.go` 访问 `internal/service.go`。
- 边界收敛：`internal/runtime.go`、`internal/path_policy.go`。

迁移说明：
- 当前应用层入口在 `internal/service.go`，仅保留轻量组装与上下文委派。
- 领域核心编排能力统一迁入 `internal/domain`（discovery/version/install/outdated/update/helpers），
  infra 辅助设施统一迁入 `internal/infra`（runtimeAdapter/catalogManager）。
