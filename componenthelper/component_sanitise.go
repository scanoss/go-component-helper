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

	"github.com/scanoss/go-component-helper/componenthelper/utils"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
)

// sanitiseComponents validates and normalises a list of ComponentDTO into Components.
// It checks for empty or invalid PURLs, extracts version constraints from the PURL when
// the requirement is missing, and moves semver operators from the version to the requirement.
func sanitiseComponents(componentDTOs []ComponentDTO) []Component {
	var components []Component
	for _, dto := range componentDTOs {
		// Check for empty purl
		if dto.Purl == "" {
			components = append(components, Component{
				Purl:        dto.Purl,
				Requirement: dto.Requirement,
				Status: domain.ComponentStatus{
					Message:    "Empty Purl",
					StatusCode: domain.InvalidPurl,
				},
			})
			continue
		}
		purlParts := strings.Split(dto.Purl, "@")
		// If version contains a semver operator, move it to requirement and strip from purl
		if len(purlParts) == 2 && utils.HasSemverOperator(purlParts[1]) {
			dto.Requirement = purlParts[1]
			dto.Purl = purlParts[0]
		}
		_, err := purlhelper.PurlFromString(dto.Purl)
		if err != nil {
			components = append(components, Component{
				Purl:        dto.Purl,
				Requirement: dto.Requirement,
				Status: domain.ComponentStatus{
					Message:    "Invalid Purl",
					StatusCode: domain.InvalidPurl,
				},
			})
			continue
		}
		if dto.Requirement == "" && len(purlParts) == 2 {
			dto.Purl = purlParts[0]
			dto.Requirement = purlParts[1]
		}
		components = append(components, Component{
			Requirement: dto.Requirement,
			Purl:        dto.Purl,
			Status: domain.ComponentStatus{
				Message:    "",
				StatusCode: domain.Success,
			},
		})
	}
	return components
}
