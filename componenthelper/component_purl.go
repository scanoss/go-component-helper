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

	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"go.uber.org/zap"
)

var pkgRegex = regexp.MustCompile(`^pkg:(?P<type>\w+)/(?P<name>.+)$`) // regex to parse purl name from purl string
var typeRegex = regexp.MustCompile(`^(npm|nuget)$`)                   // regex to parse purl types that should not be lower cased

// buildPurlInfo parses a PURL string into a PurlInfo. It decomposes the PURL,
// extracts qualifiers, derives the component name and a browsable project URL.
// Returns the parsed info and the version (if present in the PURL) so callers
// can decide what to do with it. An error is returned if the PURL cannot be
// parsed or the name cannot be extracted.
func buildPurlInfo(s *zap.SugaredLogger, purlString string) (PurlInfo, string, error) {
	packageURL, err := purlhelper.PurlFromString(purlString)
	if err != nil {
		return PurlInfo{}, "", fmt.Errorf("failed to parse purl %q: %w", purlString, err)
	}
	qualifiers := make(map[string]string, len(packageURL.Qualifiers))
	for _, q := range packageURL.Qualifiers {
		qualifiers[q.Key] = q.Value
	}
	componentName, err := ComponentNameFromString(purlString)
	if err != nil {
		return PurlInfo{}, packageURL.Version, fmt.Errorf("failed to extract component name from purl %q: %w", purlString, err)
	}
	URL, err := purlhelper.ProjectUrl(componentName, packageURL.Type)
	if err != nil {
		s.Warnf("Failed to derive project URL for PURL %q: %v. URL will be empty.", purlString, err)
		URL = ""
	}
	return PurlInfo{
		Purl:           purlString,
		Name:           componentName,
		URL:            URL,
		PurlType:       packageURL.Type,
		PurlName:       packageURL.Name,
		PurlNamespace:  packageURL.Namespace,
		PurlQualifiers: qualifiers,
		PurlSubpath:    packageURL.Subpath,
	}, packageURL.Version, nil
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
