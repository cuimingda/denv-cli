# invariants

关键不变量与测试锚点（共 9 条，5–12 条要求内）。

1. 根命令可见命令与 `--verbose`
- 描述：`denv` 的 `list/install/outdated/update` 与 `--verbose` 必须在帮助中持续可见。
- 测试文件：`cmd/understandability_invariants_test.go`
- 测试函数：`TestUnderstandabilityInvariant_RootHelpExposesCoreCommands`

2. `list` 顺序是稳定的目录顺序
- 描述：`list` 输出顺序与 `internal/catalog.go` 的 `SupportedTools` 顺序一致。
- 测试文件：`cmd/understandability_invariants_test.go`
- 测试函数：`TestUnderstandabilityInvariant_ListOutputOrderAndJSONSchemaStable`

3. 列表 JSON 契约稳定
- 描述：`list --output json` 的字段名、空值行为与顺序固定。
- 测试文件：`cmd/understandability_invariants_test.go`
- 测试函数：`TestUnderstandabilityInvariant_ListOutputOrderAndJSONSchemaStable`

4. 过期 JSON 契约稳定
- 描述：`outdated --output json` 必须包含 `name/state/current/latest`，字段语义可复算。
- 测试文件：`cmd/understandability_invariants_test.go`
- 测试函数：`TestUnderstandabilityInvariant_OutdatedJSONContract`

5. CLI 参数错误是失败模式
- 描述：非法 `--output`/非法参数应阻断并返回错误。
- 测试文件：`cmd/cli_contract_test.go`、`cmd/failure_contract_test.go`
- 测试函数：`TestContract_InvalidListOutputArgReturnsNonZeroExitAndUsage`、`TestFailureScenario_InvalidInput`

6. CLI 退出码可预测
- 描述：成功命令返回 `0`，错误命令返回非 `0`，主入口统一收敛到 `main` 退出码。
- 测试文件：`cmd/cli_contract_test.go`、`cmd/understandability_invariants_test.go`
- 测试函数：`TestContract_RootHelpHasEntrypointsAndExitZero`、`TestContract_CLIExitCodeInProcessForUnknownCommand`、`TestContract_InvalidListOutputArgReturnsNonZeroExitAndUsage`

7. 安装队列与执行顺序可复现
- 描述：多次构建 `install` 队列顺序不变，关键用于幂等和脚本可复现。
- 测试文件：`internal/service_invariants_test.go`
- 测试函数：`TestBuildInstallQueueIsStableAcrossCalls`

8. 关键失败快停止
- 描述：版本解析异常、缺少依赖时应尽早失败，不继续执行后续动作。
- 测试文件：`cmd/understandability_invariants_test.go`、`cmd/failure_contract_test.go`
- 测试函数：`TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`、`TestFailureScenario_ConfigMissingHomebrew`、`TestFailureScenario_FileNotFoundWhileResolvingVersion`

9. 外部依赖与运行时边界
- 描述：命令执行、命令路径、Homebrew 判断只能通过 `Runtime` 与 path policy 统一口子接入。
- 测试文件：`internal/runtime_test.go`、`cmd/coverage_gaps_test.go`
- 测试函数：`TestServiceIsCommandAvailable`、`TestBuildInstallOperationsSkipsInstalledTools`、`TestResolvedBrewBinaryPathFallsBackToOptBinPrefix`
