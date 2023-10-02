package owl_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ggicci/owl"
)

func exeNoop(rtm *owl.DirectiveRuntime) error {
	return nil
}

func exeEnvReader(rtm *owl.DirectiveRuntime) error {
	if len(rtm.Directive.Argv) == 0 {
		return nil
	}
	if value, ok := os.LookupEnv(rtm.Directive.Argv[0]); ok {
		rtm.Value.Elem().SetString(value)
	}
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

type ExecutedData struct {
	*owl.Directive
	FieldValue any
}

type ExecutedDataList []ExecutedData

func (edl ExecutedDataList) ExecutedDirectives() []*owl.Directive {
	dirs := make([]*owl.Directive, len(edl))
	for i, d := range edl {
		dirs[i] = d.Directive
	}
	return dirs
}

type ExecutionTracker struct {
	Executed ExecutedDataList
}

func NewExecutionTracker() *ExecutionTracker {
	return &ExecutionTracker{}
}

func (et *ExecutionTracker) Track(directive *owl.Directive, value any) {
	et.Executed = append(et.Executed, ExecutedData{directive, value})
}

func (et *ExecutionTracker) Reset() {
	et.Executed = nil
}

type ContextVerifier struct {
	Key      interface{}
	Expected interface{}
}

func (cv *ContextVerifier) Verify(ctx context.Context) error {
	if cv.Expected != ctx.Value(cv.Key) {
		return fmt.Errorf("unexpected context value for key %q: %v", cv.Key, ctx.Value(cv.Key))
	}
	return nil
}

type EchoExecutor struct {
	Name            string
	ThrowError      error
	ContextVerifier *ContextVerifier

	tracker *ExecutionTracker
}

func NewEchoExecutor(tracker *ExecutionTracker, name string, throwError error, verifier *ContextVerifier) *EchoExecutor {
	return &EchoExecutor{
		Name:            name,
		ThrowError:      throwError,
		ContextVerifier: verifier,
		tracker:         tracker,
	}
}

func (e *EchoExecutor) Execute(ctx *owl.DirectiveRuntime) error {
	e.tracker.Track(ctx.Directive, ctx.Value.Interface())
	fmt.Printf("Execute %q with args: %v, value: %v\n", ctx.Directive.Name, ctx.Directive.Argv, ctx.Value)
	if e.ContextVerifier != nil {
		if err := e.ContextVerifier.Verify(ctx.Context); err != nil {
			return err
		}
	}
	return e.ThrowError
}

func createNsForTracking() (*owl.Namespace, *ExecutionTracker) {
	return createNsForTrackingCtor(nil, nil)
}

func createNsForTrackingWithError(throwError error) (*owl.Namespace, *ExecutionTracker) {
	return createNsForTrackingCtor(throwError, nil)
}

func createNsForTrackingWithContextVerifier(verifier *ContextVerifier) (*owl.Namespace, *ExecutionTracker) {
	return createNsForTrackingCtor(nil, verifier)
}

func createNsForTrackingCtor(throwError error, verifier *ContextVerifier) (*owl.Namespace, *ExecutionTracker) {
	tracker := NewExecutionTracker()
	ns := owl.NewNamespace()
	ns.RegisterDirectiveExecutor("form", NewEchoExecutor(tracker, "form", throwError, verifier))
	ns.RegisterDirectiveExecutor("default", NewEchoExecutor(tracker, "default", throwError, verifier))
	ns.RegisterDirectiveExecutor("env", NewEchoExecutor(tracker, "env", throwError, verifier))
	return ns, tracker
}
