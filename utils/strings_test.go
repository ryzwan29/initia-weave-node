package utils

import (
	"testing"
)

func TestTransformFirstWordUpperCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Celestia", "CELESTIA"},
		{"Initia L1", "INITIA"},
		{"Cosmos Hub", "COSMOS"},
		{"Test123 Interwoven", "TEST123"},
		{"  extra spaces ", "EXTRA"},
		{"", ""},
		{"   ", ""},
	}

	for _, test := range tests {
		output := TransformFirstWordUpperCase(test.input)
		if output != test.expected {
			t.Errorf("For input '%s', expected '%s', but got '%s'", test.input, test.expected, output)
		}
	}
}
