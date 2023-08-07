package in

type ContextKey int

const (
	// ContextFieldSet is the key that indicates the field has already been set
	// by a former directive execution.
	ContextFieldSet ContextKey = iota

	// ContextNamespace is the key that indicates the namespace of the resolver.
	ContextNamespace
)
