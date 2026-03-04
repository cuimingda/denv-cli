# CONTRIBUTING

## 1. 目标与范围（先读）

- 项目目标：维护 `denv` CLI 的可交付能力：命令可见性、版本检测、安装更新编排与稳定退出码。
- 范围内：`list`、`outdated`、`install`、`update` 的参数、输出、行为一致性。
- 范围外：系统级环境安装、镜像管理、跨机器同步机制。

## 2. 架构边界（结构规则）

- `cmd/`：仅包含 CLI 入口、参数解析、错误映射、输出组织。
- `internal/app`、`internal/domain`：纯用例编排和业务规则。
- `internal/infra`：外部命令、文件系统、环境变量、路径策略等 IO 封装。
- 接口流向固定：`cmd -> internal/app -> internal/domain -> internal/infra`。

## 3. 新功能放置流程

- 新 CLI flag
  1. 在对应命令文件中定义 flag（如 `cmd/list.go`）。
  2. 在 `cmd/*_test.go` 补充解析和契约测试。
  3. 通过 `internal/app`/`internal/domain` 衔接行为（如有逻辑变更）。
  4. 确保 JSON 或文本输出有回归测试。
- 新子命令
  1. 在 `cmd/` 新增命令与测试文件。
  2. 在 `cmd/root.go` 注册。
  3. 在 `cmd/tools_test.go` 增加根命令可见性测试。
  4. 需要业务流程的改动补 `internal/service_test.go`。
- 修改输出行为
  1. 在 `cmd/` 输出层调整显示逻辑。
  2. 同步更新相关快照/契约测试。
  3. 保持 JSON 键名、exit code 及关键文本不变量稳定或同步更新相关测试。

## 4. 不允许破坏的规则（护栏）

- CLI 入口必须仅在 `cmd/`；`main` 仅通过 `cmd` 启动。
- `internal/domain` 不得直接使用 `os/exec`、文件系统、网络、环境变量、进程控制。
- `internal/infra` 承担全部外部依赖，不得将 shell 执行细节泄漏给 `cmd/domain`。
- `internal/app` 为编排层，不直接创建外部 IO。
- 每个核心模块必须有关键测试：`cmd`、`internal/app`、`internal/domain`、`internal/infra`。

## 5. 运行方式（最小与完整）

```bash
files=$(find . -name '*.go' -type f -print0 | xargs -0 gofmt -l)
if [ -n "$files" ]; then
  echo "$files"
  exit 1
fi
go vet ./...
go test ./...
```

只改 `cmd` 或单个命令时，先跑：

```bash
go test ./cmd
```

## 6. 提交规则（统一）

提交信息使用：

```text
type(scope): summary

motivation:
说明为什么需要这个修改

impact:
说明该修改影响的模块
```

- type 建议：`feat`、`fix`、`refactor`、`test`、`docs`、`chore`。
- 例：

```text
feat(cli): add list --source flag

motivation:
支持用户快速查看工具来源，降低安装排障时间

impact:
cmd/list.go, internal/domain/version.go
```

## 7. 架构修改提交流程

1. 在 `docs/adr/` 新增 ADR（标题、背景、可选方案、选择、理由、未来调整条件）。
2. 补齐新增行为的失败与边界测试。
3. 按 [6](#6-提交规则统一) 提交并在同一轮更新 `CHANGELOG.md`。
4. PR 描述必须列出：变更模块、护栏影响、测试命令、失败场景覆盖。

## 8. 接力开发验收（新执行者可按此开始）

1. 只读取 `README.md`、`CONTRIBUTING.md`、`docs/*`、`*test.go`。
2. 按现有约束完成一项可见行为改动。
3. 执行 `go test ./...`，确保通过，且命令边界未跨层。
