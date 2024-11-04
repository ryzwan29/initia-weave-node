package utils

import (
	"testing"
)

func TestDecodeCFEmail(t *testing.T) {
	tests := []struct {
		encoded     string
		expected    string
		expectError bool
	}{
		{
			encoded:     "fecf9ac79cc7cbcfcc98c7cccb9d98c6c6cec69bc998c9c89ac9cf9fc9c6c69ac6cccec6c7989bc9c8bec8cbd0cfcec6d0cfc7c6d0cfcfc6",
			expected:    "1d9b9512f925cf8808e7f76d71a788d82089fe76@65.108.198.118",
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.encoded, func(t *testing.T) {
			result, err := decodeCFEmail(test.encoded)
			if (err != nil) != test.expectError {
				t.Errorf("decodeCFEmail(%q) error = %v, expectError %v", test.encoded, err, test.expectError)
			}
			if result != test.expected {
				t.Errorf("decodeCFEmail(%q) = %q, want %q", test.encoded, result, test.expected)
			}
		})
	}
}
