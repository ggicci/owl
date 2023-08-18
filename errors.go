package owl

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidExecutorName       = errors.New("invalid executor name")
	ErrDuplicatedExecutor        = errors.New("duplicated executor")
	ErrNilExecutor               = errors.New("nil executor")
	ErrDirectiveExecutorNotFound = errors.New("directive executor not found")
)

func invalidExecutorName(name string) error {
	return fmt.Errorf("%w: %q", ErrInvalidExecutorName, name)
}

func duplicatedExecutor(name string) error {
	return fmt.Errorf("%w: %q", ErrDuplicatedExecutor, name)
}

func nilExecutor(name string) error {
	return fmt.Errorf("%w: %q", ErrNilExecutor, name)
}

type ResolveError struct {
	Err      error
	Index    int
	Resolver *Resolver
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("resolve field#%d %q failed: %s", e.Index, e.Resolver.PathString(), e.Err)
}

func (e *ResolveError) Unwrap() error {
	return e.Err
}

type DirectiveExecutionError struct {
	Err error
	Directive
}

func (e *DirectiveExecutionError) Error() string {
	return fmt.Sprintf("execute directive %q with args %v failed: %s", e.Directive.Name, e.Directive.Argv, e.Err)
}

func (e *DirectiveExecutionError) Unwrap() error {
	return e.Err
}
