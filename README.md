# denv CLI

## 3句话理解
- denv 是一个 CLI，统一描述、检测、列表、安装和更新“受支持开发工具”的本机状态与操作。
- denv 不负责系统级环境管理、外部应用栈（如编译器链或容器平台）安装，也不维护复杂依赖解析。
- denv 的边界是：只读/写入对 `internal/catalog.go` 与 `internal/` 业务服务的命令编排层，错误统一回退到 CLI 退出码。

## 先看这里（导航）
- [quickstart](./docs/quickstart.md)
- [CONTRIBUTING](./CONTRIBUTING.md)
- [project-overview](./docs/project-overview.md)
- [ADR 列表](./docs/adr)
- [architecture-map](./docs/architecture-map.md)
- [invariants](./docs/invariants.md)
- [troubleshooting](./docs/troubleshooting.md)

## 三个示例（命令+预期）

### 1) 最小示例
```bash
go run ./cmd/denv list
```
输出示例（第一批）：`php`、`python3`、`node`…

### 2) 常见用法
```bash
go run ./cmd/denv list --version --path
go run ./cmd/denv outdated --output json
go run ./cmd/denv install --dry-run
```
行为锚点：
- `list --version --path` 走 `cmd/list.go`，产出 `internal/domain` 列表模型；
- `outdated --output json` 保证 `name/state/current/latest`；
- `install --dry-run` 仅打印 `Would run:`。

### 3) 失败示例（输出与退出码）
```bash
go run ./cmd/denv list --output invalid
```
输出关键行（示例）：
```text
Error: invalid output: invalid
Usage:
  denv list [flags]
```
退出码：`1`

## 读者下一步
先按 `quickstart` 完成首次运行，再对照 `architecture-map` 找到每条命令的文件入口。
