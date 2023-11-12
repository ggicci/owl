package owl

import "context"

// Option is an option for New.
type Option interface {
	Apply(context.Context) context.Context
}

// OptionFunc is a function that implements Option.
type OptionFunc func(context.Context) context.Context

func (f OptionFunc) Apply(ctx context.Context) context.Context {
	return f(ctx)
}

// WithNamespace binds a namespace to the resolver. The namespace is used to
// lookup directive executors.
func WithNamespace(ns *Namespace) Option {
	return WithValue(ckNamespace, ns)
}

// ResolveOption is an option for Resolve.
type ResolveOption Option

// WithValue binds a value to the context. The context is DirectiveRuntime.Context.
// See DirectiveRuntime.Context for more details.
func WithValue(key, value interface{}) ResolveOption {
	return OptionFunc(func(ctx context.Context) context.Context {
		return context.WithValue(ctx, key, value)
	})
}

// WithResolveNestedDirectives controls whether to resolve nested directives.
// The default value is true. When set to false, the nested directives will not
// be executed.
func WithResolveNestedDirectives(resolve bool) ResolveOption {
	return WithValue(ckResolveNestedDirectives, resolve)
}
