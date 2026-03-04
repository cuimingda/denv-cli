// Package domain holds pure version-comparison rules and helper models.
package domain

import (
	"errors"
	"strings"
	"time"
)

type VersionValueKind int

const (
	VersionValueUnknown VersionValueKind = iota
	VersionValueDate
	VersionValueSemantic
)

type VersionValue struct {
	Raw  string
	Kind VersionValueKind
}

type VersionComparator interface {
	Compare(current string, latest string) int
}

type VersionStrategy = VersionComparator

type SemanticVersionComparator struct{}

func (SemanticVersionComparator) Compare(current string, latest string) int {
	currentParts := splitVersionParts(current)
	latestParts := splitVersionParts(latest)

	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for len(currentParts) < maxLen {
		currentParts = append(currentParts, 0)
	}
	for len(latestParts) < maxLen {
		latestParts = append(latestParts, 0)
	}

	for i := 0; i < maxLen; i++ {
		if currentParts[i] < latestParts[i] {
			return -1
		}
		if currentParts[i] > latestParts[i] {
			return 1
		}
	}

	return 0
}

type DateVersionComparator struct{}

func (DateVersionComparator) Compare(current string, latest string) int {
	currentTime, currentErr := parseDateVersion(current)
	latestTime, latestErr := parseDateVersion(latest)
	if currentErr != nil || latestErr != nil {
		return SemanticVersionComparator{}.Compare(current, latest)
	}

	switch {
	case currentTime.Before(latestTime):
		return -1
	case currentTime.After(latestTime):
		return 1
	default:
		return 0
	}
}

func CompareVersions(current string, latest string) int {
	return CompareVersionsWithStrategy(current, latest, ResolveVersionStrategy(current, latest))
}

func CompareVersionsWithComparator(current string, latest string, comparator VersionComparator) int {
	return CompareVersionsWithStrategy(current, latest, comparator)
}

func CompareVersionsWithStrategy(current string, latest string, strategy VersionStrategy) int {
	if strategy == nil {
		strategy = ResolveVersionStrategy(current, latest)
	}
	return strategy.Compare(current, latest)
}

func ResolveVersionStrategy(current string, latest string) VersionStrategy {
	return ResolveVersionStrategyFromValues(ParseVersionValue(current), ParseVersionValue(latest))
}

func ResolveVersionStrategyFromValues(current VersionValue, latest VersionValue) VersionStrategy {
	if current.Kind == VersionValueDate && latest.Kind == VersionValueDate {
		return DateVersionComparator{}
	}
	return SemanticVersionComparator{}
}

func ParseVersionValue(raw string) VersionValue {
	clean := strings.TrimSpace(raw)
	value := VersionValue{Raw: clean}
	if clean == "" {
		return value
	}

	if _, err := parseDateVersion(clean); err == nil {
		value.Kind = VersionValueDate
		return value
	}

	if isSemanticVersionLike(clean) {
		value.Kind = VersionValueSemantic
		return value
	}

	return value
}

func isSemanticVersionLike(raw string) bool {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '.' || r == '-' || r == '_'
	})
	if len(fields) == 0 {
		return false
	}

	for _, field := range fields {
		if field == "" {
			return false
		}
		for i := 0; i < len(field); i++ {
			ch := field[i]
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

func parseDateVersion(v string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"20060102",
	}

	clean := strings.TrimSpace(v)
	for _, format := range formats {
		t, err := time.Parse(format, clean)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid date format")
}

func seemsLikeDateVersion(current string, latest string) bool {
	_, currentErr := parseDateVersion(current)
	_, latestErr := parseDateVersion(latest)
	return currentErr == nil && latestErr == nil
}

func splitVersionParts(version string) []int {
	fields := strings.FieldsFunc(version, func(r rune) bool {
		return r == '.' || r == '-' || r == '_'
	})

	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		part := 0
		for i := 0; i < len(field); i++ {
			if field[i] < '0' || field[i] > '9' {
				part = -1
				break
			}
			part = part*10 + int(field[i]-'0')
		}
		if part >= 0 {
			parts = append(parts, part)
		}
	}
	return parts
}

func SplitVersionParts(version string) []int {
	return splitVersionParts(version)
}
