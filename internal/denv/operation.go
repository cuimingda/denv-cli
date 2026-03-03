package denv

import (
	"strings"
)

type CommandSpec struct {
	Name string
	Args []string
}

func NewCommandSpec(name string, args ...string) CommandSpec {
	return CommandSpec{Name: name, Args: append([]string{}, args...)}
}

func (spec CommandSpec) String() string {
	if spec.Name == "" {
		return ""
	}

	if len(spec.Args) == 0 {
		return spec.Name
	}

	return spec.Name + " " + strings.Join(spec.Args, " ")
}

type InstallOperation struct {
	Spec CommandSpec
}

// InstallQueue is an execution queue of install operations.
// Keep it as a dedicated type so orchestration layers can model
// "installation plans" explicitly while still interoperating with
// existing []InstallOperation return types.
type InstallQueue []InstallOperation

func NewInstallQueue(ops ...InstallOperation) InstallQueue {
	return append(InstallQueue(nil), ops...)
}

type InstallPlan struct {
	ToolID     string
	Operations InstallQueue
	Force      bool
}

func NewInstallPlan(toolID string, operations []InstallOperation, force bool) InstallPlan {
	return InstallPlan{
		ToolID:     toolID,
		Operations: append(InstallQueue(nil), operations...),
		Force:      force,
	}
}

func (plan InstallPlan) Len() int {
	return len(plan.Operations)
}

func (plan InstallPlan) ToOperations() []InstallOperation {
	return plan.Operations.ToOperations()
}

func (plan InstallPlan) IsEmpty() bool {
	return plan.Len() == 0
}

func (q InstallQueue) Clone() InstallQueue {
	out := make(InstallQueue, len(q))
	copy(out, q)
	return out
}

func (q InstallQueue) Append(ops ...InstallOperation) InstallQueue {
	out := q.Clone()
	out = append(out, ops...)
	return out
}

func (q InstallQueue) ToOperations() []InstallOperation {
	return q.Clone()
}

func (q InstallQueue) Len() int {
	return len(q)
}

func NewInstallOperation(spec CommandSpec) InstallOperation {
	return InstallOperation{Spec: spec}
}

func InstallOperationsFromSpecs(specs []CommandSpec) []InstallOperation {
	ops := make([]InstallOperation, 0, len(specs))
	for _, spec := range specs {
		ops = append(ops, NewInstallOperation(spec))
	}
	return ops
}

func InstallQueueFromSpecs(specs []CommandSpec) InstallQueue {
	ops := InstallOperationsFromSpecs(specs)
	return NewInstallQueue(ops...)
}

func (op InstallOperation) String() string {
	return op.Spec.String()
}

func ParseCommandSpec(raw string) (CommandSpec, error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return CommandSpec{}, nil
	}

	return NewCommandSpec(fields[0], fields[1:]...), nil
}

func commandSpecsToStrings(ops []CommandSpec) []string {
	out := make([]string, len(ops))
	for i, op := range ops {
		out[i] = op.String()
	}
	return out
}
