// Package infra owns external command boundaries.
package infra

import "io"

func ResolveCommandPath(rt Runtime, name string) (string, error) {
	rt = NormalizeRuntime(rt)
	return rt.ExecutableLookup(name)
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
	_, err := ResolveCommandPath(rt, name)
	return err == nil
}
