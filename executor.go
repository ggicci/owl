package viper

import (
	"context"
	"fmt"
	"reflect"
)

var (
	executors = make(map[string]DirectiveExecutor)
)

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

// RegisterDirectiveExecutor registers a named executor globally.
// The executor should implement the DirectiveExecutor interface.
// Will panic if the name were taken or the executor is nil.
func RegisterDirectiveExecutor(name string, exe DirectiveExecutor) {
	if _, ok := executors[name]; ok {
		panic(fmt.Errorf("in: %w: %q", ErrDuplicatedExecutor, name))
	}
	ReplaceDirectiveExecutor(name, exe)
}

// ReplaceDirectiveExecutor works like RegisterDirectiveExecutor without panic
// on duplicate names. While it will panic if the executor is nil or the name
// is reserved. Currently, the reserved names are "decoder" and "ns".
func ReplaceDirectiveExecutor(name string, exe DirectiveExecutor) {
	if exe == nil {
		panic(fmt.Errorf("in: %w: %q", ErrNilExecutor, name))
	}
	executors[name] = exe
}

// LookupExecutor returns the executor by name.
func LookupExecutor(name string) DirectiveExecutor {
	return executors[name]
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
