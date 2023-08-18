package owl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errFoo = fmt.Errorf("foo")
	errBar = fmt.Errorf("bar")
)

func exeFoo(_ *DirectiveRuntime) error {
	return errFoo
}
func exeBar(_ *DirectiveRuntime) error {
	return errBar
}

func TestNamespace(t *testing.T) {
	assert := assert.New(t)
	ns := NewNamespace()

	ns.RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	assert.Equal(ns.LookupExecutor("foo").Execute(nil), errFoo)
	assert.PanicsWithError(duplicatedExecutor("foo").Error(), func() {
		ns.RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	})

	ns.RegisterDirectiveExecutor("bar", DirectiveExecutorFunc(exeBar))
	assert.Equal(ns.LookupExecutor("bar").Execute(nil), errBar)
	assert.PanicsWithError(duplicatedExecutor("bar").Error(), func() {
		ns.RegisterDirectiveExecutor("bar", DirectiveExecutorFunc(exeBar))
	})

	ns.ReplaceDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	assert.Equal(ns.LookupExecutor("foo").Execute(nil), errFoo)
	ns.ReplaceDirectiveExecutor("foo", DirectiveExecutorFunc(exeBar))
	assert.Equal(ns.LookupExecutor("foo").Execute(nil), errBar)
}

func TestNamespaceRegisterNilExecutor(t *testing.T) {
	assert.PanicsWithError(t, nilExecutor("foo").Error(), func() {
		NewNamespace().RegisterDirectiveExecutor("foo", nil)
	})
}

func TestDefaultNamespace(t *testing.T) {
	RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	assert.Equal(t, LookupExecutor("foo").Execute(nil), errFoo)
	assert.PanicsWithError(t, duplicatedExecutor("foo").Error(), func() {
		RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	})

	ReplaceDirectiveExecutor("foo", DirectiveExecutorFunc(exeBar))
	assert.Equal(t, LookupExecutor("foo").Execute(nil), errBar)
}
