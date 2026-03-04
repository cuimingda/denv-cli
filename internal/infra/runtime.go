// Package infra owns external runtime boundaries and command invocation adapters.
package infra

import (
	"io"
	"os/exec"
)

type Runtime struct {
	ExecutableLookup        func(string) (string, error)
	CommandRunner           func(name string, args ...string) ([]byte, error)
	CommandRunnerWithOutput func(out io.Writer, name string, args ...string) error
}

func NormalizeRuntime(rt Runtime) Runtime {
	if rt.ExecutableLookup == nil {
		rt.ExecutableLookup = exec.LookPath
	}
	if rt.CommandRunner == nil {
		rt.CommandRunner = func(name string, args ...string) ([]byte, error) {
			return exec.Command(name, args...).CombinedOutput()
		}
	}
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
