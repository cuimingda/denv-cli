package infra

import "reflect"

// CatalogManager keeps catalog and path policy references with default fallbacks.
type CatalogManager[TCatalog any, TPathPolicy any] struct {
	catalog            TCatalog
	pathPolicy         TPathPolicy
	catalogDefault     func() TCatalog
	pathPolicyDefault  func() TPathPolicy
}

// NewCatalogManager creates a catalog manager with optional fallback factories.
func NewCatalogManager[TCatalog any, TPathPolicy any](catalog TCatalog, pathPolicy TPathPolicy, catalogDefault func() TCatalog, pathPolicyDefault func() TPathPolicy) *CatalogManager[TCatalog, TPathPolicy] {
	return &CatalogManager[TCatalog, TPathPolicy]{
		catalog:           catalog,
		pathPolicy:        pathPolicy,
		catalogDefault:    catalogDefault,
		pathPolicyDefault: pathPolicyDefault,
	}
}

// Catalog returns the catalog with fallback value when unset.
func (m *CatalogManager[TCatalog, TPathPolicy]) Catalog() TCatalog {
	if m == nil || isNilValue(m.catalog) {
		if m == nil || m.catalogDefault == nil {
			var zero TCatalog
			return zero
		}
		return m.catalogDefault()
	}
	return m.catalog
}

// PathPolicy returns the path policy with fallback value when unset.
func (m *CatalogManager[TCatalog, TPathPolicy]) PathPolicy() TPathPolicy {
	if m == nil || isNilValue(m.pathPolicy) {
		if m == nil || m.pathPolicyDefault == nil {
			var zero TPathPolicy
			return zero
		}
		return m.pathPolicyDefault()
	}
	return m.pathPolicy
}

func isNilValue(value any) bool {
	if value == nil {
		return true
	}
	valueOf := reflect.ValueOf(value)
	switch valueOf.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return valueOf.IsNil()
	default:
		return false
	}
}
