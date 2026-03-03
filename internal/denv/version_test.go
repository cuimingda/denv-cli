package denv

import (
	"errors"
	"strings"
	"testing"
)

func TestParseVersionValue(t *testing.T) {
	cases := []struct {
		raw  string
		kind VersionValueKind
	}{
		{"2026-08-21", VersionValueDate},
		{"2026/08/21", VersionValueDate},
		{"20260821", VersionValueDate},
		{"1.23.4", VersionValueSemantic},
		{"8.0_1", VersionValueSemantic},
		{"invalid", VersionValueUnknown},
	}

	for _, c := range cases {
		value := ParseVersionValue(c.raw)
		if value.Kind != c.kind {
			t.Fatalf("ParseVersionValue(%q).Kind = %d, expected %d", c.raw, value.Kind, c.kind)
		}
	}
}

func TestResolveVersionStrategySelection(t *testing.T) {
	cases := []struct {
		current string
		latest  string
		isDate  bool
	}{
		{"2026-08-21", "2026-08-22", true},
		{"1.2.3", "1.2.4", false},
		{"1.2.3", "2026-08-21", false},
	}

	for _, c := range cases {
		got := ResolveVersionStrategy(c.current, c.latest)
		_, gotDate := got.(DateVersionComparator)
		if gotDate != c.isDate {
			t.Fatalf("ResolveVersionStrategy(%q, %q) date=%v, expected %v", c.current, c.latest, gotDate, c.isDate)
		}
	}
}

func TestExtractVersion(t *testing.T) {
	got, err := ExtractVersion("go version go1.23.4 darwin/arm64")
	if err != nil {
		t.Fatalf("extract version failed: %v", err)
	}
	if got != "1.23.4" {
		t.Fatalf("expected 1.23.4, got %q", got)
	}
}

func TestExtractBrewCurrentInstallVersionPrefersJSONInstalledVersion(t *testing.T) {
	payload := `{"formulae":[{"name":"ffmpeg","versions":{"stable":"8.0.1"},"installed":[{"version":"8.0_1"}]}]}`

	got := extractBrewCurrentInstallVersion(payload)
	if got != "8.0_1" {
		t.Fatalf("expected installed version 8.0_1 from json, got %q", got)
	}
}

func TestExtractBrewCurrentInstallVersionFallsBackToCellarPath(t *testing.T) {
	legacyOutput := `Installed
/opt/homebrew/Cellar/php/8.0.0 (123 files) *
/opt/homebrew/Cellar/php/8.1.0 (130 files)`

	got := extractBrewCurrentInstallVersion(legacyOutput)
	if got != "8.0.0" {
		t.Fatalf("expected fallback star-marked version 8.0.0, got %q", got)
	}
}

func TestToolVersionFromBrewListUsesInstalledVersionFromJSON(t *testing.T) {
	rt := Runtime{
		CommandRunner: func(name string, args ...string) ([]byte, error) {
			if name != "brew" {
				return nil, errors.New("unexpected command")
			}

			if len(args) == 3 && args[0] == "info" && args[1] == "--json=v2" && args[2] == "ffmpeg" {
				payload := `{"formulae":[{"name":"ffmpeg","versions":{"stable":"8.0.2"},"installed":[{"version":"8.0_1"}]}]}`
				return []byte(payload), nil
			}

			return []byte(strings.TrimSpace(`ffmpeg ✔: stable 8.0.2 (bottled), HEAD`)), nil
		},
	}

	got, err := toolVersionFromBrewList(rt, "ffmpeg")
	if err != nil {
		t.Fatalf("toolVersionFromBrewList failed: %v", err)
	}
	if got != "8.0_1" {
		t.Fatalf("expected 8.0_1, got %q", got)
	}
}
