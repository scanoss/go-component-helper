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
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

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

func TestBuildPurlInfo(t *testing.T) {
	if err := zlog.NewSugaredDevLogger(); err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	tests := []struct {
		name              string
		purl              string
		wantErr           bool
		wantVersion       string
		wantName          string
		wantURL           string
		wantPurlType      string
		wantPurlName      string
		wantPurlNamespace string
		wantPurlSubpath   string
		wantQualifiers    map[string]string
	}{
		{
			name:         "Plain npm without namespace",
			purl:         "pkg:npm/lodash",
			wantName:     "lodash",
			wantURL:      "https://www.npmjs.com/package/lodash",
			wantPurlType: "npm",
			wantPurlName: "lodash",
		},
		{
			name:         "Npm with version moves to wantVersion",
			purl:         "pkg:npm/lodash@4.17.21",
			wantVersion:  "4.17.21",
			wantName:     "lodash",
			wantURL:      "https://www.npmjs.com/package/lodash",
			wantPurlType: "npm",
			wantPurlName: "lodash",
		},
		{
			name:              "Github with namespace",
			purl:              "pkg:github/scanoss/scanner.c",
			wantName:          "scanoss/scanner.c",
			wantURL:           "https://github.com/scanoss/scanner.c",
			wantPurlType:      "github",
			wantPurlName:      "scanner.c",
			wantPurlNamespace: "scanoss",
		},
		{
			name:              "Scoped npm with percent-encoded namespace",
			purl:              "pkg:npm/%40scope/name",
			wantName:          "%40scope/name",
			wantURL:           "https://www.npmjs.com/package/%40scope/name",
			wantPurlType:      "npm",
			wantPurlName:      "name",
			wantPurlNamespace: "@scope",
		},
		{
			name:         "Nuget preserves case in name",
			purl:         "pkg:nuget/Newtonsoft.Json",
			wantName:     "Newtonsoft.Json",
			wantURL:      "https://www.nuget.org/packages/Newtonsoft.Json",
			wantPurlType: "nuget",
			wantPurlName: "Newtonsoft.Json",
		},
		{
			name:              "Maven with qualifiers",
			purl:              "pkg:maven/com.example/lib?type=jar",
			wantName:          "com.example/lib",
			wantURL:           "https://mvnrepository.com/artifact/com.example/lib",
			wantPurlType:      "maven",
			wantPurlName:      "lib",
			wantPurlNamespace: "com.example",
			wantQualifiers:    map[string]string{"type": "jar"},
		},
		{
			name:              "Github with subpath",
			purl:              "pkg:github/scanoss/scanner.c#src/lib",
			wantName:          "scanoss/scanner.c",
			wantURL:           "https://github.com/scanoss/scanner.c",
			wantPurlType:      "github",
			wantPurlName:      "scanner.c",
			wantPurlNamespace: "scanoss",
			wantPurlSubpath:   "src/lib",
		},
		{
			name:         "Unsupported type leaves URL empty",
			purl:         "pkg:cargo/serde",
			wantName:     "serde",
			wantURL:      "",
			wantPurlType: "cargo",
			wantPurlName: "serde",
		},
		{
			name:    "Empty purl returns error",
			purl:    "",
			wantErr: true,
		},
		{
			name:    "Garbage string returns error",
			purl:    "not-a-purl",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, version, err := buildPurlInfo(s, tt.purl)
			if tt.wantErr {
				if err == nil {
					t.Errorf("buildPurlInfo(%q) expected error, got %+v", tt.purl, info)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildPurlInfo(%q) unexpected error: %v", tt.purl, err)
			}
			if version != tt.wantVersion {
				t.Errorf("version = %q, want %q", version, tt.wantVersion)
			}
			if info.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", info.Name, tt.wantName)
			}
			if info.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", info.URL, tt.wantURL)
			}
			if info.PurlType != tt.wantPurlType {
				t.Errorf("PurlType = %q, want %q", info.PurlType, tt.wantPurlType)
			}
			if info.PurlName != tt.wantPurlName {
				t.Errorf("PurlName = %q, want %q", info.PurlName, tt.wantPurlName)
			}
			if info.PurlNamespace != tt.wantPurlNamespace {
				t.Errorf("PurlNamespace = %q, want %q", info.PurlNamespace, tt.wantPurlNamespace)
			}
			if info.PurlSubpath != tt.wantPurlSubpath {
				t.Errorf("PurlSubpath = %q, want %q", info.PurlSubpath, tt.wantPurlSubpath)
			}
			if len(info.PurlQualifiers) != len(tt.wantQualifiers) {
				t.Errorf("PurlQualifiers size = %d, want %d", len(info.PurlQualifiers), len(tt.wantQualifiers))
			}
			for k, v := range tt.wantQualifiers {
				if info.PurlQualifiers[k] != v {
					t.Errorf("PurlQualifiers[%q] = %q, want %q", k, info.PurlQualifiers[k], v)
				}
			}
		})
	}
}

