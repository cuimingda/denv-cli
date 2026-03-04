package infra

import (
	"bytes"
	"io"
	"os/exec"
	"testing"
)

func TestNormalizeRuntimeProvidesDefaults(t *testing.T) {
	rt := NormalizeRuntime(Runtime{})
	if rt.ExecutableLookup == nil || rt.CommandRunner == nil || rt.CommandRunnerWithOutput == nil {
		t.Fatal("NormalizeRuntime must provide all default callbacks")
	}
}

func TestRunCommandRespectsRuntimeAndResolvers(t *testing.T) {
	called := false
	run := Runtime{
		CommandRunner: func(name string, args ...string) ([]byte, error) {
			called = true
			if name != "echo" || len(args) != 1 || args[0] != "ok" {
				t.Fatalf("unexpected command args: %s %v", name, args)
			}
			return []byte("ok"), nil
		},
	}
	got, err := RunCommand(run, "echo", "ok")
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if !called {
		t.Fatal("expected CommandRunner to be called")
	}
	if string(got) != "ok" {
		t.Fatalf("expected output %q, got %q", "ok", string(got))
	}
}

func TestRunCommandWithOutputForwardsWriterAndPath(t *testing.T) {
	seen := struct {
		name string
		out  []byte
	}{}

	rt := Runtime{
		CommandRunnerWithOutput: func(out io.Writer, name string, args ...string) error {
			seen.name = name
			_, err := out.Write([]byte("done"))
			return err
		},
	}
	var buf bytes.Buffer
	if err := RunCommandWithOutput(rt, &buf, "brew", "list"); err != nil {
		t.Fatalf("RunCommandWithOutput failed: %v", err)
	}
	if seen.name != "brew" || !bytes.Equal(buf.Bytes(), []byte("done")) {
		t.Fatalf("unexpected command forwarding: name=%q out=%q", seen.name, buf.String())
	}
}

func TestResolveCommandPathAndAvailability(t *testing.T) {
	rt := Runtime{
		ExecutableLookup: func(name string) (string, error) {
			if name == "go" {
				return "/usr/bin/go", nil
			}
			return "", exec.ErrNotFound
		},
	}

	got, err := ResolveCommandPath(rt, "go")
	if err != nil || got != "/usr/bin/go" {
		t.Fatalf("ResolveCommandPath failed: got=%q err=%v", got, err)
	}
	if !IsCommandAvailable(rt, "go") {
		t.Fatal("expected command available")
	}
	if IsCommandAvailable(rt, "missing") {
		t.Fatal("expected missing command unavailable")
	}
}

func TestRunCommandUsesDefaultRuntimeFallback(t *testing.T) {
	got, err := RunCommand(Runtime{}, "go", "version")
	if err != nil {
		t.Fatal("expected default runtime fallback to execute go")
	}
	if len(got) == 0 {
		t.Fatal("expected non-empty output from go version")
	}
}
