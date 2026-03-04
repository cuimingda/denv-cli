# denv 入口地图（可定位版本）

用于 10 分钟定位：主入口 → 配置 → 核心域 → 外部依赖 → 错误总线。

| 文件 | 核心入口 | 锚点测试 |
|---|---|---|
| `cmd/denv/main.go` | `main` | `TestUnderstandabilityInvariant_RootHelpExposesCoreCommands` |
| `cmd/root.go` | `NewRootCmd`、`NewRootCmdWithContext` | `TestNewRootCmdWithContextPanicsWhenDependenciesMissing` |
| `cmd/tools.go` | `NewCLIContext`、`ensureCLIContext`、命令服务接口 | `TestEnsureCLIContextOnlyCreatesContextWhenNil` |
| `cmd/list.go` | `NewListCmdWithService`、`parseListOutput` | `TestNewListCmdDefaultShowsToolsOnly` |
| `cmd/install.go` | `NewInstallCmdWithService`、`--dry-run` | `TestInstallCommandDryRunShowsOperationsOnly` |
| `cmd/outdated.go` | `NewOutdatedCmdWithService` | `TestOutdatedShowsOutdatedTool` |
| `cmd/update.go` | `NewUpdateCmdWithService` | `TestUpdateCommandUpdatesOnlyOutdatedTools` |
| `cmd/presenter.go` | `listPresenter` / `outdatedPresenter`、输出模式 | `TestUnderstandabilityInvariant_ListOutputOrderAndJSONSchemaStable` |
| `cmd/verbose.go` | `doingf`、`verbosef` | `TestRootVerboseRunsAndLogsInstallDryRun` |
| `internal/service.go` | `Service` 聚合能力，统一上下文 | `TestServiceSupportedToolsAndInstallableOrder` |
| `internal/catalog.go` | 工具定义与顺序策略 | `TestServiceSupportedToolsAndInstallableOrder` |
| `internal/install.go` | 安装计划、队列、执行 | `TestInstallCommandInstallsAllTools` |
| `internal/workflows.go` | list/outdated 核心工作流，`OutdatedError` | `TestServiceOutdatedChecksReturnsStructuredItems` |
| `internal/version.go` | 版本来源与版本比较 | `TestCmpVersions`、`TestParseBrewStableVersionUsesRevision` |
| `internal/runtime.go` | 命令执行与查找边界 | `TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion` |
| `internal/path_policy.go` | homebrew 路径归类规则 | `TestResolvedBrewBinaryPathFallsBackToOptBinPrefix` |

## 命令入口（最小反例）

- `go run ./cmd/denv --help`
- `go run ./cmd/denv list --version --path`
- `go run ./cmd/denv outdated --output json`
- `go run ./cmd/denv install --dry-run`

## 风险定位（先查测试，再查实现）

- `brew` 缺失阻断：`TestBuildInstallOperationsFailsWithoutHomebrew`
- 版本解析异常：`TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`
- 输出可预测性：`TestUnderstandabilityInvariant_OutdatedJSONContract`
