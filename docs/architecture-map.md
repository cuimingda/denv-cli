# architecture-map

一页读懂边界与数据流（12 条内），每条都带：目录 / 关键文件 / 关键测试。

1. `cmd`：CLI 入口启动与退出码总线
   - 目录：`cmd/`
   - 文件：`cmd/denv/main.go`、`cmd/root.go`
   - 测试：`cmd/cli_contract_test.go::TestContract_RootHelpHasEntrypointsAndExitZero`

2. `cmd`：子命令组装与参数公开面
   - 目录：`cmd/`
   - 文件：`cmd/root.go`、`cmd/list.go`、`cmd/install.go`、`cmd/outdated.go`、`cmd/update.go`
   - 测试：`cmd/tools_test.go::TestRootHasListCommand`、`cmd/tools_test.go::TestRootHasVerboseFlag`

3. `cmd`：命令参数解析（等同配置口）
   - 目录：`cmd/`
   - 文件：`cmd/list.go::parseListOutput`、`cmd/outdated.go`、`cmd/install.go`
   - 测试：`cmd/cli_contract_test.go::TestContract_InvalidListOutputArgReturnsNonZeroExitAndUsage`、`cmd/failure_contract_test.go::TestFailureScenario_InvalidInput`

4. `cmd`→`internal`：接口边界与依赖注入
   - 目录：`cmd/`、`internal/`
   - 文件：`cmd/tools.go`、`internal/ports.go`、`internal/service.go`
   - 测试：`cmd/tools_test.go::TestNewListCmdWithServicePanicsWhenNilService`、`cmd/tools_test.go::TestNewInstallCmdWithServicePanicsWhenPlannerMissing`

5. `internal`：目录清单与稳定顺序来源
   - 目录：`internal/`
   - 文件：`internal/catalog.go`、`internal/service.go`
   - 测试：`internal/service_invariants_test.go::TestListToolItemsOrderMatchesCatalogList`

6. `internal`：列表/版本查询数据流
   - 目录：`internal/`
   - 文件：`internal/workflows.go`、`internal/version.go`
   - 测试：`cmd/tools_test.go::TestNewListCmdWithVersionOnly`、`cmd/outdated_test.go::TestOutdatedShowsOutdatedTool`

7. `internal`：安装计划与执行分离
   - 目录：`internal/`
   - 文件：`internal/install.go`、`internal/operation.go`
   - 测试：`internal/service_invariants_test.go::TestBuildInstallQueueIsStableAcrossCalls`、`cmd/coverage_gaps_test.go::TestBuildInstallOperationsSkipsInstalledTools`

8. `internal`：更新计划与更新动作
   - 目录：`internal/`
   - 文件：`internal/workflows.go`、`cmd/update.go`
   - 测试：`cmd/understandability_invariants_test.go::TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`、`cmd/tools_test.go::TestUpdateCommandUpdatesOnlyOutdatedTools`

9. `internal/domain` + `internal/infra`：边界隔离点
   - 目录：`internal/domain`、`internal/infra`
   - 文件：`internal/domain/domain.go`、`internal/infra/infra.go`、`internal/runtime.go`
   - 测试：`internal/runtime_test.go::TestNormalizeRuntimeProvidesFallbacks`、`internal/version_test.go::TestResolveVersionStrategySelection`

10. `internal`：错误映射到命令退出语义
   - 目录：`cmd/`、`internal/`
   - 文件：`cmd/denv/main.go`、`internal/workflows.go`、`internal/version.go`
   - 测试：`cmd/cli_contract_test.go::TestContract_InvalidListOutputArgReturnsNonZeroExitAndUsage`、`cmd/failure_contract_test.go::TestFailureScenario_ConfigMissingHomebrew`

11. `cmd`：输出与可见性策略
   - 目录：`cmd/`
   - 文件：`cmd/verbose.go`、`cmd/presenter.go`、`cmd/list.go`
   - 测试：`cmd/tools_test.go::TestRootVerboseRunsAndLogsInstallDryRun`、`cmd/tools_test.go::TestInstallCommandDryRunShowsOperationsOnly`

12. `internal/path_policy`：外部依赖路径边界
   - 目录：`internal/`
   - 文件：`internal/path_policy.go`、`internal/workflows.go`
   - 测试：`cmd/coverage_gaps_test.go::TestResolvedBrewBinaryPathFallsBackToOptBinPrefix`、`cmd/coverage_gaps_test.go::TestInstallNodeWithOutputSkipsWhenNpmExists`
