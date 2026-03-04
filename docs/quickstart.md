# quickstart

目标：10 分钟内完成“读、跑、改、复测”。

## 1) clone 项目
```bash
git clone https://github.com/cuimingda/denv-cli.git
cd denv-cli
```

## 2) 运行程序
```bash
go run ./cmd/denv --help

go run ./cmd/denv list
go run ./cmd/denv list --version --path
go run ./cmd/denv outdated
go run ./cmd/denv install --dry-run
```

## 3) 运行测试
```bash
go test ./...
```

## 4) 修改一个最小功能（最小边界改动）
将 CLI 版本号从 `0.0.1` 改为 `0.0.2`（不影响主流程）：
```bash
perl -0pi -e 's/const version = "0.0.1"/const version = "0.0.2"/' cmd/root.go
```

## 5) 再次运行测试（验证修改路径）
```bash
go run ./cmd/denv --help
go test ./cmd -run 'TestUnderstandabilityInvariant_|TestUnderstandabilityInvariant_'
go test ./internal -run 'TestListToolItemsOrderMatchesCatalogList|TestBuildInstallQueueIsStableAcrossCalls'
```

可执行命令到测试的最短查找：
- `cmd/root.go`（变更点）
- `cmd/understandability_invariants_test.go`（变更的可见合同）
- `internal/service_invariants_test.go`
