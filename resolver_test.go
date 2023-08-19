package owl_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Pagination struct {
	Page int `owl:"form=page"`
	Size int `owl:"form=size"`
}

type User struct {
	Name     string `owl:"form=name"`
	Gender   string `owl:"form=gender;default=unknown"`
	Birthday string `owl:"form=birthday"`
}

type UserSignUpForm struct {
	User      User   `owl:"form=user"`
	CSRFToken string `owl:"form=csrf_token"`
}

type expectedResolver struct {
	Index      int
	LookupPath string
	NumFields  int
	Directives []*owl.Directive
	Leaf       bool
}

type BuildResolverTreeTestSuite struct {
	suite.Suite
	inputValue interface{}
	expected   []*expectedResolver
	tree       *owl.Resolver
}

func NewBuildResolverTreeTestSuite(inputValue interface{}, expected []*expectedResolver) *BuildResolverTreeTestSuite {
	return &BuildResolverTreeTestSuite{
		inputValue: inputValue,
		expected:   expected,
	}
}

func (s *BuildResolverTreeTestSuite) SetupTest() {
	tree, err := owl.New(s.inputValue)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), tree)
	s.tree = tree
}

func (s *BuildResolverTreeTestSuite) Test_0_Lookup_IsLeaf() {
	assert := assert.New(s.T())
	for _, expected := range s.expected {
		resolver := s.tree.Lookup(expected.LookupPath)
		assert.Nil(s.tree.Lookup("SomeNonExistingPath"))
		assert.NotNil(resolver)
		assert.Equal(expected.Index, resolver.Index)
		assert.Equal(expected.NumFields, len(resolver.Children))
		assert.Equal(expected.Directives, resolver.Directives)
		assert.Equal(expected.Leaf, resolver.IsLeaf())
	}
}

func (s *BuildResolverTreeTestSuite) Test_1_GetDirective() {
	assert := assert.New(s.T())
	for _, expected := range s.expected {
		resolver := s.tree.Lookup(expected.LookupPath)
		for _, directive := range expected.Directives {
			assert.Equal(directive, resolver.GetDirective(directive.Name))
			assert.Nil(resolver.GetDirective("SomeNonExistingDirective"))
		}
	}
}

func TestNew_NormalCasesSuites(t *testing.T) {
	suite.Run(t, NewBuildResolverTreeTestSuite(
		Pagination{},
		[]*expectedResolver{
			{
				Index:      0,
				LookupPath: "Page",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "page"),
				},
				Leaf: true,
			},
			{
				Index:      1,
				LookupPath: "Size",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "size"),
				},
				Leaf: true,
			},
		},
	))

	suite.Run(t, NewBuildResolverTreeTestSuite(
		UserSignUpForm{},
		[]*expectedResolver{
			{
				Index:      0,
				LookupPath: "User",
				NumFields:  3,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "user"),
				},
				Leaf: false,
			},
			{
				Index:      0,
				LookupPath: "User.Name",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "name"),
				},
				Leaf: true,
			},
			{
				Index:      1,
				LookupPath: "User.Gender",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "gender"),
					owl.NewDirective("default", "unknown"),
				},
				Leaf: true,
			},
			{
				Index:      2,
				LookupPath: "User.Birthday",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "birthday"),
				},
				Leaf: true,
			},
			{
				Index:      1,
				LookupPath: "CSRFToken",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "csrf_token"),
				},
				Leaf: true,
			},
		},
	))
}

func TestNew_WithNilType(t *testing.T) {
	_, err := owl.New(nil)
	assert.ErrorContains(t, err, "nil type")
}

func TestNew_WithPointer(t *testing.T) {
	_, err := owl.New(&Pagination{})
	assert.NoError(t, err)
}

func TestNew_WithNonStruct(t *testing.T) {
	_, err := owl.New(123)
	assert.ErrorContains(t, err, "non-struct type")
}

func TestNew_ParsingDirectives_InvalidName(t *testing.T) {
	resolver, err := owl.New(struct {
		Invalid string `owl:"invalid/name"`
	}{})
	assert.Nil(t, resolver)
	assert.ErrorContains(t, err, "parse directives")
	assert.ErrorContains(t, err, "build resolver for \"Invalid\" failed:")
	assert.ErrorIs(t, err, owl.ErrInvalidDirectiveName)
}

func TestNew_ParsingDirectives_DuplicateDirectives(t *testing.T) {
	resolver, err := owl.New(struct {
		Color string `owl:"form=red;form=blue"`
	}{})
	assert.Nil(t, resolver)
	assert.ErrorContains(t, err, "parse directives")
	assert.ErrorContains(t, err, "build resolver for \"Color\" failed:")
	assert.ErrorIs(t, err, owl.ErrDuplicateDirective)
}

func TestNew_ApplyOptionsFailed(t *testing.T) {
	failOpt := owl.OptionFunc(func(r *owl.Resolver) error {
		return fmt.Errorf("apply option failed")
	})

	resolver, err := owl.New(struct{}{}, failOpt)
	assert.Nil(t, resolver)
	assert.ErrorContains(t, err, "apply option failed")
}

func TestNew_ApplyNilNamespace(t *testing.T) {
	resolver, err := owl.New(struct{}{}, owl.WithNamespace(nil))
	assert.Nil(t, resolver)
	assert.ErrorIs(t, err, owl.ErrNilNamespace)
}

func TestNew_copyCachedResolver(t *testing.T) {
	assert := assert.New(t)

	type Appearance struct {
		Color string `owl:"form=color"`
		Size  string `owl:"form=size"`
	}
	type Settings struct {
		Profile    string      `owl:"form=profile"`
		Appearance *Appearance `owl:"form=appearance"`
	}
	ns1 := owl.NewNamespace()
	ns1.RegisterDirectiveExecutor("stdin", owl.DirectiveExecutorFunc(exeNoop))
	r1, err := owl.New(Settings{}, owl.WithNamespace(ns1))
	assert.NoError(err)

	ns2 := owl.NewNamespace()
	ns2.RegisterDirectiveExecutor("form", owl.DirectiveExecutorFunc(exeNoop))
	r2, err := owl.New(Settings{}, owl.WithNamespace(ns2))
	assert.NoError(err)

	assert.Equal(ns1, r1.Namespace())
	assert.Equal(ns2, r2.Namespace())

	for _, lookupName := range []string{
		"Profile",
		"Appearance",
		"Appearance.Color",
		"Appearance.Size",
	} {
		assert.NotEqual(r1.Lookup(lookupName), r2.Lookup(lookupName))
	}

	assert.Equal(r1.Lookup("Profile").Parent, r1)
	assert.Equal(r1.Lookup("Appearance").Parent, r1)
	assert.Equal(r1.Lookup("Appearance.Color").Parent, r1.Lookup("Appearance"))
	assert.Equal(r1.Lookup("Appearance.Size").Parent, r1.Lookup("Appearance"))

	assert.Equal(r2.Lookup("Profile").Parent, r2)
	assert.Equal(r2.Lookup("Appearance").Parent, r2)
	assert.Equal(r2.Lookup("Appearance.Color").Parent, r2.Lookup("Appearance"))
	assert.Equal(r2.Lookup("Appearance.Size").Parent, r2.Lookup("Appearance"))
}

func TestResolve_SimpleFlatStruct(t *testing.T) {
	assert := assert.New(t)

	tracker := &ExecutionTracker{}
	ns := owl.NewNamespace()
	ns.ReplaceDirectiveExecutor("env", NewEchoExecutor(tracker, "env"))
	ns.ReplaceDirectiveExecutor("form", NewEchoExecutor(tracker, "form"))
	ns.ReplaceDirectiveExecutor("default", NewEchoExecutor(tracker, "default"))

	type GenerateAccessTokenRequest struct {
		Key      string `owl:"env=ACCESS_TOKEN_KEY_GENERATION_KEY"`
		UserName string `owl:"form=username"`
		Expiry   int    `owl:"form=expiry;default=3600"`
	}

	resolver, err := owl.New(GenerateAccessTokenRequest{}, owl.WithNamespace(ns))
	assert.NoError(err)

	_, err = resolver.Resolve()
	assert.NoError(err)

	assert.Equal([]*owl.Directive{
		owl.NewDirective("env", "ACCESS_TOKEN_KEY_GENERATION_KEY"),
		owl.NewDirective("form", "username"),
		owl.NewDirective("form", "expiry"),
		owl.NewDirective("default", "3600"),
	}, tracker.Executed, "should execute all directives in order")
}

func TestResolve_EmbeddedStruct(t *testing.T) {
	assert := assert.New(t)

	tracker := &ExecutionTracker{}
	ns := owl.NewNamespace()
	ns.ReplaceDirectiveExecutor("form", NewEchoExecutor(tracker, "form"))
	ns.ReplaceDirectiveExecutor("default", NewEchoExecutor(tracker, "default"))

	type UserFilter struct {
		Gender string   `owl:"form=gender"`
		Ages   []int    `owl:"form=age,age[];default=18,999"`
		Roles  []string `owl:"form=roles,roles[]"`
	}

	type Pagination struct {
		Page int `owl:"form=page"`
		Size int `owl:"form=size"`
	}

	type UserListQuery struct {
		UserFilter
		Pagination
	}

	resolver, err := owl.New(UserListQuery{}, owl.WithNamespace(ns))
	assert.NoError(err)

	_, err = resolver.Resolve()
	assert.NoError(err)

	assert.Equal([]*owl.Directive{
		owl.NewDirective("form", "gender"),
		owl.NewDirective("form", "age", "age[]"),
		owl.NewDirective("default", "18", "999"),
		owl.NewDirective("form", "roles", "roles[]"),
		owl.NewDirective("form", "page"),
		owl.NewDirective("form", "size"),
	}, tracker.Executed, "should execute all directives in order")
}

func TestResolve_UnexportedField(t *testing.T) {
	type User struct {
		Name   string `owl:"env=OWL_TEST_NAME"`
		age    int    // should be ignored
		Gender string `owl:"env=OWL_TEST_GENDER"`
	}

	ns := owl.NewNamespace()
	ns.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))
	resolver, err := owl.New(User{}, owl.WithNamespace(ns))
	assert.NotNil(t, resolver)
	assert.NoError(t, err)

	// Set environment variables.
	os.Setenv("OWL_TEST_NAME", "owl")
	os.Setenv("OWL_TEST_GENDER", "male")

	// Resolve.
	gotValue, err := resolver.Resolve()
	assert.NoError(t, err)
	assert.NotNil(t, gotValue)
	gotUser, ok := gotValue.Interface().(*User)
	assert.True(t, ok)
	assert.Equal(t, "owl", gotUser.Name)
	assert.Equal(t, "male", gotUser.Gender)
}

func TestResolve_MissingExecutor(t *testing.T) {
	ns := owl.NewNamespace()
	resolver, err := owl.New(Pagination{}, owl.WithNamespace(ns))

	assert.NotNil(t, resolver)
	assert.NoError(t, err)

	_, err = resolver.Resolve()
	assert.ErrorContains(t, err, "resolve field \"Page (int)\" failed:")
	assert.ErrorIs(t, err, owl.ErrMissingExecutor)
}

func TestResolve_DirectiveExecutionFailure(t *testing.T) {
	ns := owl.NewNamespace()
	var errExecutionFailed = errors.New("directive execution failed")
	ns.RegisterDirectiveExecutor("error", owl.DirectiveExecutorFunc(func(dr *owl.DirectiveRuntime) error {
		return errExecutionFailed
	}))

	type Request struct {
		Name string `owl:"error"`
	}

	resolver, err := owl.New(Request{}, owl.WithNamespace(ns))
	assert.NoError(t, err)

	rv, err := resolver.Resolve()
	assert.NotNil(t, rv)
	directiveExecutionError := new(owl.DirectiveExecutionError)
	assert.ErrorAs(t, err, &directiveExecutionError)
	assert.Equal(t, "error", directiveExecutionError.Directive.Name)
	assert.Len(t, directiveExecutionError.Directive.Argv, 0)
	assert.ErrorContains(t, err, "execute directive \"error\" with args [] failed:")
	assert.ErrorContains(t, err, "directive execution failed")
	assert.ErrorIs(t, err, errExecutionFailed)
}

func TestIterate(t *testing.T) {
	resolver, err := owl.New(UserSignUpForm{})
	assert.NoError(t, err)

	type contextKey int
	const ckHello contextKey = 1

	callback := func(r *owl.Resolver) error {
		r.Context = context.WithValue(r.Context, ckHello, "world")
		return nil
	}
	resolver.Iterate(callback)

	assert.Equal(t, "world", resolver.Lookup("User").Context.Value(ckHello))
	assert.Equal(t, "world", resolver.Lookup("User.Name").Context.Value(ckHello))
	assert.Equal(t, "world", resolver.Lookup("User.Gender").Context.Value(ckHello))
	assert.Equal(t, "world", resolver.Lookup("User.Birthday").Context.Value(ckHello))
	assert.Equal(t, "world", resolver.Lookup("CSRFToken").Context.Value(ckHello))
}

func TestIterate_CallbackFail(t *testing.T) {
	resolver, err := owl.New(UserSignUpForm{})
	assert.NoError(t, err)

	type contextKey int
	const ckHello contextKey = 1

	callback := func(r *owl.Resolver) error {
		if r.Field.Name == "Gender" {
			return errors.New("callback failed")
		}
		r.Context = context.WithValue(r.Context, ckHello, "world")
		return nil
	}
	err = resolver.Iterate(callback)
	assert.ErrorContains(t, err, "callback failed")

	assert.Equal(t, "world", resolver.Lookup("User").Context.Value(ckHello))
	assert.Equal(t, "world", resolver.Lookup("User.Name").Context.Value(ckHello))
	assert.Equal(t, nil, resolver.Lookup("User.Gender").Context.Value(ckHello))
	assert.Equal(t, nil, resolver.Lookup("User.Birthday").Context.Value(ckHello))
	assert.Equal(t, nil, resolver.Lookup("CSRFToken").Context.Value(ckHello))
}

func TestTreeDebugLayout(t *testing.T) {
	var (
		tree *owl.Resolver
		err  error
	)

	tree, err = owl.New(UserSignUpForm{})
	assert.NoError(t, err)
	fmt.Println(tree.DebugLayoutText(0))

	tree, err = owl.New(Pagination{})
	assert.NoError(t, err)
	fmt.Println(tree.DebugLayoutText(0))
}
