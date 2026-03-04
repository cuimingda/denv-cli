# architecture-map

目标：在 12 个要点内把系统边界读出来。

1. `cmd/denv/main.go`（cmd）→ `cmd/root.go`（cmd）→ `cmd/understandability_invariants_test.go`（cmd）
   - 入口边界：`main` 创建根命令并返回统一退出码；测试点看是否存在 `main` 入口。

2. `cmd/root.go`（cmd）→ `cmd/list.go` `cmd/install.go` `cmd/outdated.go` `cmd/update.go`（cmd）→ `cmd/tools_test.go`（cmd）
   - 命令边界：子命令装配与 CLI 参数映射。

3. `cmd/tools.go`（cmd）→ `internal/service.go`（internal）→ `cmd/tools_compat_test.go`（cmd）
   - 依赖边界：命令层只见抽象接口，不直接拼接业务计算。

4. `internal/catalog.go`（internal）→ `internal/service.go`（internal）→ `cmd/understandability_invariants_test.go`（cmd）
   - 模块边界：工具目录与默认清单（支持工具/可安装工具）集中定义。

5. `internal/runtime.go`（internal）→ `internal/install.go`（internal）→ `cmd/understandability_invariants_test.go`（cmd）
   - 外部依赖边界：命令执行与路径查找统一走 `Runtime` 接口。

6. `internal/workflows.go`（internal）→ `internal/version.go`（internal）→ `internal/version_test.go`（internal）
   - 数据流边界：`list`/`outdated` 先拿支持清单，再逐项采集状态，再产出 `ToolListItem`/`OutdatedItem`。

7. `internal/install.go`（internal）→ `internal/operation.go`（internal）→ `cmd/coverage_gaps_test.go`（cmd）
   - 数据流边界：先规划队列（queue）再执行；失败即中断，避免半更新状态。

8. `internal/workflows.go` 的 `outdatedUpdatePlan` → `cmd/outdated.go` → `cmd/update.go`
   - 错误流边界：版本异常时产生 `OutdatedError`，更新流程阻断。

9. `internal/version.go`（internal）→ `cmd/outdated_test.go`（cmd）
   - 数据流边界：版本源分流（命令/ brew / npm）在版本层统一处理。

10. `internal/compare.go`（internal）→ `cmd/outdated_test.go`（cmd）→ `internal/version_test.go`（internal）
    - 数据流边界：统一版本比较策略保证 `outdated` 语义可复现。

11. `internal/path_policy.go`（internal）→ `internal/install.go`（internal）→ `internal/service_test.go`（internal）
    - 外部依赖边界：Homebrew 路径判断策略在内部可替换、可测试。

12. 错误与输出边界：`cmd/verbose.go`（cmd）→ `cmd/list.go`/`cmd/install.go`/`cmd/update.go`（cmd）；命令反馈边界：`[INFO]` 与 `[verbose]` 分离，stderr 不污染 stdout；测试：`cmd/tools_compat_test.go`、`cmd/cli_contract_test.go`
