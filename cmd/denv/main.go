package main

import (
	"os"

	"github.com/cuimingda/denv-cli/cmd"
)

// main 为 denv CLI 创建根命令并执行；
// 执行失败时以非零状态码退出，便于脚本与 CI 检测。
func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		// 只做统一失败码返回，不吞掉错误信息，
		// 因为具体错误已在命令层输出到标准错误。
		os.Exit(1)
	}
}
