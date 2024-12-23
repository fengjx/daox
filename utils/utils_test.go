package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils_GetLength(t *testing.T) {
	s := []string{"a", "b", "c"}
	m := map[string]string{
		"a": "a",
		"b": "b",
		"c": "c",
	}
	assert.Equal(t, 3, GetLength(s))
	assert.Equal(t, 3, GetLength(m))
}
