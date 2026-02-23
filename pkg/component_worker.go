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
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"github.com/scanoss/go-models/pkg/scanoss"
	"github.com/scanoss/go-models/pkg/types"
	"go.uber.org/zap"
	"sync"
)

// componentVersionWorker processes components from the jobs channel, resolving each component's
// version by querying the SCANOSS API. If a resolved version is found, it replaces the original;
// otherwise the existing version is preserved.
func componentVersionWorker(ctx context.Context, s *zap.SugaredLogger, db *sqlx.DB, jobs chan Component, results chan Component, wg *sync.WaitGroup) {
	defer wg.Done()
	sc := scanoss.New(db)
	for j := range jobs {
		processedComponent := Component{
			Purl:        j.Purl,
			Requirement: j.Requirement,
			Version:     j.Version,
			Status:      j.Status,
		}

		if processedComponent.Status.StatusCode != domain.Success {
			results <- processedComponent
			continue
		}

		// Check PURL exists
		count, err := sc.Component.CheckPurl(ctx, j.Purl)
		if count <= 0 || err != nil {
			s.Warnf("Purl doesn't exist: %s", j.Purl)
			processedComponent.Status.StatusCode = domain.ComponentNotFound
			processedComponent.Status.Message = "Component not found"
			results <- processedComponent
			continue
		}

		// Set by default version = requirement
		var component types.ComponentResponse
		component, err = sc.Component.GetComponent(ctx, types.ComponentRequest{
			Purl:        j.Purl,
			Requirement: j.Requirement,
		})
		if err != nil {
			s.Warnf("Failed to get component version: %s, %s", j.Purl, j.Requirement)
			processedComponent.Status.StatusCode = domain.VersionNotFound
			processedComponent.Status.Message = "Component version not found"
			results <- processedComponent
			continue
		}
		if component.Version != "" {
			processedComponent.Version = component.Version
		}
		results <- processedComponent
	}
}
