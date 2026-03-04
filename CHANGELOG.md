# CHANGELOG

## [Unreleased]

### 新功能
- 初始化交接文档流程：补齐 `CONTRIBUTING.md` 接力开发流程与架构提交流程。

### 修复
- 无行为修复；同步更新文档与规范后保持实现行为不变。

### 重构
- 规范 `docs/project-overview.md` 与提交说明，使架构边界可直接传承。
- 建立 ADR 目录化决策沉淀（含目录结构、命令边界、错误策略、依赖策略）。

### 破坏性变更
- 无（保持现有命令行为和测试契约）。

## [0.0.1] - 2026-03-04

### 新功能
- 提供可读写的文档化项目起点：`README`、`CONTRIBUTING`、`docs/project-overview.md`、`docs/architecture-map.md`。

### 修复
- 无行为修复；保持命令执行语义。

### 重构
- 固定 `cmd / internal` 分层与 `cmd -> internal` 的边界。
- 建立初始 CI 护栏（`gofmt`、`go vet`、`go test`）。

### 破坏性变更
- 无。
