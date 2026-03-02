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
