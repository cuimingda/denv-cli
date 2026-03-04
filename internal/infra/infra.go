// Package infra keeps I/O adapters and external command boundaries.
package infra

import (
	denv "github.com/cuimingda/denv-cli/internal"
	"io"
)

type Runtime = denv.Runtime

func NormalizeRuntime(rt Runtime) Runtime {
	return denv.NormalizeRuntime(rt)
}

func ResolveCommandPath(rt Runtime, name string) (string, error) {
	return denv.CommandPath(rt, name)
}

func RunCommand(rt Runtime, name string, args ...string) ([]byte, error) {
	rt = NormalizeRuntime(rt)
	return rt.CommandRunner(name, args...)
}

func RunCommandWithOutput(rt Runtime, out io.Writer, name string, args ...string) error {
	rt = NormalizeRuntime(rt)
	return rt.CommandRunnerWithOutput(out, name, args...)
}

func IsCommandAvailable(rt Runtime, name string) bool {
	return denv.IsCommandAvailable(rt, name)
}
