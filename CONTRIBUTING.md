# CONTRIBUTING

## 1. 项目目标与范围

`denv` 是一个本地开发工具状态管理 CLI。它关注命令行可见性、版本信息采集、安装/更新编排，不承担系统级环境管理、镜像管理、编译器链搭建或跨机器状态同步。

边界内：
- `list`、`outdated`、`install`、`update` 的参数、输出与退出码。
- 统一的命令执行编排与错误转码。

边界外：
- 扩展 `homebrew`、`xcode`、`nvm` 等具体实现策略；这些策略仅在 `internal/infra` 抽象中以依赖注入形式体现。

## 2. 架构边界（可修改模块）

- `cmd/`：命令行入口、参数解析、错误到退出码映射、输出组装。
  - 典型文件：`cmd/root.go`、`cmd/list.go`、`cmd/install.go`、`cmd/outdated.go`、`cmd/update.go`
- `internal/`：应用服务和协作层，负责用例编排。
  - 典型文件：`internal/service.go`、`internal/workflows.go`
- `internal/domain/`：纯业务规则（不做 IO/系统调用）。
  - 典型文件：`internal/domain/version.go`、`internal/domain/install.go`、`internal/domain/helpers.go`
- `internal/infra/`：对外部命令、运行时、路径策略等封装。
  - 典型文件：`internal/infra/runtime.go`、`internal/infra/runtime_adapter.go`

## 3. 新功能放在哪

### 新 CLI flag
1. 在对应命令文件添加 flag（如 `cmd/list.go`）。
2. 在 `cmd/*_test.go` 增加参数解析/契约测试。
3. 在 `internal/service.go` 或领域服务中接入参数含义。
4. 补充 1 个快照式回归测试（同目录下）。

### 新子命令
1. 在 `cmd/` 新增命令文件和测试文件。
2. 在 `cmd/root.go` 注册到根命令。
3. 在 `cmd/tools_test.go` 或 `cmd/cli_contract_test.go` 增加入口可见性测试。
4. 若涉及流程变更，补充 `internal/service_test.go`。

### 修改输出行为
1. 修改输出层相关文件（如 `cmd/presenter.go`、`cmd/verbose.go`）。
2. 更新或新增测试：`cmd/*_test.go`、`cmd/*_invariants_test.go`。
3. 保持 JSON 契约稳定，必要时同步更新不变量测试。

## 4. 不允许破坏的结构规则（架构保护）

1. CLI 入口只定义在 `cmd/`，`main` 只能通过 `cmd` 入口运行。
2. `internal/domain` 禁止直接访问 `os/exec`、文件系统、网络、进程环境等 IO 外部依赖。
3. `internal/infra` 统一承载外部依赖和路径策略，不向 `internal/domain` 泄露命令行执行细节。
4. `internal/service` 只做编排，不直接创建网络/文件/进程 IO。
5. 每个核心模块必须有测试：
   - `cmd/`、`internal/domain`、`internal/infra`、`internal` 至少各有一个测试文件验证关键行为。

## 5. 如何运行测试（最小必需）

```bash
gofmt -l $(find . -name '*.go' -type f) >/tmp/gofmt.out && [ ! -s /tmp/gofmt.out ] || (cat /tmp/gofmt.out && exit 1)
go vet ./...
go test ./...
```

在只改输出/参数前可运行：

```bash
go test ./cmd
go test ./internal
```

## 6. 如何提交代码（统一规则）

提交信息必须使用以下格式（提交信息中可多行）：

```
type(scope): summary

motivation:
说明为什么需要这个修改

impact:
说明该修改影响的模块
```

`type` 建议使用：`feat`、`fix`、`refactor`、`test`、`docs`、`chore`。

示例：

```
feat(cli): add list --source flag

motivation:
支持用户查看工具来源以便排查版本状态

impact:
cmd/list.go, internal/domain/version.go
```

提交前执行：
1. `go test ./...`
2. `gofmt` / `go vet` 无新增差异
3. 保持测试覆盖新增路径

## 7. 如何提出架构修改

1. 先在 `docs/adr/` 新增一条 ADR（含背景、可选方案、选择、理由、调整条件）。
2. 在相关测试文件里补上边界测试。
3. 更新 `CHANGELOG.md` 同一轮提交。
4. 在 PR 描述里列出：
   - 影响的模块
   - 是否触碰结构规则
   - 失败模式是否有覆盖
   - 回归验证命令
