// internal/compare.go 提供版本值解析与版本比较能力，统一处理语义版本、日期版本等不同语义的比较。
package denv

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

// VersionValue 表示解析后的版本原始文本及其类型。
type VersionValue struct {
	Raw  string
	Kind VersionValueKind
}

// VersionComparator 定义版本比较策略。
type VersionComparator interface {
	Compare(current string, latest string) int
}

// VersionStrategy 是兼容别名，强调比较策略语义。
type VersionStrategy = VersionComparator

// SemanticVersionComparator 使用数字段逐位比较。
type SemanticVersionComparator struct{}

func (SemanticVersionComparator) Compare(current string, latest string) int {
	// 先分解版本号为整数切片，保证长度对齐后逐段比较
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

// DateVersionComparator 用日期语义比较（兼容多种日期格式）。
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

// CompareVersions 按自动策略比较当前版本与最新版本。
func CompareVersions(current string, latest string) int {
	return CompareVersionsWithStrategy(current, latest, ResolveVersionStrategy(current, latest))
}

// CompareVersionsWithComparator 用指定比较器比较版本。
func CompareVersionsWithComparator(current string, latest string, comparator VersionComparator) int {
	return CompareVersionsWithStrategy(current, latest, comparator)
}

// CompareVersionsWithStrategy 允许通过策略对象注入比较逻辑。
func CompareVersionsWithStrategy(current string, latest string, strategy VersionStrategy) int {
	if strategy == nil {
		strategy = ResolveVersionStrategy(current, latest)
	}
	return strategy.Compare(current, latest)
}

// ResolveVersionStrategy 根据两侧版本的解析结果自动选择比较策略。
func ResolveVersionStrategy(current string, latest string) VersionStrategy {
	return ResolveVersionStrategyFromValues(ParseVersionValue(current), ParseVersionValue(latest))
}

// ResolveVersionStrategyFromValues 根据版本分类返回对应比较器。
func ResolveVersionStrategyFromValues(current VersionValue, latest VersionValue) VersionStrategy {
	if current.Kind == VersionValueDate && latest.Kind == VersionValueDate {
		return DateVersionComparator{}
	}
	return SemanticVersionComparator{}
}

// ParseVersionValue 将输入字符串归一化并识别类型。
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

// isSemanticVersionLike 判断字符串是否为由数字分段组成的版本。
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

// parseDateVersion 尝试多个日期格式解析。
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

// seemsLikeDateVersion 用于判断两端是否都可按日期解析。
func seemsLikeDateVersion(current string, latest string) bool {
	_, currentErr := parseDateVersion(current)
	_, latestErr := parseDateVersion(latest)
	return currentErr == nil && latestErr == nil
}

// splitVersionParts 将版本拆成整数切片；非法分段会被忽略。
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

// SplitVersionParts 对外导出版本切分函数。
func SplitVersionParts(version string) []int {
	return splitVersionParts(version)
}
