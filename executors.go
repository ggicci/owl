package in

import (
	"context"
	"fmt"
	"reflect"
)

var (
	executors          = make(map[string]DirectiveExecutor)
	reservedDirectives = map[string]struct{}{"ns": {}} // reserved names
)

func init() {
	RegisterDirectiveExecutor("ns", &nsDirective{})
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

// RegisterDirectiveExecutor registers a named executor globally.
// The executor should implement the DirectiveExecutor interface.
// Will panic if the name were taken or the executor is nil.
func RegisterDirectiveExecutor(name string, exe DirectiveExecutor) {
	if _, ok := executors[name]; ok {
		panic(fmt.Errorf("in: %w: %q", ErrDuplicateExecutor, name))
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
	if _, ok := reservedDirectives[name]; ok && executors[name] != nil {
		panic(fmt.Errorf("in: %w: %q", ErrReservedDirectiveName, name))
	}
	executors[name] = exe
}

// LookupExecutor returns the executor by name.
func LookupExecutor(name string) DirectiveExecutor {
	return executors[name]
}

// DirectiveBuildtime is the buildtime context of a directive. Buildtime directives
// can access the resolver and the directive itself and do alterations if needed.
type DirectiveBuildtime struct {
	Directive *Directive
	Resolver  *Resolver
}

// DirectiveRuntime is the execution runtime/context of a directive.
type DirectiveRuntime struct {
	DirectiveBuildtime
	Context context.Context
	value   reflect.Value
}

// IsValueSet returns true if the field value has been set.
func (de *DirectiveRuntime) IsValueSet() bool {
	return de.Context.Value(ContextFieldSet).(bool)
}

// SetValue sets the field value.
func (de *DirectiveRuntime) SetValue(value reflect.Value) {
	de.value = value
	de.Context = context.WithValue(de.Context, ContextFieldSet, true)
}

// GetValue returns the field value.
func (de *DirectiveRuntime) GetValue() reflect.Value {
	return de.value
}

// DirectiveBuild is the interface indicating the directive is a buildtime
// directive. Buildtime directives are executed when the resolver is being
// built.
type DirectiveBuild interface {
	Build(*DirectiveBuildtime) error

	// NoRuntime returns true if the directive does not need to be executed
	// at runtime. For example, the "ns" directive only needs to be executed
	// at buildtime.
	NoRuntime() bool
}

// nsDirective is the namespace directive. It is a buildtime directive.
// Who sets the namespace of the resolver. The namespace is used to
// distinguish the fields with the same name but in different namespaces.
// For a nested field, the namespace will be overwritten
// by concatenating the parent namespaces as a prefix. For example:
//
//	type A struct {
//		B struct {
//			C string `in:"ns=foo"`
//		} `in:"ns=bar"`
//	}
//
// The namespace of field C will be "bar.foo".
type nsDirective struct{} // reserved directive: ns

func (*nsDirective) Build(bt *DirectiveBuildtime) error {
	r, d := bt.Resolver, bt.Directive
	if len(d.Argv) == 0 || d.Argv[0] == "" {
		return ErrMissingNamespace
	}
	r.Context = context.WithValue(r.Context, ContextNamespace, d.Argv[0])
	return nil
}

func (*nsDirective) NoRuntime() bool { return true }

func (*nsDirective) Execute(*DirectiveRuntime) error { return nil }
