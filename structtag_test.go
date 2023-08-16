package viper_test

import (
	"testing"

	"github.com/ggicci/viper"
	"github.com/stretchr/testify/assert"
)

func TestUseTag(t *testing.T) {
	viper.UseTag("custom")
	assert.Equal(t, "custom", viper.Tag())

	viper.UseTag(viper.DefaultTagName) // reset to default in case other tests fail
}
