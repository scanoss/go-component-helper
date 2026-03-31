// SPDX-License-Identifier: MIT
/*
 * Copyright (c) 2026, SCANOSS
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package utils

import "testing"

func TestFindNearestVersion(t *testing.T) {
	testCases := []struct {
		name        string
		requirement string
		candidates  []string
		expected    string
	}{
		{
			name:        "exact match",
			requirement: "1.2.3",
			candidates:  []string{"1.0.0", "1.2.3", "2.0.0"},
			expected:    "1.2.3",
		},
		{
			name:        "nearest minor version",
			requirement: "1.3.0",
			candidates:  []string{"1.0.0", "1.2.0", "1.5.0", "2.0.0"},
			expected:    "1.2.0",
		},
		{
			name:        "nearest patch version prefers higher on tie",
			requirement: "1.2.5",
			candidates:  []string{"1.2.3", "1.2.7", "1.2.10"},
			expected:    "1.2.7",
		},
		{
			name:        "with >= operator strips operator and finds nearest",
			requirement: ">=1.2.0",
			candidates:  []string{"1.0.0", "1.1.0", "1.3.0"},
			expected:    "1.3.0",
		},
		{
			name:        "with ~ operator",
			requirement: "~2.0.0",
			candidates:  []string{"1.9.0", "2.1.0", "3.0.0"},
			expected:    "2.1.0",
		},
		{
			name:        "prefers higher version on tie",
			requirement: "1.5.0",
			candidates:  []string{"1.3.0", "1.7.0"},
			expected:    "1.7.0",
		},
		{
			name:        "invalid requirement returns empty",
			requirement: "not-a-version",
			candidates:  []string{"1.0.0", "2.0.0"},
			expected:    "",
		},
		{
			name:        "empty candidates returns empty",
			requirement: "1.0.0",
			candidates:  []string{},
			expected:    "",
		},
		{
			name:        "invalid candidate treated as v0.0.0",
			requirement: "0.0.1",
			candidates:  []string{"bad-version", "0.0.2"},
			expected:    "0.0.2",
		},
		{
			name:        "v prefix in candidates",
			requirement: "1.2.0",
			candidates:  []string{"v1.1.0", "v1.3.0", "v2.0.0"},
			expected:    "v1.3.0",
		},
		{
			name:        "major version difference weighted more",
			requirement: "2.0.0",
			candidates:  []string{"1.9.9", "3.0.0"},
			expected:    "3.0.0",
		},
		{
			name:        "whitespace in candidates",
			requirement: "1.0.0",
			candidates:  []string{" 1.0.1 ", "1.1.0"},
			expected:    "1.0.1",
		},
		{
			name:        "with ^ operator",
			requirement: "^1.2.0",
			candidates:  []string{"1.1.0", "1.3.0", "2.0.0"},
			expected:    "1.3.0",
		},
		{
			name:        "with = operator",
			requirement: "=1.2.0",
			candidates:  []string{"1.1.0", "1.2.0", "1.3.0"},
			expected:    "1.2.0",
		},
		{
			name:        "operator with space after",
			requirement: ">= 1.2.0",
			candidates:  []string{"1.1.0", "1.3.0", "2.0.0"},
			expected:    "1.3.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FindNearestVersion(tc.requirement, tc.candidates)
			if result != tc.expected {
				t.Errorf("FindNearestVersion(%q, %v) = %q, want %q",
					tc.requirement, tc.candidates, result, tc.expected)
			}
		})
	}
}

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
			name:     "caret operator",
			input:    "^1.2.3",
			expected: true,
		},
		{
			name:     "equals operator",
			input:    "=1.2.3",
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
