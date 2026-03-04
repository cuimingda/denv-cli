# internal/app

职责：
- 作为命令入口与领域服务之间的应用层边界。
- 协调运行时/领域输入和输出，不直接发起系统调用。

当前实现锚点：
- 入口与协调：`cmd` 通过 `cmd/tools.go` 访问应用服务。
- 边界收敛：`internal/runtime.go`、`internal/infra/runtime.go`、`internal/compare.go`、`internal/domain/compare.go`、`internal/path_policy.go`。

迁移说明：
- 当前应用层入口仍在 `internal/service.go`（与 `internal/app` 保持兼容）。
- 后续可继续将 `internal/service.go` 的细粒度实现逐步迁移至 `internal/app` 子模块。
