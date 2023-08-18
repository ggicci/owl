package viper_test

import (
	"fmt"
	"testing"

	"github.com/ggicci/viper"
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
	Directives []*viper.Directive
}

type BuildResolverTreeTestSuite struct {
	suite.Suite
	inputValue interface{}
	expected   []*expectedResolver
	tree       *viper.Resolver
}

func NewBuildResolverTreeTestSuite(inputValue interface{}, expected []*expectedResolver) *BuildResolverTreeTestSuite {
	return &BuildResolverTreeTestSuite{
		inputValue: inputValue,
		expected:   expected,
	}
}

func (s *BuildResolverTreeTestSuite) SetupTest() {
	tree, err := viper.New(s.inputValue)
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
				Directives: []*viper.Directive{
					viper.NewDirective("form", "page"),
				},
			},
			// hidden field should be ignored
			{
				Index:      []int{1},
				LookupPath: "Size",
				NumFields:  0,
				Directives: []*viper.Directive{
					viper.NewDirective("form", "size"),
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
				Directives: []*viper.Directive{
					viper.NewDirective("form", "user"),
				},
			},
			{
				Index:      []int{0, 0},
				LookupPath: "User.Name",
				NumFields:  0,
				Directives: []*viper.Directive{
					viper.NewDirective("form", "name"),
				},
			},
			{
				Index:      []int{0, 1},
				LookupPath: "User.Gender",
				NumFields:  0,
				Directives: []*viper.Directive{
					viper.NewDirective("form", "gender"),
					viper.NewDirective("default", "unknown"),
				},
			},
			{
				Index:      []int{0, 2},
				LookupPath: "User.Birthday",
				NumFields:  0,
				Directives: []*viper.Directive{
					viper.NewDirective("form", "birthday"),
				},
			},
			{
				Index:      []int{1},
				LookupPath: "CSRFToken",
				NumFields:  0,
				Directives: []*viper.Directive{
					viper.NewDirective("form", "csrf_token"),
				},
			},
		},
	))
}

func TestResolveSimpleFlatStruct(t *testing.T) {
	assert := assert.New(t)

	tracker := &ExecutionTracker{}
	ns := viper.NewNamespace()
	ns.ReplaceDirectiveExecutor("env", NewEchoExecutor(tracker, "env"))
	ns.ReplaceDirectiveExecutor("form", NewEchoExecutor(tracker, "form"))
	ns.ReplaceDirectiveExecutor("default", NewEchoExecutor(tracker, "default"))

	type GenerateAccessTokenRequest struct {
		Key      string `viper:"env=ACCESS_TOKEN_KEY_GENERATION_KEY"`
		UserName string `viper:"form=username"`
		Expiry   int    `viper:"form=expiry;default=3600"`
	}

	resolver, err := viper.New(GenerateAccessTokenRequest{}, viper.WithNamespace(ns))
	assert.NoError(err)

	_, err = resolver.Resolve()
	assert.NoError(err)

	assert.Equal([]*viper.Directive{
		viper.NewDirective("env", "ACCESS_TOKEN_KEY_GENERATION_KEY"),
		viper.NewDirective("form", "username"),
		viper.NewDirective("form", "expiry"),
		viper.NewDirective("default", "3600"),
	}, tracker.Executed, "should execute all directives in order")
}

func TestResolveEmbeddedStruct(t *testing.T) {
	assert := assert.New(t)

	tracker := &ExecutionTracker{}
	ns := viper.NewNamespace()
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

	resolver, err := viper.New(UserListQuery{}, viper.WithNamespace(ns))
	assert.NoError(err)

	_, err = resolver.Resolve()
	assert.NoError(err)

	assert.Equal([]*viper.Directive{
		viper.NewDirective("form", "gender"),
		viper.NewDirective("form", "age", "age[]"),
		viper.NewDirective("default", "18", "999"),
		viper.NewDirective("form", "roles", "roles[]"),
		viper.NewDirective("form", "page"),
		viper.NewDirective("form", "size"),
	}, tracker.Executed, "should execute all directives in order")
}

func TestTreeDebugLayout(t *testing.T) {
	var (
		tree *viper.Resolver
		err  error
	)

	tree, err = viper.New(UserSignUpForm{})
	assert.NoError(t, err)
	fmt.Println(tree.DebugLayoutText(0))

	tree, err = viper.New(Pagination{})
	assert.NoError(t, err)
	fmt.Println(tree.DebugLayoutText(0))
}

type ExecutionTracker struct {
	Executed []*viper.Directive
}

func (et *ExecutionTracker) Track(directive *viper.Directive) {
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

func (e *EchoExecutor) Execute(ctx *viper.DirectiveRuntime) error {
	e.tracker.Track(ctx.Directive)
	fmt.Printf("Execute %q with %v\n", ctx.Directive.Name, ctx.Directive.Argv)
	return nil
}
