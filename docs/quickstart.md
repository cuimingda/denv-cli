# quickstart

目标：10 分钟内完成“读、跑、改、复测”。

## 1) 面向开发者准备源码
```bash
git clone https://github.com/cuimingda/denv-cli.git
cd denv-cli
```

## 2) 安装并运行

### 面向使用者（从发布源安装）
```bash
go install github.com/cuimingda/denv-cli/cmd/denv@latest

denv --help
```

### 面向开发者（本地开发安装）
```bash
go install ./cmd/denv
denv --help
```

## 3) 验证关键命令
```bash
denv list
denv list --version --path
denv outdated
denv install --dry-run
```

## 4) 运行测试
```bash
go test ./...
```

## 5) 修改一个最小功能（最小边界改动）
将 CLI 版本号从 `0.0.1` 改为 `0.0.2`（不影响主流程）：
```bash
perl -0pi -e 's/const version = "0.0.1"/const version = "0.0.2"/' cmd/root.go
```

## 6) 再次运行测试（验证修改路径）
```bash
denv --help
go test ./cmd -run 'TestUnderstandabilityInvariant_|TestUnderstandabilityInvariant_'
go test ./internal -run 'TestListToolItemsOrderMatchesCatalogList|TestBuildInstallQueueIsStableAcrossCalls'
```

可执行命令到测试的最短查找：
- `cmd/root.go`（变更点）
- `cmd/understandability_invariants_test.go`（变更的可见合同）
- `internal/service_invariants_test.go`
