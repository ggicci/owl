package owl_test

import (
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Package struct {
	Name        string `owl:"form=name"`
	Description string `owl:"form=description"`
	License     string `owl:"form=license"`
}

type CreatePackageRequest struct {
	Owner   string `owl:"path=owner"`
	Package `owl:"body=json"`
}

var expectedCreatePackageRequestResolverTree = []*expectedResolver{
	{
		Index:      []int{0},
		LookupPath: "Owner",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("path", "owner"),
		},
		Leaf: true,
	},
	{
		Index:      []int{1},
		LookupPath: "Package",
		NumFields:  3,
		Directives: []*owl.Directive{
			owl.NewDirective("body", "json"),
		},
		Leaf: false,
	},
	{
		Index:      []int{1, 0},
		LookupPath: "Package.Name",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "name"),
		},
		Leaf: true,
	},
	{
		Index:      []int{1, 1},
		LookupPath: "Package.Description",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "description"),
		},
		Leaf: true,
	},
	{
		Index:      []int{1, 2},
		LookupPath: "Package.License",
		NumFields:  0,
		Directives: []*owl.Directive{
			owl.NewDirective("form", "license"),
		},
		Leaf: true,
	},
}

func TestOption_WithResolveNestedDirectives(t *testing.T) {
	assert := assert.New(t)
	ns, tracker := createNsForTracking("path", "body")
	tree, err := owl.New(CreatePackageRequest{}, owl.WithNamespace(ns))
	assert.NoError(err)
	assert.NotNil(tree)

	suite.Run(t, NewBuildResolverTreeTestSuite(
		tree, expectedCreatePackageRequestResolverTree,
	))

	// Resolve nested fields by default.
	_, err = tree.Resolve()
	assert.NoError(err)
	assert.Equal([]*owl.Directive{
		owl.NewDirective("path", "owner"),
		owl.NewDirective("body", "json"),
		owl.NewDirective("form", "name"),
		owl.NewDirective("form", "description"),
		owl.NewDirective("form", "license"),
	}, tracker.Executed.ExecutedDirectives(), "should resolve nested directives by default")

	// Don't resolve nested fields by passing the WithResolveNestedDirectives(false) option.
	tracker.Reset()
	_, err = tree.Resolve(owl.WithResolveNestedDirectives(false))
	assert.NoError(err)
	assert.Equal([]*owl.Directive{
		owl.NewDirective("path", "owner"),
		owl.NewDirective("body", "json"),
	}, tracker.Executed.ExecutedDirectives(), "should not resolve nested directives")
}
