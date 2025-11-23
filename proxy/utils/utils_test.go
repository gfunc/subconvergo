package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlEncode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test", "test"},
		{"test space", "test%20space"},
		{"test#hash", "test%23hash"},
		{"test space#hash", "test%20space%23hash"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, UrlEncode(tt.input))
		})
	}
}
