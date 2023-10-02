package owl

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrUnsupportedType      = errors.New("unsupported type")
	ErrNilNamespace         = errors.New("nil namespace")
	ErrInvalidDirectiveName = errors.New("invalid directive/executor name")
	ErrDuplicateExecutor    = errors.New("duplicate executor")
	ErrDuplicateDirective   = errors.New("duplicate directive")
	ErrNilExecutor          = errors.New("nil executor")
	ErrMissingExecutor      = errors.New("missing executor")
	ErrTypeMismatch         = errors.New("type mismatch")
	ErrScanNilField         = errors.New("scan nil field")
)

func invalidDirectiveName(name string) error {
	return fmt.Errorf("%w: %q (should comply with %s)", ErrInvalidDirectiveName, name, reDirectiveName.String())
}

func duplicateExecutor(name string) error {
	return fmt.Errorf("%w: %q (registered to the same namespace)", ErrDuplicateExecutor, name)
}

func duplicateDirective(name string) error {
	return fmt.Errorf("%w: %q (defined in the same struct tag)", ErrDuplicateDirective, name)
}

func nilExecutor(name string) error {
	return fmt.Errorf("%w: %q", ErrNilExecutor, name)
}

type ResolveError struct {
	fieldError
}

func (e *ResolveError) Error() string {
	return fmt.Sprintf("resolve field %q failed: %s", e.Resolver.String(), e.Err)
}

type ScanError struct {
	fieldError
}

func (e *ScanError) Error() string {
	return fmt.Sprintf("scan field %q failed: %s", e.Resolver.String(), e.Err)
}

type ScanErrors []*ScanError

func (e ScanErrors) Error() string {
	var errs []string
	pe := e
	if len(e) > 3 {
		pe = e[:3]
	}
	for _, se := range pe {
		errs = append(errs, se.Error())
	}
	rest := len(e) - 3
	if rest > 0 {
		errs = append(errs, fmt.Sprintf("...(%d more)", rest))
	}
	return fmt.Sprintf("scan errors: %s", strings.Join(errs, "; "))
}

type fieldError struct {
	Err      error
	Resolver *Resolver
}

func (e *fieldError) Unwrap() error {
	return e.Err
}

func (e *fieldError) AsDirectiveExecutionError() *DirectiveExecutionError {
	var de *DirectiveExecutionError
	errors.As(e.Err, &de)
	return de
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

type scanErrorSink struct {
	errors ScanErrors
}

func (es *scanErrorSink) Add(resolver *Resolver, err error) {
	es.errors = append(es.errors, &ScanError{
		fieldError: fieldError{
			Err:      err,
			Resolver: resolver,
		},
	})
}
