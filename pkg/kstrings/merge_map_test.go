package kstrings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMap(t *testing.T) {
	mapA := make(map[string]string)

	mapA["HELLO"] = "world"
	mapA["WORLD"] = "hello"

	mapB := make(map[string]string)
	mapB["HELLO"] = "planet"

	mergedMap := MergeMaps(mapA, mapB)

	assert.Equal(t, mergedMap["HELLO"], "planet")
	assert.Equal(t, mergedMap["WORLD"], "hello")
}
