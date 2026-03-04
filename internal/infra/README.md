# internal/infra

职责：
- 执行外部依赖：命令发现、命令运行、文件系统/环境探测。

当前实现锚点：
- 运行时抽象：`internal/runtime.go`
- 外部调用与安装：`internal/install.go`
- 版本采集：`internal/version.go`

使用方式：
- 通过 `Runtime` 与 catalog 工厂函数做可替换注入。
- 测试通过覆盖 `executableLookup`、`commandRunner`、`commandRunnerWithOutput`。
