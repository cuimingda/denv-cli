// internal provides the CLI-facing contract; domain logic is implemented in internal/domain.
package denv

import "github.com/cuimingda/denv-cli/internal/domain"

type VersionValueKind = domain.VersionValueKind

const (
	VersionValueUnknown  = domain.VersionValueUnknown
	VersionValueDate     = domain.VersionValueDate
	VersionValueSemantic = domain.VersionValueSemantic
)

type VersionValue = domain.VersionValue

type VersionComparator = domain.VersionComparator

type VersionStrategy = domain.VersionStrategy

type SemanticVersionComparator = domain.SemanticVersionComparator

type DateVersionComparator = domain.DateVersionComparator

func CompareVersions(current string, latest string) int {
	return domain.CompareVersions(current, latest)
}

func CompareVersionsWithComparator(current string, latest string, comparator VersionComparator) int {
	return domain.CompareVersionsWithComparator(current, latest, comparator)
}

func CompareVersionsWithStrategy(current string, latest string, strategy VersionStrategy) int {
	return domain.CompareVersionsWithStrategy(current, latest, strategy)
}

func ResolveVersionStrategy(current string, latest string) VersionStrategy {
	return domain.ResolveVersionStrategy(current, latest)
}

func ResolveVersionStrategyFromValues(current VersionValue, latest VersionValue) VersionStrategy {
	return domain.ResolveVersionStrategyFromValues(current, latest)
}

func ParseVersionValue(raw string) VersionValue {
	return domain.ParseVersionValue(raw)
}

func SplitVersionParts(version string) []int {
	return domain.SplitVersionParts(version)
}
