package oapi

import (
	"context"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/unbindapp/zigzag-api/ent/schema"
)

const bodyReadTimeout = 60 * time.Second

type Action int

const (
	Read Action = iota
	Create
	Update
	Delete
	Invoke
)

type profile struct {
	errors      []int
	hints       map[string]any
	sessionOnly bool
	privilege   schema.KeyPrivilege
	mcp         *MCPChoice
}

type MCPChoice struct {
	Exposed bool
	Reason  string
}

const (
	metadataAction      = "zigzag.action"
	metadataSessionOnly = "zigzag.session_only"
	metadataPrivilege   = "zigzag.privilege"
	metadataMCP         = "zigzag.mcp"
)

func agentHints(readOnly, destructive, idempotent bool, risk string, confirm bool) map[string]any {
	return map[string]any{
		"readOnlyHint":         readOnly,
		"destructiveHint":      destructive,
		"idempotentHint":       idempotent,
		"openWorldHint":        false,
		"risk":                 risk,
		"requiresConfirmation": confirm,
	}
}

func baseProfile(a Action) profile {
	switch a {
	case Read:
		return profile{
			errors: []int{400, 401, 403, 404, 422, 500},
			hints:  agentHints(true, false, true, "low", false),
		}
	case Create:
		return profile{
			errors: []int{400, 401, 403, 404, 409, 422, 500},
			hints:  agentHints(false, false, false, "medium", false),
		}
	case Update:
		return profile{
			errors: []int{400, 401, 403, 404, 409, 422, 500},
			hints:  agentHints(false, false, true, "medium", false),
		}
	case Delete:
		return profile{
			errors: []int{401, 403, 404, 409, 422, 500},
			hints:  agentHints(false, true, false, "high", true),
		}
	case Invoke:
		return profile{
			errors: []int{400, 401, 403, 404, 409, 422, 500},
			hints:  agentHints(false, false, false, "medium", false),
		}
	default:
		return profile{errors: []int{500}, hints: agentHints(false, false, false, "medium", false)}
	}
}

type Option func(*profile)

func Public(p *profile) {
	p.errors = without(p.errors, 401, 403)
}

func OpenWorld(p *profile) {
	p.hints["openWorldHint"] = true
}

func Confirm(p *profile) {
	p.hints["requiresConfirmation"] = true
}

func Risk(level string) Option {
	return func(p *profile) { p.hints["risk"] = level }
}

func SessionOnly(p *profile) {
	p.sessionOnly = true
}

func Needs(privilege schema.KeyPrivilege) Option {
	return func(p *profile) { p.privilege = privilege }
}

func MCP(p *profile) {
	p.mcp = &MCPChoice{Exposed: true}
}

func NoMCP(reason string) Option {
	return func(p *profile) { p.mcp = &MCPChoice{Reason: reason} }
}

func MarkSessionOnly(op *huma.Operation) {
	if op.Metadata == nil {
		op.Metadata = map[string]any{}
	}
	op.Metadata[metadataSessionOnly] = true
}

func IsSessionOnly(op *huma.Operation) bool {
	if op == nil {
		return false
	}
	v, _ := op.Metadata[metadataSessionOnly].(bool)
	return v
}

func PrivilegeOf(op *huma.Operation) (schema.KeyPrivilege, bool) {
	if op == nil {
		return "", false
	}
	privilege, ok := op.Metadata[metadataPrivilege].(schema.KeyPrivilege)
	return privilege, ok && privilege != ""
}

func MCPChoiceOf(op *huma.Operation) (MCPChoice, bool) {
	if op == nil {
		return MCPChoice{}, false
	}
	choice, ok := op.Metadata[metadataMCP].(MCPChoice)
	return choice, ok
}

func ActionOf(op *huma.Operation) (Action, bool) {
	if op == nil {
		return 0, false
	}
	action, ok := op.Metadata[metadataAction].(Action)
	return action, ok
}

func Register[I, O any](api huma.API, action Action, op huma.Operation, handler func(context.Context, *I) (*O, error), opts ...Option) {
	Apply(action, &op, opts...)
	huma.Register(api, op, handler)
}

func Apply(action Action, op *huma.Operation, opts ...Option) {
	p := baseProfile(action)
	for _, o := range opts {
		o(&p)
	}

	op.Errors = mergeInts(p.errors, op.Errors)

	if op.BodyReadTimeout == 0 {
		op.BodyReadTimeout = bodyReadTimeout
	}

	if op.Metadata == nil {
		op.Metadata = map[string]any{}
	}
	op.Metadata[metadataAction] = action
	if p.sessionOnly {
		MarkSessionOnly(op)
	}
	if p.privilege != "" {
		op.Metadata[metadataPrivilege] = p.privilege
	}
	if p.mcp != nil {
		op.Metadata[metadataMCP] = *p.mcp
	}

	if op.Extensions == nil {
		op.Extensions = map[string]any{}
	}
	if _, ok := op.Extensions["x-agent"]; !ok {
		op.Extensions["x-agent"] = p.hints
	}
}

func without(in []int, drop ...int) []int {
	out := in[:0:0]
	for _, v := range in {
		if !slices.Contains(drop, v) {
			out = append(out, v)
		}
	}
	return out
}

func mergeInts(a, b []int) []int {
	out := append([]int{}, a...)
	for _, v := range b {
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}
