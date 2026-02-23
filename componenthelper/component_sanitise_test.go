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

package componenthelper

import (
	"testing"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
)

func TestSanitiseComponents(t *testing.T) {
	tests := []struct {
		name           string
		components     []ComponentDTO
		expectedStatus []domain.StatusCode
	}{
		{
			name: "All valid components",
			components: []ComponentDTO{
				{Purl: "pkg:npm/lodash@4.17.21"},
				{Purl: "pkg:github/scanoss/scanoss.js@1.0.0"},
			},
			expectedStatus: []domain.StatusCode{domain.Success, domain.Success},
		},
		{
			name: "Mixed valid and invalid components",
			components: []ComponentDTO{
				{Purl: "pkg:npm/lodash@4.17.21"},
				{Purl: "invalid-purl"},
			},
			expectedStatus: []domain.StatusCode{domain.Success, domain.InvalidPurl},
		},
		{
			name: "Component with empty requirement gets extracted from purl",
			components: []ComponentDTO{
				{Purl: "pkg:npm/lodash@4.17.21", Requirement: ""},
			},
			expectedStatus: []domain.StatusCode{domain.Success},
		},
		{
			name:           "All invalid components",
			components:     []ComponentDTO{{Purl: "invalid"}, {Purl: "also-invalid"}},
			expectedStatus: []domain.StatusCode{domain.InvalidPurl, domain.InvalidPurl},
		},
		{
			name:           "Invalid Purl with semver",
			components:     []ComponentDTO{{Purl: "pkg:npm/lodash@>=4.17.21"}},
			expectedStatus: []domain.StatusCode{domain.Success},
		},
		{
			name:           "Empty purl",
			components:     []ComponentDTO{{Purl: ""}},
			expectedStatus: []domain.StatusCode{domain.InvalidPurl},
		},
		{
			name:           "Empty components",
			components:     []ComponentDTO{},
			expectedStatus: []domain.StatusCode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitiseComponents(tt.components)

			if len(result) != len(tt.expectedStatus) {
				t.Fatalf("expected %d components, got %d", len(tt.expectedStatus), len(result))
			}

			for i, expectedCode := range tt.expectedStatus {
				if result[i].Status.StatusCode != expectedCode {
					t.Errorf("component[%d]: expected status %s, got %s", i, expectedCode, result[i].Status.StatusCode)
				}
			}
		})
	}
}
