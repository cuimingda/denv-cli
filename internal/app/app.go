// Package app composes domain workflows used by command entrypoints.
package app

import denv "github.com/cuimingda/denv-cli/internal"

// Service is an adapter alias for the concrete orchestrator used by command layer.
type Service = denv.Service

// NewService builds the runtime-parameterized app service used by command handlers.
func NewService(rt denv.Runtime) *Service {
	return denv.NewService(rt)
}
