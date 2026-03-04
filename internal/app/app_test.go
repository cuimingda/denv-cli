package app

import (
	"testing"

	denv "github.com/cuimingda/denv-cli/internal"
)

func TestNewService_ComposesDomainWorkflow(t *testing.T) {
	rt := denv.Runtime{
		ExecutableLookup: func(name string) (string, error) {
			return "/bin/" + name, nil
		},
		CommandRunner: func(name string, args ...string) ([]byte, error) {
			return []byte(""), nil
		},
	}

	svc := NewService(rt)
	if svc == nil {
		t.Fatal("expected service, got nil")
	}

	items, err := svc.ListToolItems(denv.ListOptions{})
	if err != nil {
		t.Fatalf("list tool items failed: %v", err)
	}

	if len(items) != len(svc.SupportedTools()) {
		t.Fatalf("unexpected item count: got %d, want %d", len(items), len(svc.SupportedTools()))
	}
}

