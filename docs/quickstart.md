# quickstart

目标：按顺序完成“读、跑、改、复测”，每一步都有下一步命令。

## 1) clone 项目
```bash
git clone https://github.com/cuimingda/denv-cli.git
cd denv-cli
```

## 2) 运行程序（源码直接执行）
```bash
go run ./cmd/denv --help
go run ./cmd/denv list
go run ./cmd/denv outdated
```

## 3) 运行测试
```bash
go test ./cmd
go test ./internal
```

## 4) 修改一个最小功能
修改根命令的短描述（不改流程）：
```bash
perl -0pi -e 's/denv command line interface/denv dev-tool workflow cli/' cmd/root.go
```

## 5) 再次运行测试（回归改动边界）
```bash
go test ./cmd -run 'TestContract_RootHelpHasEntrypointsAndExitZero|TestUnderstandabilityInvariant_RootHelpExposesCoreCommands'
go test ./internal -run 'TestBuildInstallQueueIsStableAcrossCalls|TestListToolItemsOrderMatchesCatalogList'
go test ./cmd -run 'TestFailureScenario_InvalidInput'
```

## 关键定位（改完后立刻看）
- 文件：`cmd/root.go`（入口描述）
- 目录：`cmd/`（命令层）
- 测试：`cmd/understandability_invariants_test.go`、`cmd/cli_contract_test.go`
