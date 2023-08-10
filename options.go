package rstruct

import "context"

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
