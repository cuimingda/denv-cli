# invariants

所有关键不变量均有测试锚点（5–12 条）。

1. CLI 帮助入口可见性
- 描述：`denv` 主命令必须暴露 `list/install/outdated/update` 与 `--verbose`。
- 测试：`cmd/understandability_invariants_test.go`
- 函数：`TestUnderstandabilityInvariant_RootHelpExposesCoreCommands`

2. 工具列表顺序稳定
- 描述：`list` 与 `SupportedTools` 必须保持一致顺序。
- 测试：`internal/service_invariants_test.go`
- 函数：`TestListToolItemsOrderMatchesCatalogList`

3. 全量安装队列顺序稳定
- 描述：多次构建安装队列返回一致结果，便于脚本复现。
- 测试：`internal/service_invariants_test.go`
- 函数：`TestBuildInstallQueueIsStableAcrossCalls`

4. 输出结构稳定
- 描述：`list --output json` 与 `outdated --output json` 字段与顺序为稳定契约。
- 测试：`cmd/understandability_invariants_test.go`
- 函数：`TestUnderstandabilityInvariant_ListOutputOrderAndJSONSchemaStable`

5. 过期检测状态语义稳定
- 描述：`OutdatedState` 必须映射到可读 JSON 字段。
- 测试：`cmd/understandability_invariants_test.go`
- 函数：`TestUnderstandabilityInvariant_OutdatedJSONContract`

6. 版本比较策略稳定
- 描述：语义版本与日期版本比较必须按一致策略返回正确先后关系。
- 测试：`internal/version_test.go`、`cmd/outdated_test.go`
- 函数：`TestResolveVersionStrategySelection`、`TestCmpVersions`

7. 失败快速阻断
- 描述：安装/更新流程在关键失败（如版本异常）必须快速返回，不执行后续不安全动作。
- 测试：`cmd/understandability_invariants_test.go`
- 函数：`TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion`

8. 安装行为受已有安装状态保护
- 描述：无需强制执行时，对已存在二进制的工具不应重复执行安装动作。
- 测试：`cmd/coverage_gaps_test.go`
- 函数：`TestBuildInstallOperationsSkipsInstalledTools`

9. 依赖判定边界不侵入业务
- 描述：外部依赖访问必须通过 `Runtime`（非直接 `exec.Command`）；测试可注入桩。
- 测试：`internal/runtime_test.go`、`internal/service_test.go`
- 函数：`TestServiceIsCommandAvailable`

10. Homebrew 管理路径识别
- 描述：由 path policy 推断的 brew 路径与推导逻辑要稳定。
- 测试：`cmd/coverage_gaps_test.go`
- 函数：`TestResolvedBrewBinaryPathFallsBackToOptBinPrefix`
