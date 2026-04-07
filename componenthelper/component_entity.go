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
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
)

// Component represents the component entity used across all services.
type Component struct {
	// Purl is the Package URL identifying the component.
	Purl string `json:"purl"`
	// OriginalPurl is the original Package URL as provided before any sanitisation or resolution.
	OriginalPurl string `json:"original_purl"`
	// OriginalRequirement is the original version requirement as provided before any sanitisation or resolution.
	OriginalRequirement string `json:"original_requirement"`
	// Requirement is the version constraint used to resolve the component.
	Requirement string `json:"requirement,omitempty"`
	// Version is the resolved concrete version after processing.
	Version string `json:"version,omitempty"`
	// Versions list of component versions
	Versions []string `json:"versions,omitempty"`
	// Name namespace + name
	Name string `json:"component_name,omitempty"`
	// URL component URL
	URL string `json:"url,omitempty"`
	// PurlType is the package type (e.g., "golang", "npm", "maven").
	PurlType string `json:"purl_type,omitempty"`
	// PurlName is the package name.
	PurlName string `json:"purl_name,omitempty"`
	// PurlNamespace is the package namespace (e.g., "github.com/scanoss").
	PurlNamespace string `json:"purl_namespace,omitempty"`
	// PurlQualifiers holds key-value pairs for extra qualifying data.
	PurlQualifiers map[string]string `json:"purl_qualifiers,omitempty"`
	// PurlSubpath is the subpath within the package.
	PurlSubpath string `json:"purl_subpath,omitempty"`
	// Component Status
	Status domain.ComponentStatus
}

func (c *Component) FindNearestVersion() string {
	return FindNearestVersion(c.Requirement, c.Versions)
}
