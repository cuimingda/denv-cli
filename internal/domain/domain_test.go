package domain

import (
	"testing"
)

func TestDomainWrappersPreserveVersionSemantics(t *testing.T) {
	got := ParseVersionValue("1.23.4")
	if got.Kind != VersionValueSemantic {
		t.Fatalf("ParseVersionValue semantic mismatch: got kind=%v", got.Kind)
	}

	if cmp := CompareVersions("1.2.3", "1.2.4"); cmp != -1 {
		t.Fatalf("CompareVersions mismatch: got %d, want -1", cmp)
	}

	want := ResolveVersionStrategy("2026-08-01", "2026-08-02")
	if gotStrategy := ResolveVersionStrategyFromValues(ParseVersionValue("2026-08-01"), ParseVersionValue("2026-08-02")); want != gotStrategy {
		t.Fatalf("ResolveVersionStrategyFromValues mismatch")
	}
}

func TestDomainTypeAliasesMatchInternalTypes(t *testing.T) {
	var got VersionValue
	if got.Kind != VersionValueUnknown && got.Kind != VersionValueDate && got.Kind != VersionValueSemantic {
		t.Fatalf("unexpected VersionValue kind: %v", got.Kind)
	}
}
