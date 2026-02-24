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

import "github.com/scanoss/go-grpc-helper/pkg/grpc/domain"

// Component represents the component entity used across all services.
type Component struct {
	// Purl is the Package URL identifying the component.
	Purl string `json:"purl"`
	// Requirement is the version constraint used to resolve the component.
	Requirement string `json:"requirement,omitempty"`
	// Version is the resolved concrete version after processing.
	Version string `json:"version,omitempty"`
	// Status indicates the outcome of the resolution (e.g., success, not found, invalid purl).
	Status domain.ComponentStatus
}
