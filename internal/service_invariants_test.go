package denv

import (
	"io"
	"os/exec"
	"reflect"
	"testing"
)

func TestBuildInstallQueueIsStableAcrossCalls(t *testing.T) {
	rt := Runtime{
		ExecutableLookup: func(name string) (string, error) {
			// Keep test stable and environment-independent: pretend brew exists so
			// install planning does not short-circuit on host environment differences.
			if name == "brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", exec.ErrNotFound
		},
		CommandRunner: func(string, ...string) ([]byte, error) {
			return []byte{}, nil
		},
		CommandRunnerWithOutput: func(io.Writer, string, ...string) error {
			return nil
		},
	}
	svc := NewService(rt)

	first, err := svc.BuildInstallQueue(false)
	if err != nil {
		t.Fatalf("build install queue failed: %v", err)
	}

	second, err := svc.BuildInstallQueue(false)
	if err != nil {
		t.Fatalf("build install queue second time failed: %v", err)
	}

	if len(first) != len(second) {
		t.Fatalf("expected stable queue lengths, got %d and %d", len(first), len(second))
	}

	if !reflect.DeepEqual(first.ToOperations(), second.ToOperations()) {
		t.Fatalf("expected stable build results for repeated calls")
	}
}

func TestListToolItemsOrderMatchesCatalogList(t *testing.T) {
	rt := Runtime{
		ExecutableLookup: func(string) (string, error) {
			return "", exec.ErrNotFound
		},
		CommandRunner: func(string, ...string) ([]byte, error) {
			return []byte{}, nil
		},
	}

	svc := NewService(rt)
	items, err := svc.ListToolItems(ListOptions{})
	if err != nil {
		t.Fatalf("list tool items failed: %v", err)
	}

	expected := svc.SupportedTools()
	if len(items) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(items))
	}

	for i, want := range expected {
		if items[i].Name != want {
			t.Fatalf("order drift at index %d: want=%q got=%q", i, want, items[i].Name)
		}
	}
}
