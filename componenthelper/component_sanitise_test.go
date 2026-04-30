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
	"context"
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func TestSanitiseComponents(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()
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
			result := sanitiseComponents(s, tt.components)

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

func TestSanitiseComponentsOriginalFields(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	tests := []struct {
		name                 string
		input                ComponentDTO
		expectedOriginalPurl string
		expectedOriginalReq  string
		expectedPurl         string
		expectedRequirement  string
	}{
		{
			name:                 "Version in purl is extracted and original purl preserved",
			input:                ComponentDTO{Purl: "pkg:npm/lodash@4.17.21", Requirement: ""},
			expectedOriginalPurl: "pkg:npm/lodash@4.17.21",
			expectedOriginalReq:  "",
			expectedPurl:         "pkg:npm/lodash",
			expectedRequirement:  "4.17.21",
		},
		{
			name:                 "Requirement provided without version in purl",
			input:                ComponentDTO{Purl: "pkg:npm/lodash", Requirement: "^4.0.0"},
			expectedOriginalPurl: "pkg:npm/lodash",
			expectedOriginalReq:  "^4.0.0",
			expectedPurl:         "pkg:npm/lodash",
			expectedRequirement:  "^4.0.0",
		},
		{
			name:                 "Version in purl overrides existing requirement",
			input:                ComponentDTO{Purl: "pkg:npm/lodash@4.17.21", Requirement: "^4.0.0"},
			expectedOriginalPurl: "pkg:npm/lodash@4.17.21",
			expectedOriginalReq:  "^4.0.0",
			expectedPurl:         "pkg:npm/lodash",
			expectedRequirement:  "4.17.21",
		},
		{
			name:                 "Empty purl preserves original fields",
			input:                ComponentDTO{Purl: "", Requirement: "1.0.0"},
			expectedOriginalPurl: "",
			expectedOriginalReq:  "1.0.0",
			expectedPurl:         "",
			expectedRequirement:  "1.0.0",
		},
		{
			name:                 "Invalid purl preserves original fields",
			input:                ComponentDTO{Purl: "invalid-purl", Requirement: "2.0.0"},
			expectedOriginalPurl: "invalid-purl",
			expectedOriginalReq:  "2.0.0",
			expectedPurl:         "invalid-purl",
			expectedRequirement:  "2.0.0",
		},
		{
			name:                 "Scoped npm purl with version preserves original",
			input:                ComponentDTO{Purl: "pkg:npm/%40scope/name@1.2.3", Requirement: ""},
			expectedOriginalPurl: "pkg:npm/%40scope/name@1.2.3",
			expectedOriginalReq:  "",
			expectedPurl:         "pkg:npm/%40scope/name",
			expectedRequirement:  "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitiseComponents(s, []ComponentDTO{tt.input})
			if len(result) != 1 {
				t.Fatalf("expected 1 component, got %d", len(result))
			}
			c := result[0]
			if c.OriginalPurl != tt.expectedOriginalPurl {
				t.Errorf("OriginalPurl = %q, want %q", c.OriginalPurl, tt.expectedOriginalPurl)
			}
			if c.OriginalRequirement != tt.expectedOriginalReq {
				t.Errorf("OriginalRequirement = %q, want %q", c.OriginalRequirement, tt.expectedOriginalReq)
			}
			if c.Purl != tt.expectedPurl {
				t.Errorf("Purl = %q, want %q", c.Purl, tt.expectedPurl)
			}
			if c.Requirement != tt.expectedRequirement {
				t.Errorf("Requirement = %q, want %q", c.Requirement, tt.expectedRequirement)
			}
		})
	}
}

