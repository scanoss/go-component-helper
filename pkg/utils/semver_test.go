package utils

import "testing"

func TestHasSemverOperator(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "greater than operator",
			input:    ">2.3.0",
			expected: true,
		},
		{
			name:     "greater than or equal operator",
			input:    ">=1.0.0",
			expected: true,
		},
		{
			name:     "less than operator",
			input:    "<3.0.0",
			expected: true,
		},
		{
			name:     "no operator",
			input:    "1.2.3",
			expected: false,
		},
		{
			name:     "version with 'v'",
			input:    "v1.2.3",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := HasSemverOperator(tc.input)
			if result != tc.expected {
				t.Errorf("HasSemverOperator(%q) = %t, want %t",
					tc.input, result, tc.expected)
			}
		})
	}
}
