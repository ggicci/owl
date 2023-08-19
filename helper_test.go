package owl_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ggicci/owl"
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

type ExecutionTracker struct {
	Executed []*owl.Directive
}

func (et *ExecutionTracker) Track(directive *owl.Directive) {
	et.Executed = append(et.Executed, directive)
}

type EchoExecutor struct {
	Name string

	tracker *ExecutionTracker
}

func NewEchoExecutor(tracker *ExecutionTracker, name string) *EchoExecutor {
	return &EchoExecutor{
		Name:    name,
		tracker: tracker,
	}
}

func (e *EchoExecutor) Execute(ctx *owl.DirectiveRuntime) error {
	e.tracker.Track(ctx.Directive)
	fmt.Printf("Execute %q with %v\n", ctx.Directive.Name, ctx.Directive.Argv)
	return nil
}
