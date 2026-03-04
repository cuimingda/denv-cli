package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/cuimingda/denv-cli/internal"
)

func TestUnderstandabilityInvariant_RootHelpExposesCoreCommands(t *testing.T) {
	root := NewRootCmd()
	out := &strings.Builder{}
	root.SetOut(io.Writer(out))
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("root help failed: %v", err)
	}

	got := out.String()
	required := []string{"list", "install", "outdated", "update", "--verbose"}
	for _, token := range required {
		if !strings.Contains(got, token) {
			t.Fatalf("help output missing required token %q", token)
		}
	}
}

func TestUnderstandabilityInvariant_ListOutputOrderAndJSONSchemaStable(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(string) (string, error) { return "", exec.ErrNotFound }
	commandRunner = func(string, ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	supported := SupportedTools()

	plainCmd := NewListCmdWithService(testCommandService())
	plainOut := &strings.Builder{}
	plainCmd.SetOut(io.Writer(plainOut))
	if err := plainCmd.Execute(); err != nil {
		t.Fatalf("plain list failed: %v", err)
	}
	plainLines := strings.Split(strings.TrimSpace(plainOut.String()), "\n")
	if len(plainLines) != len(supported) {
		t.Fatalf("plain list count=%d, want=%d", len(plainLines), len(supported))
	}
	for i, name := range supported {
		if strings.TrimSpace(plainLines[i]) != name {
			t.Fatalf("plain list order drift at %d: got=%q want=%q", i, plainLines[i], name)
		}
	}

	jsonCmd := NewListCmdWithService(testCommandService())
	jsonOut := &strings.Builder{}
	jsonCmd.SetArgs([]string{"--output", "json"})
	jsonCmd.SetOut(io.Writer(jsonOut))
	if err := jsonCmd.Execute(); err != nil {
		t.Fatalf("json list failed: %v", err)
	}

	type item struct {
		Name          string `json:"name"`
		DisplayName   string `json:"display_name"`
		Installed     bool   `json:"installed"`
		Path          string `json:"path"`
		ManagedByBrew bool   `json:"managed_by_brew"`
	}

	var got []item
	if err := json.Unmarshal([]byte(jsonOut.String()), &got); err != nil {
		t.Fatalf("parse list json failed: %v", err)
	}
	if len(got) != len(supported) {
		t.Fatalf("json list count=%d, want=%d", len(got), len(supported))
	}
	for i, item := range got {
		if item.Name != supported[i] {
			t.Fatalf("json order drift at %d: got=%q want=%q", i, item.Name, supported[i])
		}
		if item.DisplayName == "" {
			t.Fatalf("json item %q missing display_name", item.Name)
		}
		if item.Path != "" && item.Installed != true {
			t.Fatalf("path should be empty when item not installed: %q", item.Name)
		}
	}
}

func TestUnderstandabilityInvariant_OutdatedJSONContract(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(string) (string, error) { return "", exec.ErrNotFound }
	commandRunner = func(name string, args ...string) ([]byte, error) {
		if name == "brew" && len(args) >= 3 && args[0] == "info" {
			formula := args[len(args)-1]
			payload := fmt.Sprintf(`{"formulae":[{"name":"%s","versions":{"stable":"8.0.0"}}]}`, formula)
			return []byte(payload), nil
		}
		if name == "npm" && len(args) == 3 && args[0] == "view" && args[1] == "npm" && args[2] == "version" {
			return []byte("11.0.0"), nil
		}
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	cmd := NewOutdatedCmdWithService(testCommandService())
	cmd.SetArgs([]string{"--output", "json"})
	out := &strings.Builder{}
	cmd.SetOut(io.Writer(out))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("outdated json failed: %v", err)
	}

	type row struct {
		Name    string `json:"name"`
		State   string `json:"state"`
		Current string `json:"current"`
		Latest  string `json:"latest"`
	}
	var rows []row
	if err := json.Unmarshal([]byte(out.String()), &rows); err != nil {
		t.Fatalf("parse outdated json failed: %v", err)
	}
	if len(rows) != len(SupportedTools()) {
		t.Fatalf("outdated row count=%d, want=%d", len(rows), len(SupportedTools()))
	}

	seenNotInstalled := 0
	for _, row := range rows {
		if row.Name == "" {
			t.Fatal("outdated json missing name")
		}
		if row.State == "" {
			t.Fatalf("outdated %q missing state", row.Name)
		}
		if row.Current == "" {
			t.Fatalf("outdated %q missing current field", row.Name)
		}
		if row.Current == "<not installed>" {
			seenNotInstalled++
		}
		if row.State == string(denv.OutdatedStateNotInstalled) && row.Latest == "" {
			t.Fatalf("outdated %q missing latest when not installed", row.Name)
		}
	}
	if seenNotInstalled == 0 {
		t.Fatal("expected at least one not_installed row in outdated json")
	}
}

func TestUnderstandabilityInvariant_UpdatePlanFailsFastOnInvalidCurrentVersion(t *testing.T) {
	oldLookup := executableLookup
	oldRunner := commandRunner
	executableLookup = func(name string) (string, error) {
		if name == "tree" {
			return "/usr/local/bin/tree", nil
		}
		return "", exec.ErrNotFound
	}
	commandRunner = func(string, ...string) ([]byte, error) {
		return []byte(""), nil
	}
	defer func() {
		executableLookup = oldLookup
		commandRunner = oldRunner
	}()

	svc := testCommandService()
	_, err := svc.OutdatedUpdatePlan()
	if err == nil {
		t.Fatal("expected update plan to fail when current version is invalid for installed tool")
	}

	var outdatedErr *denv.OutdatedError
	if !errors.As(err, &outdatedErr) {
		t.Fatalf("expected OutdatedError, got %T: %v", err, err)
	}
	if outdatedErr.ToolName != "tree" {
		t.Fatalf("expected fail tool=%q, got=%q", "tree", outdatedErr.ToolName)
	}
	if outdatedErr.State != denv.OutdatedStateInvalidCurrent {
		t.Fatalf("expected state=%s, got=%s", denv.OutdatedStateInvalidCurrent, outdatedErr.State)
	}
}
