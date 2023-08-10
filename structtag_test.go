package rstruct

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUseTag(t *testing.T) {
	UseTag("custom")
	assert.Equal(t, "custom", tagName)
}
