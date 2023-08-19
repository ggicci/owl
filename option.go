package owl

import "context"

type Option interface {
	Apply(*Resolver) error
}

type OptionFunc func(*Resolver) error

func (f OptionFunc) Apply(r *Resolver) error {
	return f(r)
}

type withNamespace struct {
	ns *Namespace
}

func (wn *withNamespace) Apply(r *Resolver) error {
	r.Context = context.WithValue(r.Context, ckNamespace, wn.ns)
	return nil
}

func WithNamespace(ns *Namespace) Option {
	return &withNamespace{ns}
}

func normalizeOptions(opts []Option) []Option {
	hasNamespace := false
	for _, opt := range opts {
		if optWithNamespace, ok := opt.(*withNamespace); ok && optWithNamespace != nil {
			hasNamespace = true
			break
		}
	}

	// Add default namespace if no namespace option provided.
	if !hasNamespace {
		opts = append(opts, WithNamespace(defaultNS))
	}
	return opts
}

type ResolveOption interface {
	Apply(context.Context) context.Context
}

type ResolveOptionFunc func(context.Context) context.Context

func (f ResolveOptionFunc) Apply(ctx context.Context) context.Context {
	return f(ctx)
}

func WithValue(key, value interface{}) ResolveOption {
	return ResolveOptionFunc(func(ctx context.Context) context.Context {
		return context.WithValue(ctx, key, value)
	})
}
