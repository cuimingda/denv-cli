# troubleshooting

读完文档后先执行故障复现命令，再按路径-测试定位。

1) `--output` 非法值
- 命令：`go run ./cmd/denv list --output invalid`
- 预期：标准错误包含 `invalid output`，退出码 `1`。
- 定位：`cmd/list.go` 的 `parseListOutput`，`cmd/cli_contract_test.go`。

2) 未知命令
- 命令：`go run ./cmd/denv does-not-exist`
- 预期：`unknown command`，退出码 `1`。
- 定位：`cmd/root.go`（cobra 子命令挂载），`cmd/cli_contract_test.go::TestContract_CLIExitCodeInProcessForUnknownCommand`。

3) Homebrew 缺失时安装失败
- 命令：`go run ./cmd/denv install --dry-run`
- 预期：在构建安装计划处报 `homebrew is not installed`（仅当本机未安装）。
- 定位：`internal/install.go`、`internal/runtime.go`、`cmd/failure_contract_test.go::TestFailureScenario_ConfigMissingHomebrew`。

4) 版本解析异常导致更新停止
- 命令：`go run ./cmd/denv update`
- 预期：构建更新计划失败并终止后续更新。
- 定位：`internal/workflows.go::outdatedUpdatePlan`、`internal/version.go`、`cmd/understandability_invariants_test.go::TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`。

5) 失败路径回查命令
- 命令：`go test ./cmd ./internal -run 'TestFailureScenario_|TestUnderstandabilityInvariant_'`
- 预期：失败测试名快速指向当前失效锚点。
- 定位：与 `docs/invariants.md` 的不变量条目逐条对应。
