// SPDX-License-Identifier: GPL-2.0-or-later
/*
 * Copyright (C) 2026 SCANOSS.COM
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package pkg

import (
	"github.com/scanoss/go-component-helper/pkg/utils"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	purlhelper "github.com/scanoss/go-purl-helper/pkg"
	"strings"
)

func sanitizeComponents(componentDTOs []ComponentDTO) []Component {
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
