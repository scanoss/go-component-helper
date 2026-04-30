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
	"strings"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"go.uber.org/zap"
)

// sanitiseComponents validates and normalises a list of ComponentDTO into Components.
// It checks for empty or invalid PURLs, extracts version constraints from the PURL when
// the requirement is missing, and moves semver operators from the version to the requirement.
func sanitiseComponents(s *zap.SugaredLogger, componentDTOs []ComponentDTO) []Component {
	var components []Component
	for _, dto := range componentDTOs {
		// Check for empty purl
		if dto.Purl == "" {
			components = append(components, Component{
				PurlInfo: PurlInfo{
					Purl: dto.Purl,
				},
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

		purlInfo, purlVersion, err := buildPurlInfo(s, dto.Purl)
		if err != nil {
			s.Warnf("%v", err)
			components = append(components, Component{
				PurlInfo: PurlInfo{
					Purl: dto.Purl,
				},
				OriginalPurl:        originalPurl,
				OriginalRequirement: originalRequirement,
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
		if purlVersion != "" {
			dto.Requirement = purlVersion
			dto.Purl = strings.Replace(dto.Purl, "@"+purlVersion, "", 1)
		}
		purlInfo.Purl = dto.Purl
		components = append(components, Component{
			OriginalPurl:        originalPurl,
			Requirement:         dto.Requirement,
			OriginalRequirement: originalRequirement,
			PurlInfo:            purlInfo,
			Status: domain.ComponentStatus{
				Message:    "",
				StatusCode: domain.Success,
			},
		})
	}
	return components
}
