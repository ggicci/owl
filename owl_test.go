package owl_test

import (
	"os"
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
)

func exeEnvReader(rtm *owl.DirectiveRuntime) error {
	if len(rtm.Directive.Argv) == 0 {
		return nil
	}
	value := os.Getenv(rtm.Directive.Argv[0])
	rtm.Value.Elem().SetString(value)
	return nil
}

func TestEnvReader(t *testing.T) {
	assert := assert.New(t)
	owl.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))

	type EnvConfig struct {
		Workspace string `owl:"env=OWL_HOME"`
		User      string `owl:"env=OWL_USER"`
		Debug     string `owl:"env=OWL_DEBUG"`
	}

	v, err := owl.New(EnvConfig{})
	assert.NoError(err)
	assert.NotNil(v)

	// Set environment variables.
	os.Setenv("OWL_HOME", "/home/ggicci/.owl")
	os.Setenv("OWL_USER", "owl")

	// Resolve.
	gotValue, err := v.Resolve()
	assert.NoError(err)
	assert.NotNil(gotValue)
	gotConfig, ok := gotValue.Interface().(*EnvConfig)
	assert.True(ok)
	assert.Equal("/home/ggicci/.owl", gotConfig.Workspace)
	assert.Equal("owl", gotConfig.User)
	assert.Equal("", gotConfig.Debug)
}
