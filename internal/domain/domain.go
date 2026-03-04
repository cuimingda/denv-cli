// Package domain keeps domain-level contracts and value models.
package domain

import denv "github.com/cuimingda/denv-cli/internal"

type (
	OutdatedState     = denv.OutdatedState
	ToolListItem      = denv.ToolListItem
	OutdatedItem      = denv.OutdatedItem
	ToolCheckResult   = denv.ToolCheckResult
	VersionValue      = denv.VersionValue
	VersionValueKind  = denv.VersionValueKind
	VersionComparator = denv.VersionComparator
	VersionStrategy   = denv.VersionStrategy
)

var (
	VersionValueDate     = denv.VersionValueDate
	VersionValueSemantic = denv.VersionValueSemantic
	VersionValueUnknown  = denv.VersionValueUnknown
)

func ParseVersionValue(raw string) denv.VersionValue    { return denv.ParseVersionValue(raw) }
func CompareVersions(current string, latest string) int { return denv.CompareVersions(current, latest) }
func CompareVersionsWithComparator(current string, latest string, comparator VersionComparator) int {
	return denv.CompareVersionsWithComparator(current, latest, comparator)
}
func ResolveVersionStrategy(current string, latest string) VersionStrategy {
	return denv.ResolveVersionStrategy(current, latest)
}
func ResolveVersionStrategyFromValues(current VersionValue, latest VersionValue) VersionStrategy {
	return denv.ResolveVersionStrategyFromValues(current, latest)
}
