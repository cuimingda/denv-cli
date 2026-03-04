# internal/app

职责：
- 协调命令输入到领域流程的执行顺序。
- 组合领域服务与外部输出，不直接发起系统调用。

当前实现锚点：
- 入口与协调：`cmd` 通过 `internal.Service` 调用。
- 错误与过程日志：`cmd/*`。

迁移说明：
- 现阶段为保持行为兼容，具体实现仍集中在 `internal/`。
- 后续重构可将 `internal/service.go` 的聚合能力逐步迁移到 `internal/app`。
