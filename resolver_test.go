package owl_test

import (
	"fmt"
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Pagination struct {
	Page   int  `viper:"form=page"`
	hidden bool `viper:"form=hidden"` // should be ignored
	Size   int  `viper:"form=size"`
}

type User struct {
	Name     string `viper:"form=name"`
	Gender   string `viper:"form=gender;default=unknown"`
	Birthday string `viper:"form=birthday"`
}

type UserSignUpForm struct {
	User      User   `viper:"form=user"`
	CSRFToken string `viper:"form=csrf_token"`
}

type expectedResolver struct {
	Index      []int
	LookupPath string
	NumFields  int
	Directives []*owl.Directive
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

func (s *BuildResolverTreeTestSuite) Test_0_Path_Lookup() {
	assert := assert.New(s.T())
	for _, expected := range s.expected {
		fieldByIndex := s.tree.LookupByIndex(expected.Index)
		fieldByName := s.tree.Lookup(expected.LookupPath)
		assert.NotNil(fieldByIndex)
		assert.Equal(fieldByIndex, fieldByName)
		assert.Equal(expected.LookupPath, fieldByIndex.PathString())
	}
}

func (s *BuildResolverTreeTestSuite) Test_1_Directives() {
	assert := assert.New(s.T())
	for _, expected := range s.expected {
		field := s.tree.Lookup(expected.LookupPath)
		assert.Equal(expected.Directives, field.Directives)
	}
}

func TestBuildResolverTree(t *testing.T) {
	suite.Run(t, NewBuildResolverTreeTestSuite(
		Pagination{},
		[]*expectedResolver{
			{
				Index:      []int{0},
				LookupPath: "Page",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "page"),
				},
			},
			// hidden field should be ignored
			{
				Index:      []int{1},
				LookupPath: "Size",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "size"),
				},
			},
		},
	))

	suite.Run(t, NewBuildResolverTreeTestSuite(
		UserSignUpForm{},
		[]*expectedResolver{
			{
				Index:      []int{0},
				LookupPath: "User",
				NumFields:  3,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "user"),
				},
			},
			{
				Index:      []int{0, 0},
				LookupPath: "User.Name",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "name"),
				},
			},
			{
				Index:      []int{0, 1},
				LookupPath: "User.Gender",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "gender"),
					owl.NewDirective("default", "unknown"),
				},
			},
			{
				Index:      []int{0, 2},
				LookupPath: "User.Birthday",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "birthday"),
				},
			},
			{
				Index:      []int{1},
				LookupPath: "CSRFToken",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "csrf_token"),
				},
			},
		},
	))
}

func TestResolveSimpleFlatStruct(t *testing.T) {
	assert := assert.New(t)

	tracker := &ExecutionTracker{}
	ns := owl.NewNamespace()
	ns.ReplaceDirectiveExecutor("env", NewEchoExecutor(tracker, "env"))
	ns.ReplaceDirectiveExecutor("form", NewEchoExecutor(tracker, "form"))
	ns.ReplaceDirectiveExecutor("default", NewEchoExecutor(tracker, "default"))

	type GenerateAccessTokenRequest struct {
		Key      string `viper:"env=ACCESS_TOKEN_KEY_GENERATION_KEY"`
		UserName string `viper:"form=username"`
		Expiry   int    `viper:"form=expiry;default=3600"`
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

func TestResolveEmbeddedStruct(t *testing.T) {
	assert := assert.New(t)

	tracker := &ExecutionTracker{}
	ns := owl.NewNamespace()
	ns.ReplaceDirectiveExecutor("form", NewEchoExecutor(tracker, "form"))
	ns.ReplaceDirectiveExecutor("default", NewEchoExecutor(tracker, "default"))

	type UserFilter struct {
		Gender string   `viper:"form=gender"`
		Ages   []int    `viper:"form=age,age[];default=18,999"`
		Roles  []string `viper:"form=roles,roles[]"`
	}

	type Pagination struct {
		Page int `viper:"form=page"`
		Size int `viper:"form=size"`
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
