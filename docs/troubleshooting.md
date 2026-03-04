# troubleshooting

## 常见故障与定位路径（按命令级别）

1) `--output` 参数错误
- 命令：`denv list --output invalid`
- 预期：返回错误，退出码 1。
- 位置：`cmd/list.go` → `parseListOutput`。
- 测试：`cmd/failure_contract_test.go::TestFailureScenario_InvalidInput`

2) Homebrew 不存在但需要安装
- 命令：`denv install`（机器未装 brew）
- 预期：在构建队列阶段报 `homebrew is not installed`。
- 位置：`internal/install.go` → `tool.PlanInstallByPolicy`。
- 测试：`cmd/coverage_gaps_test.go::TestBuildInstallOperationsFailsWithoutHomebrew`

3) 工具版本异常阻断更新
- 命令：`denv update`（当某已安装工具 current 版本解析失败）
- 预期：更新计划返回错误，不执行后续更新。
- 位置：`internal/workflows.go` → `outdatedUpdatePlan`。
- 测试：`cmd/understandability_invariants_test.go::TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`

4) 测试锚点失效（行为变化）
- 命令：`go test ./...`（先只读测试名）
- 处理：对比本文件中的测试锚点路径，必要时更新 `docs/invariants.md`。
