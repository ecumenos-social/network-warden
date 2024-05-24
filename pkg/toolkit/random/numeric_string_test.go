package random_test

import (
	"testing"

	"github.com/ecumenos-social/network-warden/pkg/toolkit/primitives"
	"github.com/ecumenos-social/network-warden/pkg/toolkit/random"
	"github.com/stretchr/testify/assert"
)

func TestGenNumericString(t *testing.T) {
	length := 10
	n := length * 5_000
	generated := make([]string, 0, n)
	for i := 0; i < n; i++ {
		g := random.GenNumericString(length)
		assert.Equal(t, length, len(g))
		generated = append(generated, g)
	}
	assert.Equal(t, n, len(primitives.Unique(generated)))
}
