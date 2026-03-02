# denv

`denv`（Developer Environment）是面向 macOS 软件开发者的命令行工具，用于统一维护常用开发工具链。

## 当前定位

- 产品名：Developer Environment
- 平台：macOS
- 二进制命令：`denv`
- 当前版本：`0.0.1`
- 代码结构：基于 Go + Cobra，子命令驱动的 CLI 架构

## 目标能力

- 一键安装（Install）开发常用软件
- 一键更新（Update）已管理工具
- 查看过期版本（Outdated）
- 切换版本（Use/Switch）用于快速切换活动版本

## 适用场景

为开发者维护一份“常用软件清单”，并在新机器或团队环境中快速完成环境搭建与同步。

## 快速上手

### 构建

```bash
go build ./cmd/denv
```

默认生成的可执行文件名为：`denv`

### 运行

```bash
denv
```

展示当前命令帮助，所有功能预期将以子命令形式提供。

## 说明

该仓库目前已完成 CLI 初始化与版本信息（`0.0.1`）的基础能力。
