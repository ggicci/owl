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
	assert.PanicsWithError("owl: "+duplicateExecutor("foo").Error(), func() {
		ns.RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	})

	ns.RegisterDirectiveExecutor("bar", DirectiveExecutorFunc(exeBar))
	assert.Equal(ns.LookupExecutor("bar").Execute(nil), errBar)
	assert.PanicsWithError("owl: "+duplicateExecutor("bar").Error(), func() {
		ns.RegisterDirectiveExecutor("bar", DirectiveExecutorFunc(exeBar))
	})

	ns.RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo), true)
	assert.Equal(ns.LookupExecutor("foo").Execute(nil), errFoo)
	ns.RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeBar), true)
	assert.Equal(ns.LookupExecutor("foo").Execute(nil), errBar)
}

func TestNamespace_RegisterNilExecutor(t *testing.T) {
	assert.PanicsWithError(t, "owl: "+nilExecutor("foo").Error(), func() {
		NewNamespace().RegisterDirectiveExecutor("foo", nil)
	})
}

func TestNamespace_RegisterInvalidName(t *testing.T) {
	assert.PanicsWithError(t, "owl: "+invalidDirectiveName(".foo").Error(), func() {
		NewNamespace().RegisterDirectiveExecutor(".foo", DirectiveExecutorFunc(exeFoo))
	})
}

func TestDefaultNamespace(t *testing.T) {
	RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	assert.Equal(t, LookupExecutor("foo").Execute(nil), errFoo)
	assert.PanicsWithError(t, "owl: "+duplicateExecutor("foo").Error(), func() {
		RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeFoo))
	})

	RegisterDirectiveExecutor("foo", DirectiveExecutorFunc(exeBar), true)
	assert.Equal(t, LookupExecutor("foo").Execute(nil), errBar)
}
