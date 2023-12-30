package owl_test

import (
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
)

func TestNewDirective(t *testing.T) {
	assert := assert.New(t)

	d1 := owl.NewDirective("form")
	assert.Equal("form", d1.Name)
	assert.Len(d1.Argv, 0)

	d2 := owl.NewDirective("form", "page", "page_index")
	assert.Equal("form", d2.Name)
	assert.NotNil(d2.Argv)
	assert.Len(d2.Argv, 2)
	assert.Equal("page", d2.Argv[0])
	assert.Equal("page_index", d2.Argv[1])
}

func TestParseDirective(t *testing.T) {
	testcases := []struct {
		content  string
		expected *owl.Directive
		err      error
	}{
		{
			content:  "form",
			expected: owl.NewDirective("form"),
			err:      nil,
		},
		{
			content:  "form=page,page_index",
			expected: owl.NewDirective("form", "page", "page_index"),
			err:      nil,
		},
		{
			content:  "header=x-api-token",
			expected: owl.NewDirective("header", "x-api-token"),
			err:      nil,
		},
		{
			content:  "",
			expected: nil,
			err:      owl.ErrInvalidDirectiveName,
		},
		{
			content:  "=name",
			expected: nil,
			err:      owl.ErrInvalidDirectiveName,
		},
		{
			content:  "    =name",
			expected: nil,
			err:      owl.ErrInvalidDirectiveName,
		},
	}

	for _, testcase := range testcases {
		directive, err := owl.ParseDirective(testcase.content)
		assert.Equal(t, testcase.expected, directive)
		assert.ErrorIs(t, err, testcase.err)
	}
}

func TestDirective_String(t *testing.T) {
	d := owl.NewDirective("form", "page", "page_index")
	assert.Equal(t, "form=page,page_index", d.String())

	d = owl.NewDirective("required")
	assert.Equal(t, "required", d.String())
}
