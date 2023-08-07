package in

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Pagination struct {
	Page   int  `in:"form=page"`
	hidden bool `in:"form=hidden"`
	Size   int  `in:"form=size"`
}

type User struct {
	Name     string `in:"form=name"`
	Gender   string `in:"form=gender;default=unknown"`
	Birthday string `in:"form=birthday"`
}

type UserSignUpForm struct {
	User      User   `in:"form=user"`
	CSRFToken string `in:"form=csrf_token"`
}

type expectedFieldResolver struct {
	Index      []int
	LookupPath string
	NumFields  int
	Directives []*Directive
}

type BuildResolverTreeTestSuite struct {
	suite.Suite
	inputValue interface{}
	expected   []*expectedFieldResolver
	tree       *Resolver
}

func NewBuildResolverTreeTestSuite(inputValue interface{}, expected []*expectedFieldResolver) *BuildResolverTreeTestSuite {
	return &BuildResolverTreeTestSuite{
		inputValue: inputValue,
		expected:   expected,
	}
}

func (s *BuildResolverTreeTestSuite) SetupTest() {
	tree, err := NewResolver(s.inputValue)
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
		[]*expectedFieldResolver{
			{
				Index:      []int{0},
				LookupPath: "Page",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "page"),
				},
			},
			{
				Index:      []int{1},
				LookupPath: "hidden",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "hidden"),
				},
			},
			{
				Index:      []int{2},
				LookupPath: "Size",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "size"),
				},
			},
		},
	))

	suite.Run(t, NewBuildResolverTreeTestSuite(
		UserSignUpForm{},
		[]*expectedFieldResolver{
			{
				Index:      []int{0},
				LookupPath: "User",
				NumFields:  3,
				Directives: []*Directive{
					NewDirective("form", "user"),
				},
			},
			{
				Index:      []int{0, 0},
				LookupPath: "User.Name",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "name"),
				},
			},
			{
				Index:      []int{0, 1},
				LookupPath: "User.Gender",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "gender"),
					NewDirective("default", "unknown"),
				},
			},
			{
				Index:      []int{0, 2},
				LookupPath: "User.Birthday",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "birthday"),
				},
			},
			{
				Index:      []int{1},
				LookupPath: "CSRFToken",
				NumFields:  0,
				Directives: []*Directive{
					NewDirective("form", "csrf_token"),
				},
			},
		},
	))
}

func TestTreeDebugLayout(t *testing.T) {
	var (
		tree *Resolver
		err  error
	)

	tree, err = NewResolver(UserSignUpForm{})
	assert.NoError(t, err)
	fmt.Println(tree.DebugLayoutText(0))

	tree, err = NewResolver(Pagination{})
	assert.NoError(t, err)
	fmt.Println(tree.DebugLayoutText(0))
}
