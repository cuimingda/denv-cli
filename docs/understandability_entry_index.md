# denv 入口地图单元索引（脚本化核验版）

目标：不逐行读源码，也能 10 分钟内定位“入口、边界、核心行为与风险点”。

每条都给出：
- 文件
- 最核心函数
- 验证意图
- 一条可执行反例命令（用于快速回归）

## 命令启动与组装层

| 文件 | 核心函数 | 核心职责 | 反例命令 |
|---|---|---|---|
| `cmd/denv/main.go` | `main` | 程序执行入口、统一返回码 | `go run ./cmd/denv --help` |
| `cmd/root.go` | `NewRootCmdWithContext` | 注册 `list/install/outdated/update`、`--verbose` | `go run ./cmd/denv help | rg "list|install|outdated|update|--verbose"` |
| `cmd/tools.go` | `NewCLIContext` | 命令层/服务层的上下文组装、端口定义 | `go test ./cmd -run 'TestNewListCmdWithServicePanicsWhenNilService|TestNewRootCmdWithContextPanicsWhenDependenciesMissing' -count=1` |

## 命令行为层

| 文件 | 核心函数 | 核心职责 | 反例命令 |
|---|---|---|---|
| `cmd/list.go` | `NewListCmdWithService` | 参数解析与列表服务调用 | `go run ./cmd/denv list --version --path` |
| `cmd/install.go` | `NewInstallCmdWithService` | `--force`/`--dry-run` 与安装队列执行编排 | `go run ./cmd/denv install --dry-run` |
| `cmd/outdated.go` | `NewOutdatedCmdWithService` | 过期检测入口、状态渲染分发 | `go run ./cmd/denv outdated --output table` |
| `cmd/update.go` | `NewUpdateCmdWithService` | 过期计划生成与更新执行入口 | `go run ./cmd/denv update` |
| `cmd/presenter.go` | `listPresenter.Render` / `outdatedPresenter.Render` | plain/json/table 的输出边界 | `go run ./cmd/denv list --output json` |
| `cmd/verbose.go` | `doingf` / `verbosef` | 日志输出总线（stderr）边界 | `go run ./cmd/denv --verbose list --version | head -n 3` |

## 核心服务与能力边界

| 文件 | 核心函数 | 核心职责 | 反例命令 |
|---|---|---|---|
| `internal/service.go` | `NewService` / `Service.BuildInstallQueue` / `Service.UpdateToolWithOutput` | Facade 聚合，统一外部能力调用 | `go test ./internal -run TestServiceSupportedToolsAndInstallableOrder -count=1` |
| `internal/catalog.go` | `NewToolCatalog` / `Tool.IsInstalled` / `Tool.PlanInstall` | 配置驱动的领域规则（顺序、可安装、命令映射） | `go test ./internal -run TestServiceOutdatedChecksReturnsStructuredItems -count=1` |
| `internal/install.go` | `BuildInstallQueue` / `BuildInstallQueueForTool` / `UpdateToolWithOutputWithCatalog` | 安装计划编排、registry、执行策略 | `go run ./cmd/denv install --dry-run` |
| `internal/workflows.go` | `listToolItems` / `outdatedChecks` / `outdatedUpdatePlan` | list/outdated 的核心工作流 | `go test ./cmd -run TestOutdatedHandlesMissingTool -count=1` |
| `internal/version.go` | `ToolVersionWithPathWithCatalog` / `ToolLatestVersionWithCatalog` / `CompareVersions` | 版本源（brew/npm）与版本比较规则 | `go test ./cmd -run TestOutdatedShowsOutdatedTool -count=1` |
| `internal/runtime.go` | `NormalizeRuntime` | 外部系统交互默认行为边界 | `go test ./internal -run TestServiceSupportedToolsAndInstallableOrder -count=1` |
| `internal/path_policy.go` | `DefaultPathPolicy` / `IsManagedByHomebrew` | Homebrew 路径判断边界 | `go test ./cmd -run TestResolvedBrewBinaryPathFallsBackToOptBinPrefix -count=1` |

## 脚本化 10 分钟核验模板

```bash
# 1) 快速定位入口与输出
go run ./cmd/denv --help

# 2) 关键路径可达性（不会修改本机状态）
go run ./cmd/denv list --version --path --output table
go run ./cmd/denv outdated --output json
go run ./cmd/denv install --dry-run

# 3) 主要边界（失败模式）
go run ./cmd/denv update || true   # 在无 brew 或无更新时应失败或输出 no updates available

# 4) 直接命中不变量锚点测试
go test ./internal -run 'TestServiceSupportedToolsAndInstallableOrder|TestBuildInstallQueueIsStableAcrossCalls' -count=1
go test ./cmd -run 'TestListCommandPlainOutputLineCountAndOrder' -count=1
```

## 命令级风险点定位清单（按文件查找）

- `internal/install.go`：`buildInstallQueueByTool`、`buildInstallPlanForToolWithRunPolicy`  
  - 风险：`--force` 误用导致不必要操作。  
- `internal/version.go`：`ExtractVersion`、`toolVersionFromCommandSource`  
  - 风险：命令输出格式变更导致版本提取失败。  
- `cmd/update.go` / `internal/workflows.go`：`OutdatedStateInvalidCurrent`、`OutdatedStateInvalidLatest`  
  - 风险：部分工具版本异常会阻断批量更新。  
- `internal/path_policy.go`：`IsManagedByHomebrew`  
  - 风险：非标准 Homebrew 安装路径会导致 `managed_by_brew` 判定偏差。  
- `cmd/root.go` + `cmd/verbose.go`：`--verbose` 与 `--output` 联动  
  - 风险：日志/结果通道分离（stderr vs stdout）造成脚本解析误判。  
