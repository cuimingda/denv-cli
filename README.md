# denv CLI

## 这是什么（3 句话）
- denv 是一个围绕“本机开发工具集合”组织的轻量 CLI：检测、列表、安装与版本更新由同一组命令统一呈现。
- denv 不做系统级补丁、编译环境托管、以及应用层依赖安装；它只处理固定清单中的工具命令。
- denv 的边界是“本机可见工具的探测、版本比较、以及执行受控命令流水线”，不负责安装脚本编排外的工作流。

- [快速入门](./docs/quickstart.md)
- [架构地图](./docs/architecture-map.md)
- [不变量集合](./docs/invariants.md)
- [故障排查](./docs/troubleshooting.md)

## 你可以直接看到的能力
```bash
denv list

denv install --dry-run

denv outdated

denv update
```

## 三个示例

### 1) 最小示例（仅列出工具）
```bash
go run ./cmd/denv list
```
输出（按稳定顺序）：`php`、`python3`、`node`...

### 2) 常见用法（json + update）
```bash
go run ./cmd/denv list --output json --version
go run ./cmd/denv outdated --output json
go run ./cmd/denv install --dry-run
```
输出要求：
- `list --output json` 必须和 `--version` 一致返回每条字段。
- `outdated --output json` 的每条必须包含 `name/state/current/latest`。
- `install --dry-run` 只能输出 `Would run: ...`，不可执行命令。

### 3) 失败示例（含错误与退出码）
```bash
go run ./cmd/denv list --output invalid
```
预期：
```text
Error: invalid output: invalid
Usage:
  denv list [flags]
...
```
退出码：`1`

## 约定
- 运行 CLI 命令时默认入口是 `cmd/denv/main.go`。
- 命令组装入口是 `cmd/root.go`。
- 核心实现分层入口是 `internal/`。
- 测试锚点优先看命名：
  - `cmd/`：`*_test.go` 表示 CLI/流程合同。
  - `internal/`：`*_test.go` 表示核心规则和版本规则。

## 你在本仓库的第一组路径（按理解路径）
1. `README -> quickstart`
2. `README -> architecture-map`
3. `README -> invariants`
4. `README -> cmd/understandability_invariants_test.go`
