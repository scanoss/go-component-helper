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

// ComponentDTO is the standard input structure for component operations across all services.
type ComponentDTO struct {
	// Purl is the Package URL identifying the component (e.g., "pkg:github/scanoss/scanner.c").
	Purl string `json:"purl"`
	// Requirement is an optional version constraint (e.g., ">=1.0.0", "^2.3.0").
	Requirement string `json:"requirement,omitempty"`
}
