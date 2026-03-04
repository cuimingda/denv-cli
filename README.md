# denv CLI

`denv`（Developer Environment）是一个面向 macOS 的开发工具管理命令行工具，当前版本 `0.0.1`。  

该仓库核心能力基于 `cobra` 的子命令实现，主命令为：

- `denv list`
- `denv install`
- `denv outdated`
- `denv update`

> 注意：项目当前导出的行为并未提供 `switch/use` 这类版本切换命令。

---

## 运行前提

1. macOS 环境
2. 建议安装 Homebrew：`brew`
   - `install` / `update` 会直接调用 Homebrew 命令
3. 具备可执行文件 `denv`
   - 文档中的命令为最终打包后的 CLI 用法

---

## 全局行为

### 基础调用

```bash
denv                # 显示帮助
denv --help         # 显示命令帮助
denv --version      # 查看版本
```

### 全局日志

```bash
denv --verbose <subcommand>
```

`--verbose` 会在标准错误输出上打印 `[INFO]` 风格的执行进度。

---

## 支持工具

`denv list` 会扫描以下工具（顺序固定）：

`php, python3, node, go, npm, curl, gh, git, ffmpeg, tree`

其中 `install` 子命令默认可处理（可安装）：

`php, python3, node, go, curl, gh, git, ffmpeg, tree`

`npm` 仅在 `list/outdated/update` 中参与检测，不会被安装命令直接安装。

---

## 命令说明

## `denv list`

列出支持列表，可附加版本、路径和多种输出格式。

```bash
denv list [--version] [--path] [--output plain|json|table|no-color]
```

- `--version`：显示已发现工具的当前版本
- `--path`：显示命令路径
- `--output`
  - `plain`（默认）：纯文本
  - `json`：JSON 数组
  - `table`：TAB 分隔表格
  - `no-color`：禁用 ANSI 颜色（即使在终端输出）

示例：

```bash
denv list
denv list --version
denv list --path
denv list --version --path --output table
denv list --version --output json
```

JSON 字段说明：

- `name`：工具名
- `display_name`：展示名
- `installed`：是否可发现
- `version`：版本（仅 `--version`）
- `path`：可执行路径
- `managed_by_brew`：是否判定为 Homebrew 管理

---

## `denv install`

按固定顺序安装所有可安装工具。

```bash
denv install [--force] [--dry-run]
```

- `--dry-run`：只展示待执行计划，不真的执行
- `--force`：无视“已存在”判断，强制生成安装动作

默认行为：

- 未找到 Homebrew 时直接报错
- 已安装且非 `--force` 时会跳过
- `curl` 安装时会在 `brew install curl` 后执行 `brew link curl --force`
- 所有安装动作都通过标准输出透传 Homebrew 输出

示例：

```bash
denv install --dry-run
denv install --force
```

`--dry-run` 典型输出：

```text
Would run: brew install php
Would run: brew install python3
Would run: brew install node
Would run: brew install go
Would run: brew install curl
Would run: brew link curl --force
Would run: brew install gh
Would run: brew install git
Would run: brew install ffmpeg
Would run: brew install tree
```

---

## `denv outdated`

检查支持工具的“当前版本/最新版本/状态”。

```bash
denv outdated [--output plain|json|table|no-color]
```

状态含义：

- `up_to_date`：已是最新
- `outdated`：需要更新（会显示 `current < latest`）
- `not_installed`：未安装，显示 `<not installed> latest`
- `invalid_current`：当前版本解析失败
- `invalid_latest`：最新版本查询失败

示例：

```bash
denv outdated
denv outdated --output table
denv outdated --output json
```

---

## `denv update`

扫描所有支持工具，仅更新处于 `outdated` 状态的工具：

- `brew` 管理的工具：`brew upgrade <formula>`
- `npm`：`npm install -g npm@latest`

```bash
denv update
```

- 若有可更新项：依次执行更新动作
- 若没有可更新项：输出 `no updates available`
- 若某个已安装工具处于版本状态异常（`invalid_current / invalid_latest`）：
  - 会在该问题修复前中止并返回错误

---

## 版本判断逻辑（简要）

`outdated` 使用以下来源计算版本：

- 对大多数工具：`brew info` 获取当前版本/最新稳定版本
- 对 `npm`：`npm view npm version` 查询最新
- 版本比较规则：优先按语义版本比较，版本看起来像日期时按日期比较

---

## 常见问题

- `Homebrew` 未安装
  - `list` 仍可用于基础扫描
  - `install / update` 依赖 `brew`，会返回错误
- `--verbose` 不会改变命令结果，只显示过程日志
- 彩色输出只会在 TTY 中出现；`--output no-color` 强制关闭

---

## 快速上手清单

```bash
denv list --version --path
denv install --dry-run
denv install
denv outdated
denv update
```

## 可理解性导览（用于 10 分钟入口图）

### 一眼定位核心职责

- 本项目解决什么：在 macOS 上管理一组开发者工具（list/install/outdated/update）的安装与状态判断。
- 不解决什么：
  - 不做版本回滚（只做当前/最新判断及更新，不做降级）。
  - 不做多平台适配（默认针对 macOS + Homebrew 流程）。
  - 不做持久化状态存储（所有状态来自系统命令探测）。

### 入口地图（可定位文件）

1. `cmd/denv/main.go`：程序入口，执行 `cmd.NewRootCmd()`，统一失败返回码 `1`。
2. `cmd/root.go`：根命令装配（`list/install/outdated/update`）与 `--verbose`。
3. `cmd/tools.go`：命令层端口定义（`ListCommandService` 等）和 `CLIContext` 组装。
4. `cmd/list.go`：参数解析、日志埋点、调用列表服务。
5. `cmd/install.go`：安装命令参数、dry-run、执行委托。
6. `cmd/outdated.go`：过期状态读取、状态聚合展示。
7. `cmd/update.go`：更新计划与执行委托。
8. `cmd/presenter.go`：输出层（plain/json/table/no-color）。
9. `cmd/verbose.go`：日志总线（`doingf/verbosef`）。
10. `internal/service.go`：核心门面（Service），将 `discovery/version/install/outdated/update` 子能力组装到统一 API。
11. `internal/install.go`：安装编排与执行（安装计划、队列、操作执行）。
12. `internal/workflows.go`：list/outdated 工作流核心数据模型（tool item / check state / update plan）。
13. `internal/catalog.go`：工具元数据与安装策略（支持边界、顺序、命令映射）。
14. `internal/version.go`：版本解析与来源（brew/npm）处理。
15. `internal/runtime.go` + `internal/path_policy.go`：系统依赖边界（命令执行、路径归类）。

### 错误总线

- 命令层约定：所有子命令通过 `RunE` 返回 `error`，由 Cobra 打印；`main.go` 统一 `os.Exit(1)`。
- 结构化错误：`internal/workflows.go` 的 `OutdatedError{ToolName,State}` 作为更新阶段阻断信号。
- 退出码策略：当前仅 0/1 两级（成功/失败），未使用子错误码。

### 可复述性验证（建议流程）

#### 1) 冷启动复述测试
- 先只看 `README.md` 和 `*_test.go` 名称。
- 再用以下反例快速验收：
  - `denv outdated`：列出未安装工具时应输出 `<not installed> latest`。
  - `denv install --dry-run`：输出是否是稳定且完整的操作计划。
  - `denv install`（缺少 brew）：应返回失败，而不是静默成功。

#### 2) 入口地图测试
- 10 分钟内应能从上述文件直接找到：入口、配置/目录、核心域、外部依赖边界、错误总线。
- 若找不到，优先从当前文档补齐映射，再补测试锚点。

#### 3) 单元级入口索引（脚本化）
- 详见：[docs/understandability_entry_index.md](/Users/cuimingda/Projects/denv-cli/docs/understandability_entry_index.md)

#### 4) 不变量显性度
- 稳定顺序：
  - `internal/service_test.go`：`TestServiceSupportedToolsAndInstallableOrder`
- idempotence/幂等：
  - `cmd/denv`/`internal` 新增用例见 `internal/service_invariants_test.go`
- 输出格式/数量：
- `TestListCommandPlainOutputLineCountAndOrder`（`cmd/understandability_invariants_test.go`）
