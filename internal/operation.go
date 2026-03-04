// internal/operation.go 定义命令规格与安装动作模型（Operation/Queue/Plan），并提供构造、转换和基本操作能力。
package denv

import (
	"strings"
)

type CommandSpec struct {
	Name string
	Args []string
}

// NewCommandSpec 创建命令规格，参数会被拷贝，避免外部切片被修改影响内部状态。
func NewCommandSpec(name string, args ...string) CommandSpec {
	return CommandSpec{Name: name, Args: append([]string{}, args...)}
}

// String 返回可读命令文本，便于日志与 dry-run 展示。
func (spec CommandSpec) String() string {
	if spec.Name == "" {
		return ""
	}

	if len(spec.Args) == 0 {
		return spec.Name
	}

	return spec.Name + " " + strings.Join(spec.Args, " ")
}

// InstallOperation 表示一个待执行安装动作。
type InstallOperation struct {
	Spec CommandSpec
}

// InstallQueue 是可执行安装动作队列。
type InstallQueue []InstallOperation

// NewInstallQueue 按顺序创建安装队列。
func NewInstallQueue(ops ...InstallOperation) InstallQueue {
	return append(InstallQueue(nil), ops...)
}

// InstallPlan 表示单个工具的安装执行计划。
type InstallPlan struct {
	ToolID     string
	Operations InstallQueue
	Force      bool
}

// NewInstallPlan 创建安装计划快照。
func NewInstallPlan(toolID string, operations []InstallOperation, force bool) InstallPlan {
	return InstallPlan{
		ToolID:     toolID,
		Operations: append(InstallQueue(nil), operations...),
		Force:      force,
	}
}

// Len 返回队列长度。
func (plan InstallPlan) Len() int {
	return len(plan.Operations)
}

// ToOperations 通过副本暴露底层动作，避免外部直接修改内部切片。
func (plan InstallPlan) ToOperations() []InstallOperation {
	return plan.Operations.ToOperations()
}

// IsEmpty 判断是否有可执行动作。
func (plan InstallPlan) IsEmpty() bool {
	return plan.Len() == 0
}

// Clone 克隆一份独立队列。
func (q InstallQueue) Clone() InstallQueue {
	out := make(InstallQueue, len(q))
	copy(out, q)
	return out
}

// Append 追加动作并返回新队列。
func (q InstallQueue) Append(ops ...InstallOperation) InstallQueue {
	out := q.Clone()
	out = append(out, ops...)
	return out
}

// ToOperations 返回可修改的副本，保证调用方安全。
func (q InstallQueue) ToOperations() []InstallOperation {
	return q.Clone()
}

// Len 返回动作数量。
func (q InstallQueue) Len() int {
	return len(q)
}

// NewInstallOperation 创建安装动作。
func NewInstallOperation(spec CommandSpec) InstallOperation {
	return InstallOperation{Spec: spec}
}

// InstallOperationsFromSpecs 将命令规格序列转换为安装动作序列。
func InstallOperationsFromSpecs(specs []CommandSpec) []InstallOperation {
	ops := make([]InstallOperation, 0, len(specs))
	for _, spec := range specs {
		ops = append(ops, NewInstallOperation(spec))
	}
	return ops
}

// InstallQueueFromSpecs 创建安装队列。
func InstallQueueFromSpecs(specs []CommandSpec) InstallQueue {
	ops := InstallOperationsFromSpecs(specs)
	return NewInstallQueue(ops...)
}

// String 输出该动作用于展示或日志。
func (op InstallOperation) String() string {
	return op.Spec.String()
}

// ParseCommandSpec 按空白切分命令文本，返回 CommandSpec。
func ParseCommandSpec(raw string) (CommandSpec, error) {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return CommandSpec{}, nil
	}

	return NewCommandSpec(fields[0], fields[1:]...), nil
}

// commandSpecsToStrings 将命令列表序列化为可读字符串列表。
func commandSpecsToStrings(ops []CommandSpec) []string {
	out := make([]string, len(ops))
	for i, op := range ops {
		out[i] = op.String()
	}
	return out
}
