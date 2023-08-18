package owl_test

import (
	"encoding/json"
	"os"
	"strings"
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

func exeConfigLoader(rtm *owl.DirectiveRuntime) error {
	if len(rtm.Directive.Argv) == 0 {
		return nil
	}
	key := rtm.Directive.Argv[0]
	overrideKey := strings.ToUpper("MYAPP_" + strings.ReplaceAll(key, ".", "_"))

	// Override by environment variable.
	value, exists := os.LookupEnv(overrideKey)
	if exists {
		rtm.Value.Elem().SetString(value)
		return nil
	}

	// Last resort, load from config file. The file will be opened and closed
	// each time the directive is executed. Which is not a performance-wise
	// implementation. But it's just a sample here.
	configFile := rtm.Context.Value("ConfigFile").(string)
	file, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var config map[string]string
	if err := decoder.Decode(&config); err != nil {
		return err
	}
	rtm.Value.Elem().SetString(config[key])
	return nil
}

// TestOwlUseCaseEnvReader is a sample that represents how to implement an "env" directive
// and use it with owl to read environment variables.
func TestOwlUseCaseEnvReader(t *testing.T) {
	assert := assert.New(t)
	owl.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))

	type EnvConfig struct {
		Workspace string `owl:"env=OWL_HOME"`
		User      string `owl:"env=OWL_USER"`
		Debug     string `owl:"env=OWL_DEBUG"`
	}

	resolver, err := owl.New(EnvConfig{})
	assert.NoError(err)
	assert.NotNil(resolver)

	// Set environment variables.
	os.Setenv("OWL_HOME", "/home/ggicci/.owl")
	os.Setenv("OWL_USER", "owl")

	// Resolve.
	gotValue, err := resolver.Resolve()
	assert.NoError(err)
	assert.NotNil(gotValue)
	gotConfig, ok := gotValue.Interface().(*EnvConfig)
	assert.True(ok)
	assert.Equal("/home/ggicci/.owl", gotConfig.Workspace)
	assert.Equal("owl", gotConfig.User)
	assert.Equal("", gotConfig.Debug)
}

func prepareConfigFile(t *testing.T) (string, error) {
	filename := t.TempDir() + "/config.json"
	content := `{
		"workspace.root": "$HOME/.owl",
		"workspace.owner": "owl",
		"workspace.group": "owl",
		"workspace.permission": "0755",
		"admin": "owl",
		"debug": "true"
	}`
	return filename, os.WriteFile(filename, []byte(content), 0644)
}

// TestOwlUseCaseConfigLoader is a sample that represents how to implement a
// "config" directive and use it with owl to load configuration values from a
// file and also overridable by environment variables.
func TestOwlUseCaseConfigLoader(t *testing.T) {
	assert := assert.New(t)
	owl.RegisterDirectiveExecutor("config", owl.DirectiveExecutorFunc(exeConfigLoader))

	type WorkspaceConfig struct {
		Root       string `owl:"config=workspace.root"`
		Owner      string `owl:"config=workspace.owner"`
		Group      string `owl:"config=workspace.group"`
		Permission string `owl:"config=workspace.permission"`
	}

	type MyAppConfig struct {
		Workspace *WorkspaceConfig
		Admin     string `owl:"config=admin"`
		Debug     string `owl:"config=debug"`
	}

	resolver, err := owl.New(MyAppConfig{})
	assert.NoError(err)
	assert.NotNil(resolver)

	// Values only from config file.
	filename, err := prepareConfigFile(t)
	assert.NotEmpty(filename)
	assert.NoError(err)

	// Resolve.
	gotValue, err := resolver.Resolve(owl.WithValue("ConfigFile", filename))
	assert.NoError(err)
	assert.NotNil(gotValue)

	gotConfig, ok := gotValue.Interface().(*MyAppConfig)
	assert.True(ok)
	assert.Equal("$HOME/.owl", gotConfig.Workspace.Root)
	assert.Equal("owl", gotConfig.Workspace.Owner)
	assert.Equal("owl", gotConfig.Workspace.Group)
	assert.Equal("0755", gotConfig.Workspace.Permission)
	assert.Equal("owl", gotConfig.Admin)
	assert.Equal("true", gotConfig.Debug)

	// Set environment variables. Some values will be overridden.
	os.Setenv("MYAPP_WORKSPACE_ROOT", "/home/ggicci/.owl")
	os.Setenv("MYAPP_ADMIN", "ggicci")
	os.Setenv("MYAPP_DEBUG", "0")

	// Resolve again.
	gotValue, err = resolver.Resolve(owl.WithValue("ConfigFile", filename))
	assert.NoError(err)
	assert.NotNil(gotValue)

	gotConfig, ok = gotValue.Interface().(*MyAppConfig)
	assert.True(ok)
	assert.Equal("/home/ggicci/.owl", gotConfig.Workspace.Root)
	assert.Equal("ggicci", gotConfig.Admin)
	assert.Equal("0", gotConfig.Debug)
}
