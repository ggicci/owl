package in

import (
	"errors"
	"fmt"
)

var (
	ErrDuplicateExecutor         = errors.New("duplicate executor")
	ErrNilExecutor               = errors.New("nil executor")
	ErrReservedDirectiveName     = errors.New("reserved directive name")
	ErrDirectiveExecutorNotFound = errors.New("directive executor not found")
	ErrMissingNamespace          = errors.New("missing namespace")
)

type FieldResolveError struct {
	Err   error
	Index int
	Field *Resolver
}

func (e *FieldResolveError) Error() string {
	return fmt.Sprintf("resolve field#%d %q failed: %s", e.Index, e.Field.PathString(), e.Err)
}

func (e *FieldResolveError) Unwrap() error {
	return e.Err
}

type DirectiveExecutionError struct {
	Err error
	Directive
}

func (e *DirectiveExecutionError) Error() string {
	return fmt.Sprintf("execute directive %q failed: %s", e.Directive.Name, e.Err)
}

func (e *DirectiveExecutionError) Unwrap() error {
	return e.Err
}
