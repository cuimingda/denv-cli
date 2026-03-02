package cmd

import (
    "bytes"
    "strings"
    "testing"
)

func TestNewListCmd(t *testing.T) {
    cmd := NewListCmd()
    out := &bytes.Buffer{}
    cmd.SetOut(out)

    if err := cmd.Execute(); err != nil {
        t.Fatalf("list command failed: %v", err)
    }

    got := strings.TrimSpace(out.String())
    want := "php\npython\nnode\ngo"
    if got != want {
        t.Fatalf("unexpected list output:\nwant:\n%q\ngot:\n%q", want, got)
    }
}

func TestIsSupportedTool(t *testing.T) {
    if !IsSupportedTool("php") {
        t.Fatal("expected php to be supported")
    }

    if IsSupportedTool("ruby") {
        t.Fatal("expected ruby to be unsupported")
    }
}

func TestRootHasListCommand(t *testing.T) {
    cmd := NewRootCmd()
    found := false
    for _, sub := range cmd.Commands() {
        if sub.Name() == "list" {
            found = true
            break
        }
    }
    if !found {
        t.Fatal("root command should include list subcommand")
    }
}
