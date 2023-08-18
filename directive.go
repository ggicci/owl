package viper

import (
	"context"
	"reflect"
	"strings"
)

// Directive defines the profile to locate a `DirectiveExecutor` instance
// and drives it with essential arguments.
type Directive struct {
	Name string   // name of the executor
	Argv []string // argv
}

// NewDirective creates a Directive instance.
func NewDirective(name string, argv ...string) *Directive {
	return &Directive{
		Name: name,
		Argv: argv,
	}
}

// ParseDirective creates a Directive instance by parsing a directive string
// extracted from the struct tag.
//
// Example directives are:
//
//	"form=page,page_index" -> { Executor: "form", Args: ["page", "page_index"] }
//	"header=x-api-token"   -> { Executor: "header", Args: ["x-api-token"] }
func ParseDirective(directive string) (*Directive, error) {
	directive = strings.TrimSpace(directive)
	parts := strings.SplitN(directive, "=", 2)
	executor := parts[0]
	var argv []string
	if len(parts) == 2 {
		// Split the remained string by delimiter `,` as argv.
		// NOTE: the whiltespaces are kept here.
		// e.g. "query=page, index" -> { Executor: "query", Args: ["page", " index"] }
		argv = strings.Split(parts[1], ",")
	}

	if executor == "" {
		return nil, invalidExecutorName(executor)
	}

	return NewDirective(executor, argv...), nil
}

// DirectiveExecutor is the interface that wraps the Execute method.
// Execute executes the directive by passing the runtime context.
type DirectiveExecutor interface {
	Execute(*DirectiveRuntime) error
}

// DirecrtiveExecutorFunc is an adapter to allow the use of ordinary functions
// as DirectiveExecutors.
type DirectiveExecutorFunc func(*DirectiveRuntime) error

func (f DirectiveExecutorFunc) Execute(de *DirectiveRuntime) error {
	return f(de)
}

// DirectiveRuntime is the execution runtime/context of a directive. NOTE: the
// Directive and Resolver are both exported for the convenience but in an unsafe
// way. The user should not modify them. If you want to modify them, please call
// Resolver.Iterate to iterate the resolvers and modify them in the callback.
// And make sure this be done before any callings to Resolver.Resolve.
type DirectiveRuntime struct {
	Directive *Directive
	Resolver  *Resolver
	Context   context.Context
	Value     reflect.Value
}
