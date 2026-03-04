# denv 可理解性地图与验证清单

目标：让人不阅读实现全量代码，也能定位系统意图、边界和风险。

## 核心意图

- 目标用户：开发者环境初始化脚本或日常工具维护。
- 解决问题：统一检测、安装、检查与更新常用工具，输出可读/脚本友好格式。
- 当前边界：
  - 仅支持 macOS + Homebrew 流程。
  - 仅在检测到不一致或异常时返回错误退出；不做自动修复除更新/安装外的额外变更。

## 文件级地图（关键路径）

### 命令入口
- `cmd/denv/main.go`：`main` 只做 root 命令执行与错误退出。
- `cmd/root.go`：`NewRootCmd` 聚合根命令与子命令，声明 `--verbose`。
- `cmd/list.go` / `cmd/install.go` / `cmd/outdated.go` / `cmd/update.go`：每个子命令都只做参数到服务的映射。
- `cmd/presenter.go`：输出层，控制 `plain/json/table/no-color`。
- `cmd/verbose.go`：日志总线，保证命令日志行为集中在可定位位置。

### 内部领域边界
- `internal/catalog.go`：工具模型与 install policy（`ToolDefinition`、`Tool`、`toolCatalog`）。
- `internal/workflows.go`：`list` 与 `outdated` 的流程模型（`ToolListItem`、`ToolCheckResult`、`OutdatedItem`）。
- `internal/install.go`：安装编排，含 registry、plan/queue、执行接口。
- `internal/version.go`：版本解析与版本来源（brew/npm）。
- `internal/runtime.go` / `internal/path_policy.go`：外部 IO 与路径策略隔离。
- `internal/service.go`：Facade 门面，统一服务边界，隔离命令层与领域行为。

### 错误与退出语义
- 命令层：`RunE` 返回错误给 `cobra`，`main.go` 统一 `os.Exit(1)`。
- 关键结构化错误：`internal/workflows.go` 的 `OutdatedError`（`ToolName` + `State`）。
- 兼容/边界错误：`unsupported tool`、`homebrew is not installed`、`tool ... is not installed` 之类字符串错误。

## 风险边界（显式追踪）

1. `Homebrew` 依赖失效：`install/update` 命令对 brew 的可用性与输出格式敏感。
2. 版本抽取假设：`RegexVersionParser` 与 brew 数据源可能对少见工具/输出格式失效。
3. 输出策略分叉：`--verbose` 与 `--output` 同时存在时行为叠加需要按文档理解（目前日志始终写 `stderr`）。

## 不变量锚点与测试映射

- 顺序稳定（列表/安装顺序）：
  - `internal/service_test.go:TestServiceSupportedToolsAndInstallableOrder`
  - `internal/service_invariants_test.go:TestBuildInstallQueueIsStableAcrossCalls`
- 幂等（同一输入重复构建一致）：
  - `internal/service_invariants_test.go:TestBuildInstallQueueIsStableAcrossCalls`
- 输出可预期（数量/顺序/字段完整）：
  - `cmd/understandability_invariants_test.go:TestListCommandPlainOutputLineCountAndOrder`

## 验证口径（建议跑一遍）

1. 冷启动复述：只看 `README.md` + 测试名，不看实现源码，复述「目标、非目标、失败模式」。
2. 入口图定位：10 分钟内在上述文件中找到入口、外部依赖、错误总线。
3. 抽样反例：
   - `go run ./cmd/denv install --dry-run`
   - `HOME=false go run ./cmd/denv install`（或移除 brew）验证失败路径
   - 在无 tool PATH 的环境中观察 `go run ./cmd/denv outdated` 输出是否仍稳定

## 入口地图执行索引

更细颗粒度的“文件→核心函数→反例命令”清单见：
[docs/understandability_entry_index.md](/Users/cuimingda/Projects/denv-cli/docs/understandability_entry_index.md)
