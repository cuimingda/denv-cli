# denv 可理解性地图与验证清单

目标：保证你不读实现全量，也能在有限时间内重建系统边界。通过文档 + 测试名 + `--help`，找到“意图、边界、风险、可预期行为”。

## 1) 冷启动复述测试

只看以下三类输入进行一次复述：

- `README.md`
- `go run ./cmd/denv --help`
- 测试目录与测试名（不用源码）

应能稳定回答：

1. 这套 CLI 解决什么问题？
2. 明确不解决什么问题？
3. 最危险的 3 个失败模式？

如果你只能凭猜测回答“没看版本/边界”，说明可理解性不足。

建议复述失败模式：

- 没有 `brew` 时，`install/update` 必须可预见失败，不应静默。
- `brew`/工具版本输出格式异常时，`outdated`/`update` 会产生可定位错误（包含工具名和状态）。
- `--output`、`--verbose` 并行时输出通道与语义应清楚（stdout 业务输出、stderr 日志）。

## 2) 入口地图测试（10 分钟）

目标单位顺序（每一项必须能指向文件/模块/命令）：

1. 主入口
2. 配置与命令目录（工具列表/安装规则）
3. 核心域行为（list/outdated/install/update 的边界）
4. 外部依赖边界（命令执行、路径策略）
5. 错误总线（错误类型、错误传播、退出码）

验收方式：

- `go run ./cmd/denv --help` 能看到主命令集合。
- `go run ./cmd/denv list --version --path` 能确认 list 流程。
- `go run ./cmd/denv outdated --output json` 能确认检查流程。
- `go run ./cmd/denv install --dry-run` 能确认安装计划流程。

对应文件索引见：

- [docs/understandability_entry_index.md](/Users/cuimingda/Projects/denv-cli/docs/understandability_entry_index.md)

## 3) 不变量显性度测试（必须有测试锚点）

每个关键不变量都要有可读测试名：

- 顺序稳定：列表顺序、安装顺序不变。
  - `internal/service_invariants_test.go:TestListToolItemsOrderMatchesCatalogList`
  - `internal/service_invariants_test.go:TestBuildInstallQueueIsStableAcrossCalls`
- 输出可预测：JSON/plain 模式语义一致、可脚本化。
  - `cmd/understandability_invariants_test.go:TestUnderstandabilityInvariant_ListOutputOrderAndJSONSchemaStable`
  - `cmd/understandability_invariants_test.go:TestUnderstandabilityInvariant_OutdatedJSONContract`
- 风险可阻断：检测异常不能盲目更新。
  - `cmd/understandability_invariants_test.go:TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`

## 可执行最小核验

```bash
go run ./cmd/denv --help
go run ./cmd/denv list --output table
go run ./cmd/denv outdated --output json
go run ./cmd/denv install --dry-run

go test ./cmd -run 'TestUnderstandabilityInvariant_' -count=1
go test ./internal -run 'TestBuildInstallQueueIsStableAcrossCalls|TestListToolItemsOrderMatchesCatalogList' -count=1
```
