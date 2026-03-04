package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestListCommandPlainOutputLineCountAndOrder(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		return "", exec.ErrNotFound
	}
	commandRunner = func(_ string, _ ...string) ([]byte, error) {
		return nil, nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	cmd := NewListCmdWithService(testCommandService())
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	want := SupportedTools()
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %q", len(want), len(lines), out.String())
	}

	for i, name := range want {
		if strings.TrimSpace(lines[i]) != name {
			t.Fatalf("order drift at line %d: want=%q got=%q", i, name, strings.TrimSpace(lines[i]))
		}
	}
}
