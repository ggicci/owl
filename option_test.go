package owl_test

import (
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestOption_WithResolveNestedDirectives_false(t *testing.T) {
	type Package struct {
		Name        string `owl:"form=name"`
		Description string `owl:"form=description"`
		License     string `owl:"form=license"`
	}

	type CreatePackageRequest struct {
		Owner   string `owl:"path=owner"`
		Package `owl:"body=json"`
	}

	assert := assert.New(t)
	ns, tracker := createNsForTracking("path", "body")
	tree, err := owl.New(
		CreatePackageRequest{},
		owl.WithNamespace(ns),
		owl.WithResolveNestedDirectives(false),
	)
	assert.NoError(err)
	assert.NotNil(tree)

	suite.Run(t, NewBuildResolverTreeTestSuite(
		tree,
		[]*expectedResolver{
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
		},
	))

	_, err = tree.Resolve()
	assert.NoError(err)

	// NOTE: the nested directives of the "Package" field should not be resolved.
	assert.Equal([]*owl.Directive{
		owl.NewDirective("path", "owner"),
		owl.NewDirective("body", "json"),
	}, tracker.Executed.ExecutedDirectives(), "should not resolve nested directives")
}
