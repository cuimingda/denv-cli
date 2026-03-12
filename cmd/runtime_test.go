// cmd/runtime_test.go 提供命令层测试可复用的 Runtime 伪造对象与测试服务创建入口。
package cmd

import (
	"io"
	"os/exec"

	"github.com/cuimingda/denv-cli/internal"
)

var (
	executableLookup        = exec.LookPath
	commandRunner           = func(name string, args ...string) ([]byte, error) { return exec.Command(name, args...).CombinedOutput() }
	commandRunnerWithOutput = func(out io.Writer, name string, args ...string) error {
		output, err := commandRunner(name, args...)
		if len(output) > 0 {
			_, _ = out.Write(output)
		}
		return err
	}
)

func testRuntime() denv.Runtime {
	return denv.Runtime{
		ExecutableLookup:        executableLookup,
		CommandRunner:           commandRunner,
		CommandRunnerWithOutput: commandRunnerWithOutput,
	}
}

func testCommandService() *denv.Service {
	return NewCLIContextWithRuntime(testRuntime()).service
}
