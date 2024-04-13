package owl_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
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

var expectedUserSignUpFormResolverTree = []*expectedResolver{
	{
		Index:      []int{0},
		LookupPath: "User",
		NumFields:  3,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "user"),
		},
		Leaf: false,
	},
	{
		Index:      []int{0, 0},
		LookupPath: "User.Name",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "name"),
		},
		Leaf: true,
	},
	{
		Index:      []int{0, 1},
		LookupPath: "User.Gender",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "gender"),
			owl.NewDirective("default", "unknown"),
		},
		Leaf: true,
	},
	{
		Index:      []int{0, 2},
		LookupPath: "User.Birthday",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "birthday"),
		},
		Leaf: true,
	},
	{
		Index:      []int{1},
		LookupPath: "CSRFToken",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "csrf_token"),
		},
		Leaf: true,
	},
}

type expectedResolver struct {
	Index      []int
	LookupPath string
	NumFields  int
	Directives []*owl.Directive
	Leaf       bool
}

type BuildResolverTreeTestSuite struct {
	suite.Suite
	expected []*expectedResolver
	tree     *owl.Resolver
}

func NewBuildResolverTreeTestSuite(tree *owl.Resolver, expected []*expectedResolver) *BuildResolverTreeTestSuite {
	return &BuildResolverTreeTestSuite{
		expected: expected,
		tree:     tree,
	}
}

func (s *BuildResolverTreeTestSuite) Test_0_Lookup_IsLeaf() {
	assert := assert.New(s.T())
	for _, expected := range s.expected {
		resolver := s.tree.Lookup(expected.LookupPath)
		assert.NotNil(resolver)
		assert.Equal(expected.Index, resolver.Index)
		assert.Equal(expected.NumFields, len(resolver.Children))
		assert.Equal(expected.Directives, resolver.Directives)
		assert.Equal(expected.Leaf, resolver.IsLeaf())
	}

	assert.Nil(s.tree.Lookup("SomeNonExistingPath"))
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
	tree, err := owl.New(Pagination{})
	assert.NoError(t, err)
	assert.NotNil(t, tree)

	suite.Run(t, NewBuildResolverTreeTestSuite(
		tree,
		[]*expectedResolver{
			{
				Index:      []int{0},
				LookupPath: "Page",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "page"),
				},
				Leaf: true,
			},
			{
				Index:      []int{1},
				LookupPath: "Size",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "size"),
				},
				Leaf: true,
			},
		},
	))

	tree, err = owl.New(UserSignUpForm{})
	assert.NoError(t, err)
	assert.NotNil(t, tree)
	suite.Run(t, NewBuildResolverTreeTestSuite(
		tree,
		expectedUserSignUpFormResolverTree,
	))
}

func TestNew_SkipFieldsHavingNoDirectives(t *testing.T) {
	type AnotherForm struct {
		Username   string      `owl:"form=username"`
		Password   string      `owl:"form=password"`
		Hidden     string      // should be ignored
		Pagination *Pagination // should not be ignored
	}

	tree, err := owl.New(AnotherForm{})
	assert.NotNil(t, tree)
	assert.NoError(t, err)

	suite.Run(t, NewBuildResolverTreeTestSuite(
		tree,
		[]*expectedResolver{
			{
				Index:      []int{},
				LookupPath: "",
				NumFields:  3,
				Directives: []*owl.Directive{},
				Leaf:       false,
			},
			{
				Index:      []int{0},
				LookupPath: "Username",
				Directives: []*owl.Directive{
					owl.NewDirective("form", "username"),
				},
				Leaf: true,
			},
			{
				Index:      []int{1},
				LookupPath: "Password",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "password"),
				},
				Leaf: true,
			},
			{
				Index:      []int{3}, // Pagination is the 4th field, Hidden is the 3rd field.
				LookupPath: "Pagination",
				NumFields:  2,
				Directives: []*owl.Directive{},
				Leaf:       false,
			},
			{
				Index:      []int{3, 0},
				LookupPath: "Pagination.Page",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "page"),
				},
				Leaf: true,
			},
			{
				Index:      []int{3, 1},
				LookupPath: "Pagination.Size",
				NumFields:  0,
				Directives: []*owl.Directive{
					owl.NewDirective("form", "size"),
				},
				Leaf: true,
			},
		},
	))
}

func TestNew_WithNilType(t *testing.T) {
	_, err := owl.New(nil)
	assert.ErrorIs(t, err, owl.ErrUnsupportedType)
	assert.ErrorContains(t, err, "nil type")
}

func TestNew_WithPointer(t *testing.T) {
	_, err := owl.New(&Pagination{})
	assert.NoError(t, err)
}

func TestNew_WithNonStruct(t *testing.T) {
	_, err := owl.New(123)
	assert.ErrorIs(t, err, owl.ErrUnsupportedType)
	assert.ErrorContains(t, err, "non-struct type")
}

func TestNew_WithRecursion(t *testing.T) {
	type RecursiveRef struct {
		Loop *RecursiveRef
	}
	_, err := owl.New(RecursiveRef{})
	assert.NoError(t, err)
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

func TestNew_OptionNilNamespace(t *testing.T) {
	resolver, err := owl.New(struct{}{}, owl.WithNamespace(nil))
	assert.Nil(t, resolver)
	assert.ErrorContains(t, err, "nil namespace")
}

func TestNew_OptionCustomValue(t *testing.T) {
	resolver, err := owl.New(Pagination{}, owl.WithValue("hello", "world"))
	assert.NotNil(t, resolver)
	assert.NoError(t, err)

	resolver.Iterate(func(r *owl.Resolver) error {
		assert.Equal(t, "world", r.Context.Value("hello"))
		return nil
	})
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

func TestRemoveDirective(t *testing.T) {
	type User struct {
		Name string `owl:"form=name;query=name;header=X-Name;required"`
	}

	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	form := resolver.RemoveDirective("form")
	assert.Nil(t, form)

	nameResolver := resolver.Lookup("Name")
	assert.NotNil(t, nameResolver)

	form = nameResolver.RemoveDirective("form")
	assert.NotNil(t, form)
	assert.Equal(t, "form", form.Name)
	assert.Equal(t, []string{"name"}, form.Argv)
	assert.Nil(t, nameResolver.GetDirective("form"))

	required := nameResolver.RemoveDirective("required")
	assert.NotNil(t, required)
	assert.Equal(t, "required", required.Name)
	assert.Len(t, required.Argv, 0)
	assert.Nil(t, nameResolver.GetDirective("required"))
}

func TestResolve_SimpleFlatStruct(t *testing.T) {
	assert := assert.New(t)
	ns, tracker := createNsForTracking()
	type GenerateAccessTokenRequest struct {
		Key      string `owl:"env=ACCESS_TOKEN_KEY_GENERATION_KEY"`
		UserName string `owl:"form=username"`
		Expiry   int    `owl:"form=expiry;default=3600"`
	}

	resolver, err := owl.New(GenerateAccessTokenRequest{}, owl.WithNamespace(ns))
	assert.NoError(err)
	expectedExecutedDirectives := []*owl.Directive{
		owl.NewDirective("env", "ACCESS_TOKEN_KEY_GENERATION_KEY"),
		owl.NewDirective("form", "username"),
		owl.NewDirective("form", "expiry"),
		owl.NewDirective("default", "3600"),
	}

	// Resolve
	_, err = resolver.Resolve()
	assert.NoError(err)
	assert.Equal(expectedExecutedDirectives, tracker.Executed.ExecutedDirectives(), "should execute all directives in order")

	// ResolveTo
	tracker.Reset()
	var targetValue = new(GenerateAccessTokenRequest)
	err = resolver.ResolveTo(targetValue)
	assert.NoError(err)
	assert.Equal(expectedExecutedDirectives, tracker.Executed.ExecutedDirectives(), "should execute all directives in order")
}

func TestResolve_EmbeddedStruct(t *testing.T) {
	assert := assert.New(t)
	ns, tracker := createNsForTracking()
	type UserFilter struct {
		Gender string   `owl:"form=gender"`
		Ages   []int    `owl:"form=age,age[];default=18,999"`
		Roles  []string `owl:"form=roles,roles[]"`
	}

	type UserListQuery struct {
		UserFilter
		Pagination
	}

	resolver, err := owl.New(UserListQuery{}, owl.WithNamespace(ns))
	assert.NoError(err)
	expectedExecutedDirectives := []*owl.Directive{
		owl.NewDirective("form", "gender"),
		owl.NewDirective("form", "age", "age[]"),
		owl.NewDirective("default", "18", "999"),
		owl.NewDirective("form", "roles", "roles[]"),
		owl.NewDirective("form", "page"),
		owl.NewDirective("form", "size"),
	}

	// Resolve
	_, err = resolver.Resolve()
	assert.NoError(err)
	assert.Equal(expectedExecutedDirectives, tracker.Executed.ExecutedDirectives(), "should execute all directives in order")

	// ResolveTo
	tracker.Reset()
	var targetValue = new(UserListQuery)
	err = resolver.ResolveTo(targetValue)
	assert.NoError(err)
	assert.Equal(expectedExecutedDirectives, tracker.Executed.ExecutedDirectives(), "should execute all directives in order")
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
	gotUser, ok := gotValue.Interface().(*User)
	assert.True(t, ok)
	assert.Equal(t, "owl", gotUser.Name)
	assert.Equal(t, "male", gotUser.Gender)

	os.Clearenv()
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
	assert := assert.New(t)
	ns := owl.NewNamespace()
	var errExecutionFailed = errors.New("directive execution failed")
	ns.RegisterDirectiveExecutor("error", owl.DirectiveExecutorFunc(func(dr *owl.DirectiveRuntime) error {
		return errExecutionFailed
	}))

	type Request struct {
		Name string `owl:"error"`
	}

	resolver, err := owl.New(Request{}, owl.WithNamespace(ns))
	assert.NoError(err)

	rv, err := resolver.Resolve()
	assert.NotNil(rv)

	resolveError := new(owl.ResolveError)
	assert.ErrorAs(err, &resolveError)
	assert.Equal("Name", resolveError.Resolver.Field.Name)

	directiveExecutionError := new(owl.DirectiveExecutionError)
	assert.ErrorAs(err, &directiveExecutionError)
	assert.Equal(directiveExecutionError, resolveError.AsDirectiveExecutionError())
	assert.Equal("error", directiveExecutionError.Directive.Name)
	assert.Len(directiveExecutionError.Directive.Argv, 0)
	assert.ErrorContains(err, "execute directive \"error\" with args [] failed:")
	assert.ErrorContains(err, "directive execution failed")
	assert.ErrorIs(err, errExecutionFailed)
}

func TestResolve_DirectiveRuntimeContext(t *testing.T) {
	type contextKey int
	const ckSet contextKey = 1
	exeSetField := func(dr *owl.DirectiveRuntime) error {
		dr.Context = context.WithValue(dr.Context, ckSet, true)
		return nil
	}
	exeUnsetField := func(dr *owl.DirectiveRuntime) error {
		dr.Context = context.WithValue(dr.Context, ckSet, false)
		return nil
	}
	exeRequired := func(dr *owl.DirectiveRuntime) error {
		alreadySet := dr.Context.Value(ckSet)
		if alreadySet == nil || !alreadySet.(bool) {
			return errors.New("field is required")
		}
		return nil
	}

	ns := owl.NewNamespace()
	ns.RegisterDirectiveExecutor("set", owl.DirectiveExecutorFunc(exeSetField))
	ns.RegisterDirectiveExecutor("unset", owl.DirectiveExecutorFunc(exeUnsetField))
	ns.RegisterDirectiveExecutor("required", owl.DirectiveExecutorFunc(exeRequired))

	type RequestSet struct {
		// In required directive, the context value of ckSet should be true here.
		Name string `owl:"set;required"`
	}

	resolver, err := owl.New(RequestSet{}, owl.WithNamespace(ns))
	assert.NoError(t, err)
	_, err = resolver.Resolve()
	assert.NoError(t, err)

	type RequestUnset struct {
		// In required directive, the context value of ckSet should be false here.
		// Thus, the required directive should fail.
		Name string `owl:"unset;required"`
	}

	resolver, err = owl.New(RequestUnset{}, owl.WithNamespace(ns))
	assert.NoError(t, err)
	_, err = resolver.Resolve()
	assert.ErrorContains(t, err, "field is required")
}

func TestResolve_NestedDirectives(t *testing.T) {
	type User struct {
		Name string `owl:"env=OWL_TEST_NAME"`
		Role string `owl:"env=OWL_TEST_ROLE"`
	}

	type Request struct {
		// NOTE: Login will be created and updated by login directive,
		// and its fields will also be updated by env directive.
		Login  User   `owl:"login"`
		Action string `owl:"env=OWL_TEST_ACTION"`
	}

	expected := &Request{
		Login: User{
			Name: "owl",   // set by env
			Role: "admin", // set by login
		},
		Action: "addAccount",
	}

	ns := owl.NewNamespace()
	ns.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))
	ns.RegisterDirectiveExecutor("login", owl.DirectiveExecutorFunc(func(dr *owl.DirectiveRuntime) error {
		u := User{Name: "hello", Role: "admin"}
		dr.Value.Elem().Set(reflect.ValueOf(u))
		return nil
	}))
	os.Setenv("OWL_TEST_NAME", "owl")
	os.Setenv("OWL_TEST_ACTION", "addAccount")

	resolver, err := owl.New(Request{}, owl.WithNamespace(ns))
	assert.NoError(t, err)
	gotValue, err := resolver.Resolve()
	assert.NoError(t, err)
	assert.Equal(t, expected, gotValue.Interface().(*Request))

	os.Clearenv()
}

func TestResolve_WithNestedDirectivesEnabled_false(t *testing.T) {
	assert := assert.New(t)
	ns, tracker := createNsForTracking()
	tree, err := owl.New(
		UserSignUpForm{},
		owl.WithNamespace(ns),
		owl.WithNestedDirectivesEnabled(false), // disable resolving nested directives
	)
	assert.NoError(err)
	assert.NotNil(tree)

	suite.Run(t, NewBuildResolverTreeTestSuite(
		tree, expectedUserSignUpFormResolverTree,
	))

	// Won't resolve nested directives because WithNestedDirectivesEnabled(false).
	_, err = tree.Resolve()
	assert.NoError(err)
	assert.Equal([]*owl.Directive{
		owl.NewDirective("form", "user"),
		owl.NewDirective("form", "csrf_token"),
	}, tracker.Executed.ExecutedDirectives(), "should not resolve nested directives")

	// The value set in New will be overridden by the value set in Resolve or Scan.
	tracker.Reset()
	_, err = tree.Resolve(owl.WithNestedDirectivesEnabled(true)) // override
	assert.NoError(err)
	assert.Equal([]*owl.Directive{
		owl.NewDirective("form", "user"),
		// User nested directives.
		owl.NewDirective("form", "name"),
		owl.NewDirective("form", "gender"),
		owl.NewDirective("default", "unknown"),
		owl.NewDirective("form", "birthday"),

		owl.NewDirective("form", "csrf_token"),
	}, tracker.Executed.ExecutedDirectives(), "should resolve nested directives")
}

func TestResolveTo_InstantializeOnlyNilPointerForNestedStruct(t *testing.T) {
	type Owner struct {
		Type string `owl:"env=type"`
		Name string `owl:"env=name"`
	}

	type AddOwnershipRequest struct {
		ResourceId string `owl:"env=resource_id"`
		Owner      *Owner
	}

	ns := owl.NewNamespace()
	ns.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))
	resolver, err := owl.New(AddOwnershipRequest{}, owl.WithNamespace(ns))
	assert.NoError(t, err)

	os.Setenv("type", "usergroup")
	os.Setenv("name", "admin")
	os.Setenv("resource_id", "123")

	useOwner := &Owner{}
	reqWithOwnerInstantiated := &AddOwnershipRequest{
		ResourceId: "",
		Owner:      useOwner,
	}
	err = resolver.ResolveTo(reqWithOwnerInstantiated)
	assert.NoError(t, err)

	// The Owner field is already instantiated, so we only populate the fields,
	// but not create a new instance and assign it to the Owner field.
	assert.Same(t, useOwner, reqWithOwnerInstantiated.Owner)
	assert.Equal(t, "usergroup", reqWithOwnerInstantiated.Owner.Type)
	assert.Equal(t, "admin", reqWithOwnerInstantiated.Owner.Name)
	assert.Equal(t, "123", reqWithOwnerInstantiated.ResourceId)

	// The Owner field is nil, so we create a new instance when resolving.
	reqWithOwnerNotInstantiated := &AddOwnershipRequest{
		ResourceId: "",
		Owner:      nil,
	}
	err = resolver.ResolveTo(reqWithOwnerNotInstantiated)
	assert.NoError(t, err)
	assert.Equal(t, &Owner{Type: "usergroup", Name: "admin"}, reqWithOwnerNotInstantiated.Owner)
	assert.Equal(t, "123", reqWithOwnerNotInstantiated.ResourceId)

	os.Clearenv()
}

func TestResolveTo_PopulateFieldsOnDemand(t *testing.T) {
	type User struct {
		Name string `owl:"env=OWL_TEST_NAME"`
	}

	ns := owl.NewNamespace()
	ns.RegisterDirectiveExecutor("env", owl.DirectiveExecutorFunc(exeEnvReader))
	resolver, err := owl.New(User{}, owl.WithNamespace(ns))
	assert.NoError(t, err)

	user := &User{Name: "admin"}
	err = resolver.ResolveTo(user)
	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Name) // not changed

	os.Setenv("OWL_TEST_NAME", "owl")
	err = resolver.ResolveTo(user)
	assert.NoError(t, err)
	assert.Equal(t, "owl", user.Name) // changed
	os.Clearenv()
}

func TestResolveTo_ErrNilValue(t *testing.T) {
	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	err = resolver.ResolveTo(nil)
	assert.ErrorContains(t, err, "nil")

	err = resolver.ResolveTo((*User)(nil))
	assert.ErrorContains(t, err, "nil pointer")
}

func TestResolveTo_ErrNonPointerValue(t *testing.T) {
	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	var user User
	err = resolver.ResolveTo(user)
	assert.ErrorContains(t, err, "non-pointer")
}

func TestResolveTo_ErrTypeMismatch(t *testing.T) {
	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	var user = new(Pagination)
	err = resolver.ResolveTo(user)
	assert.ErrorIs(t, err, owl.ErrTypeMismatch)
}

func TestScan(t *testing.T) {
	ns, tracker := createNsForTracking()
	resolver, err := owl.New(User{}, owl.WithNamespace(ns))
	assert.NoError(t, err)

	user := &User{
		Name:     "Ggicci",
		Gender:   "male",
		Birthday: "1991-11-10",
	}
	expected := ExecutedDataList{
		{owl.NewDirective("form", "name"), "Ggicci"},
		{owl.NewDirective("form", "gender"), "male"},
		{owl.NewDirective("default", "unknown"), "male"},
		{owl.NewDirective("form", "birthday"), "1991-11-10"},
	}

	// scan on pointer value
	err = resolver.Scan(user)
	assert.NoError(t, err)
	assert.Equal(t, expected, tracker.Executed)

	// scan on non-pointer value
	tracker.Reset()
	err = resolver.Scan(*user)
	assert.NoError(t, err)
	assert.Equal(t, expected, tracker.Executed)
}

func TestScan_withOpts(t *testing.T) {
	ns, _ := createNsForTrackingWithContextVerifier(&ContextVerifier{"hello", "world"})
	resolver, err := owl.New(User{}, owl.WithNamespace(ns))
	assert.NoError(t, err)
	err = resolver.Scan(User{}, owl.WithValue("hello", "world"))
	assert.NoError(t, err)

	err = resolver.Scan(User{}, owl.WithValue("hello", "golang"))
	assert.ErrorContains(t, err, "unexpected context value")
}

func TestScan_overrideNamespace(t *testing.T) {
	resolver, err := owl.New(User{}) // using default namespace
	assert.NoError(t, err)
	err = resolver.Scan(User{})
	assert.Error(t, err) // should fail because default namespace has no directive executors

	ns, _ := createNsForTracking()
	err = resolver.Scan(User{}, owl.WithNamespace(ns))
	assert.NoError(t, err) // should success because namespace is overrided
}

func TestScan_onNilValue(t *testing.T) {
	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	err = resolver.Scan(nil)
	assert.ErrorContains(t, err, "nil")
}

func TestScan_ErrTypeMismatch(t *testing.T) {
	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	err = resolver.Scan(Pagination{})
	assert.ErrorIs(t, err, owl.ErrTypeMismatch)
}

func TestScan_ErrMissingExecutor(t *testing.T) {
	resolver, err := owl.New(User{})
	assert.NoError(t, err)

	err = resolver.Scan(User{})
	assert.ErrorIs(t, err, owl.ErrMissingExecutor)
	assert.Len(t, err.(interface{ Unwrap() []error }).Unwrap(), 3)
}

func TestScan_NestedDirectives(t *testing.T) {
	ns, tracker := createNsForTracking()
	resolver, err := owl.New(UserSignUpForm{}, owl.WithNamespace(ns))
	assert.NoError(t, err)

	form := &UserSignUpForm{
		User: User{
			Name:     "Ggicci",
			Gender:   "male",
			Birthday: "1991-11-10",
		},
		CSRFToken: "123456",
	}

	expected := ExecutedDataList{
		{owl.NewDirective("form", "user"), form.User}, // nested struct

		{owl.NewDirective("form", "name"), "Ggicci"},
		{owl.NewDirective("form", "gender"), "male"},
		{owl.NewDirective("default", "unknown"), "male"},
		{owl.NewDirective("form", "birthday"), "1991-11-10"},

		{owl.NewDirective("form", "csrf_token"), "123456"},
	}

	err = resolver.Scan(form)
	assert.NoError(t, err)
	assert.Equal(t, expected, tracker.Executed)
}

func TestScan_WithNestedDirectivesEnabled_false(t *testing.T) {
	ns, tracker := createNsForTracking()
	resolver, err := owl.New(
		UserSignUpForm{},
		owl.WithNamespace(ns),
		owl.WithNestedDirectivesEnabled(false), // disable resolving nested directives
	)
	assert.NoError(t, err)

	form := &UserSignUpForm{
		User: User{
			Name:     "Ggicci",
			Gender:   "male",
			Birthday: "1991-11-10",
		},
		CSRFToken: "123456",
	}

	expected := ExecutedDataList{
		{owl.NewDirective("form", "user"), form.User},
		// Nested directives are not resolved:
		// {owl.NewDirective("form", "name"), "Ggicci"},
		// ...
		{owl.NewDirective("form", "csrf_token"), "123456"},
	}

	err = resolver.Scan(form)
	assert.NoError(t, err)
	assert.Equal(t, expected, tracker.Executed)
}

func TestScan_NestedDirectives_ScanErrors_executeDirectiveFailed(t *testing.T) {
	ns, _ := createNsForTrackingWithError(errors.New("TestScan_NestedDirectives_ScanErrors"))
	resolver, err := owl.New(UserSignUpForm{}, owl.WithNamespace(ns))
	assert.NoError(t, err)

	form := &UserSignUpForm{
		User: User{
			Name:     "Ggicci",
			Gender:   "male",
			Birthday: "1991-11-10",
		},
		CSRFToken: "123456",
	}

	err = resolver.Scan(form)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "TestScan_NestedDirectives_ScanErrors")
	assert.Len(t, err.(interface{ Unwrap() []error }).Unwrap(), 5)
}

func TestScan_NestedDirectives_ScanErrors_ErrScanNilField(t *testing.T) {
	type MyUserSignUpForm struct {
		User  *User
		Token string
	}
	resolver, err := owl.New(MyUserSignUpForm{})
	assert.NoError(t, err)

	form := &MyUserSignUpForm{User: nil, Token: "123456"}
	err = resolver.Scan(form)
	assert.ErrorIs(t, err, owl.ErrScanNilField)
	assert.Len(t, err.(interface{ Unwrap() []error }).Unwrap(), 3) // User has 3 fields that defined owl directives.
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

func TestWithNestedDirectivesEnabled_definitionOfNestedDirectives(t *testing.T) {
	type CreateUserRequest struct {
		ApiVersion string `owl:"form=api_version"`
		User
	}

	ns, tracker := createNsForTracking()
	resolver, err := owl.New(CreateUserRequest{}, owl.WithNamespace(ns))
	assert.NoError(t, err)
	resolver.Resolve(owl.WithNestedDirectivesEnabled(false))

	assert.Equal(t, []*owl.Directive{
		owl.NewDirective("form", "api_version"),

		// Below are NOT nested directives. Because the field User has no directives.
		// If we added any directive to User, then the directives below will be nested directives.
		// Ex: User `owl:"form=user"`
		owl.NewDirective("form", "name"),
		owl.NewDirective("form", "gender"),
		owl.NewDirective("default", "unknown"),
		owl.NewDirective("form", "birthday"),
	}, tracker.Executed.ExecutedDirectives(), "tell the difference between nested and non-nested directives")
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
