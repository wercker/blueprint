package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_contains(t *testing.T) {
	tests := []struct {
		needle   string
		haystack []string
		expected bool
	}{
		{"foofoo", []string{}, false},
		{"foofoo", []string{"foofoo"}, true},
		{"foofoo", []string{"barbar", "foofoo"}, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			actual := contains(tt.needle, tt.haystack)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
