package owl_test

import (
	"testing"

	"github.com/ggicci/owl"
	"github.com/stretchr/testify/assert"
)

func TestUseTag(t *testing.T) {
	owl.UseTag("custom")
	assert.Equal(t, "custom", owl.Tag())

	owl.UseTag(owl.DefaultTagName) // reset to default in case other tests fail
}
