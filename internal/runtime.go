package denv

import (
	"io"
	"os/exec"
)

type Runtime struct {
	// ExecutableLookup 查找命令完整路径
	ExecutableLookup       func(string) (string, error)
	// CommandRunner 执行命令并捕获输出
	CommandRunner          func(name string, args ...string) ([]byte, error)
	// CommandRunnerWithOutput 执行命令并将输出直接写入指定 Writer
	CommandRunnerWithOutput func(out io.Writer, name string, args ...string) error
}

// NormalizeRuntime 在缺失回调时补齐默认行为，保证 runtime 对象可直接使用。
func NormalizeRuntime(rt Runtime) Runtime {
	// 未提供查找函数时使用系统 PATH 查询
	if rt.ExecutableLookup == nil {
		rt.ExecutableLookup = exec.LookPath
	}
	// 未提供输出聚合执行器时，默认使用 CombinedOutput
	if rt.CommandRunner == nil {
		rt.CommandRunner = func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		}
	}
	// 未提供带流式输出执行器时，默认将 stdout/stderr 同步到 out
	if rt.CommandRunnerWithOutput == nil {
		rt.CommandRunnerWithOutput = func(out io.Writer, name string, args ...string) error {
			cmd := exec.Command(name, args...)
			cmd.Stdout = out
			cmd.Stderr = out
			return cmd.Run()
		}
	}
	return rt
}
