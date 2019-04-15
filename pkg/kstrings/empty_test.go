package kstrings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyString(t *testing.T) {
	empty := ""
	var str string

	assert.True(t, IsEmpty(&str))
	assert.True(t, IsEmpty(nil))
	assert.True(t, IsEmpty(&empty))
	str = "Hello"
	assert.False(t, IsEmpty(&str))
}
