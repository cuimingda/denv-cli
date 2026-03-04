// internal keeps directory-level path-policy boundary as an adapter to domain model.
package denv

import "github.com/cuimingda/denv-cli/internal/domain"

type PathPolicy = domain.PathPolicy

func DefaultPathPolicy() PathPolicy {
	return domain.DefaultPathPolicy()
}
