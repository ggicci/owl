package in

import (
	"fmt"
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
	parts := strings.SplitN(directive, "=", 2)
	executor := parts[0]
	var argv []string
	if len(parts) == 2 {
		// Split the remained string by delimiter `,` as argv.
		argv = strings.Split(parts[1], ",")
	}

	if executor == "" {
		return nil, fmt.Errorf("invalid executor name: %q", executor)
	}

	return NewDirective(executor, argv...), nil
}
