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
	"fmt"
	"regexp"
	"strings"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"go.uber.org/zap"
)

var pkgRegex = regexp.MustCompile(`^pkg:(?P<type>\w+)/(?P<name>.+)$`) // regex to parse purl name from purl string
var typeRegex = regexp.MustCompile(`^(npm|nuget)$`)                   // regex to parse purl types that should not be lower cased

// sanitiseComponents validates and normalises a list of ComponentDTO into Components.
// It checks for empty or invalid PURLs, extracts version constraints from the PURL when
// the requirement is missing, and moves semver operators from the version to the requirement.
func sanitiseComponents(s *zap.SugaredLogger, componentDTOs []ComponentDTO) []Component {
	var components []Component
	for _, dto := range componentDTOs {
		// Check for empty purl
		if dto.Purl == "" {
			components = append(components, Component{
				Purl:                dto.Purl,
				OriginalPurl:        dto.Purl,
				OriginalRequirement: dto.Requirement,
				Requirement:         dto.Requirement,
				Status: domain.ComponentStatus{
					Message:    "Empty Purl",
					StatusCode: domain.InvalidPurl,
				},
			})
			continue
		}
		originalPurl := dto.Purl
		originalRequirement := dto.Requirement
		packageURL, err := purlhelper.PurlFromString(dto.Purl)
		if err != nil {
			s.Warnf("Failed to parse PURL %q (requirement: %q): %v", dto.Purl, dto.Requirement, err)
			components = append(components, Component{
				Purl:                dto.Purl,
				OriginalPurl:        dto.Purl,
				OriginalRequirement: dto.Requirement,
				Requirement:         dto.Requirement,
				Status: domain.ComponentStatus{
					Message:    "Invalid Purl",
					StatusCode: domain.InvalidPurl,
				},
			})
			continue
		}
		// If the PURL contains a version, extract it and move it to the requirement field.
		// Note: the version in the PURL always takes precedence over any existing requirement.
		// We use strings.Replace on the original PURL string (rather than packageURL.ToString())
		// to preserve percent-encoded characters (e.g., %40 in scoped npm packages like
		// pkg:npm/%40scope/name). This matters because component names in the database are
		// stored with %40, not @.
		if packageURL.Version != "" {
			dto.Requirement = packageURL.Version
			dto.Purl = strings.Replace(dto.Purl, "@"+packageURL.Version, "", 1)
		}
		qualifiers := make(map[string]string, len(packageURL.Qualifiers))
		for _, q := range packageURL.Qualifiers {
			qualifiers[q.Key] = q.Value
		}

		componentName, err := ComponentNameFromString(dto.Purl)
		if err != nil {
			s.Warnf("Failed to extract component name from PURL %q (requirement: %q): %v", dto.Purl, dto.Requirement, err)
			components = append(components, Component{
				Purl:                dto.Purl,
				OriginalPurl:        originalPurl,
				Requirement:         dto.Requirement,
				OriginalRequirement: originalRequirement,
				Status: domain.ComponentStatus{
					Message:    "Invalid Purl",
					StatusCode: domain.InvalidPurl,
				},
			})
			continue
		}

		URL, err := purlhelper.ProjectUrl(componentName, packageURL.Type)
		if err != nil {
			s.Warnf("Failed to derive project URL for PURL %q (requirement: %q): %v. URL will be empty.", dto.Purl, dto.Requirement, err)
			URL = ""
		}

		components = append(components, Component{
			Purl:                dto.Purl,
			OriginalPurl:        originalPurl,
			Requirement:         dto.Requirement,
			OriginalRequirement: originalRequirement,
			PurlType:            packageURL.Type,
			PurlName:            packageURL.Name,
			PurlQualifiers:      qualifiers,
			PurlNamespace:       packageURL.Namespace,
			PurlSubpath:         packageURL.Subpath,
			Name:                componentName,
			URL:                 URL,
			Status: domain.ComponentStatus{
				Message:    "",
				StatusCode: domain.Success,
			},
		})
	}
	return components
}

// ComponentNameFromString take an input Purl string and returns the component name only.
func ComponentNameFromString(purlString string) (string, error) {
	if len(purlString) == 0 {
		return "", fmt.Errorf("no purl string supplied to parse")
	}
	matches := pkgRegex.FindStringSubmatch(purlString)
	if len(matches) > 0 {
		ti := pkgRegex.SubexpIndex("type")
		ni := pkgRegex.SubexpIndex("name")
		if ni >= 0 {
			// Remove any version@/subpath?/qualifiers# info from the PURL
			pn := strings.Split(strings.Split(strings.Split(matches[ni], "@")[0], "?")[0], "#")[0]
			// Lowercase the purl name if it's not on the exclusion list (defined in the regex)
			if ti >= 0 && !typeRegex.MatchString(matches[ti]) {
				pn = strings.ToLower(pn)
			}
			return pn, nil
		}
	}
	return "", fmt.Errorf("no purl name found in '%v'", purlString)
}
