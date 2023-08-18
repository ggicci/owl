package viper_test

import (
	"errors"
	"testing"

	"github.com/ggicci/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewDirective(t *testing.T) {
	assert := assert.New(t)

	d1 := viper.NewDirective("form")
	assert.Equal("form", d1.Name)
	assert.Len(d1.Argv, 0)

	d2 := viper.NewDirective("form", "page", "page_index")
	assert.Equal("form", d2.Name)
	assert.NotNil(d2.Argv)
	assert.Len(d2.Argv, 2)
	assert.Equal("page", d2.Argv[0])
	assert.Equal("page_index", d2.Argv[1])
}

func TestParseDirective(t *testing.T) {
	testcases := []struct {
		content  string
		expected *viper.Directive
		err      error
	}{
		{
			content:  "form",
			expected: viper.NewDirective("form"),
			err:      nil,
		},
		{
			content:  "form=page,page_index",
			expected: viper.NewDirective("form", "page", "page_index"),
			err:      nil,
		},
		{
			content:  "header=x-api-token",
			expected: viper.NewDirective("header", "x-api-token"),
			err:      nil,
		},
		{
			content:  "",
			expected: nil,
			err:      viper.ErrInvalidExecutorName,
		},
		{
			content:  "=name",
			expected: nil,
			err:      viper.ErrInvalidExecutorName,
		},
		{
			content:  "    =name",
			expected: nil,
			err:      viper.ErrInvalidExecutorName,
		},
	}

	for _, testcase := range testcases {
		directive, err := viper.ParseDirective(testcase.content)
		assert.Equal(t, testcase.expected, directive)
		assert.True(t, errors.Is(err, testcase.err))
	}
}
