# internal/domain

职责：
- 存放不依赖 IO/网络的规则与不变量。
- 包含工具元数据、版本语义与状态模型。

当前规则锚点：
- 工具元数据与顺序：`internal/catalog.go`
- 版本比较策略：`internal/compare.go`
- 流程模型：`internal/workflows.go`

规则约束：
- 禁止直接调用 `exec.Command`。
- 通过 `infra` 暴露的能力注入输入。
