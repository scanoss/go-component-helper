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

func TestComponentNameFromString(t *testing.T) {
	tests := []struct {
		name      string
		purl      string
		expected  string
		expectErr bool
	}{
		{
			name:     "Scoped npm with percent-encoded @",
			purl:     "pkg:npm/%40medius-ui/control",
			expected: "%40medius-ui/control",
		},
		{
			name:     "Conan package without namespace",
			purl:     "pkg:conan/gtest",
			expected: "gtest",
		},
		{
			name:     "Scoped npm @grpc with percent-encoded @",
			purl:     "pkg:npm/%40grpc/grpc-js",
			expected: "%40grpc/grpc-js",
		},
		{
			name:     "Npm without namespace",
			purl:     "pkg:npm/cli-progress",
			expected: "cli-progress",
		},
		{
			name:     "Golang with multi-level namespace",
			purl:     "pkg:golang/github.com/Masterminds/semver/v3",
			expected: "github.com/masterminds/semver/v3",
		},
		{
			name:     "Golang simple namespace",
			purl:     "pkg:golang/github.com/jubittajohn/kubernetes",
			expected: "github.com/jubittajohn/kubernetes",
		},
		{
			name:     "Golang deep namespace path",
			purl:     "pkg:golang/go.unistack.org/micro/v3/util/http",
			expected: "go.unistack.org/micro/v3/util/http",
		},
		{
			name:     "Scoped npm @istanbuljs with percent-encoded @",
			purl:     "pkg:npm/%40istanbuljs/nyc-config-typescript",
			expected: "%40istanbuljs/nyc-config-typescript",
		},
		{
			name:     "Npm plain package",
			purl:     "pkg:npm/scanoss",
			expected: "scanoss",
		},
		{
			name:     "Purl with version is stripped",
			purl:     "pkg:npm/lodash@4.17.21",
			expected: "lodash",
		},
		{
			name:     "Nuget preserves case",
			purl:     "pkg:nuget/Newtonsoft.Json",
			expected: "Newtonsoft.Json",
		},
		{
			name:      "Empty string returns error",
			purl:      "",
			expectErr: true,
		},
		{
			name:      "Invalid purl returns error",
			purl:      "not-a-purl",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ComponentNameFromString(tt.purl)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ComponentNameFromString() expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ComponentNameFromString() unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ComponentNameFromString() = %q, want %q", got, tt.expected)
			}
		})
	}
}
