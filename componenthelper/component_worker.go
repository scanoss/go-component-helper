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
	"errors"
	"fmt"
	"sync"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"github.com/scanoss/go-models/pkg/services"
	"github.com/scanoss/go-models/pkg/types"
	"go.uber.org/zap"
)

// componentResolver abstracts the component lookup operations used by the worker.
type componentResolver interface {
	GetComponent(ctx context.Context, req types.ComponentRequest) (types.ComponentResponse, error)
	GetComponentVersions(ctx context.Context, purl string) (types.ComponentVersionsResponse, error)
	// GetProject returns the projects-table row for the given PURL. The same
	// row drives two enrichments: a PurlInfo fallback when the component
	// lookup fails, and the SourcePurl built from the linked source-mine
	// fields when the component resolves successfully. Implementations
	// should return services.ErrProjectNotFound when no project exists.
	GetProject(ctx context.Context, purl string) (types.Project, error)
}

// componentVersionWorker processes components from the jobs channel, resolving each component's
// version by querying the SCANOSS API. If a resolved version is found, it replaces the original;
// otherwise the existing version is preserved.
func componentVersionWorker(ctx context.Context, s *zap.SugaredLogger, resolver componentResolver, jobs chan Component, results chan Component, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {
		processedComponent := j

		if processedComponent.Status.StatusCode != domain.Success {
			results <- processedComponent
			continue
		}
		// Add list of component versions
		processedComponent.Versions = []string{}

		// Set by default version = requirement
		var component types.ComponentResponse
		component, err := resolver.GetComponent(ctx, types.ComponentRequest{
			Purl:        j.Purl,
			Requirement: j.Requirement,
		})

		switch {
		case errors.Is(err, services.ErrComponentNotFound):
			s.Warnf("Purl doesn't exist: %s, %v", j.Purl, err)
			processedComponent.Status.StatusCode = domain.ComponentNotFound
			processedComponent.Status.Message = "Component not found"
		case errors.Is(err, services.ErrVersionNotFound):
			s.Warnf("Failed to get component version: %s, %s", j.Purl, j.Requirement)
			processedComponent.Status.StatusCode = domain.VersionNotFound
			processedComponent.Status.Message = "Component version not found"
		case err != nil:
			s.Errorf("Failed to get component: %s, %s, %v", j.Purl, j.Requirement, err)
			processedComponent.Status.StatusCode = domain.ComponentNotFound
			processedComponent.Status.Message = "Component not found"
		default:
			if component.Version != "" {
				processedComponent.Version = component.Version
			}
		}

		// Single project lookup: drives the PurlInfo fallback when the
		// component lookup fails AND the SourcePurl when it succeeds.
		project, errProj := resolver.GetProject(ctx, j.Purl)
		if errProj != nil && !errors.Is(errProj, services.ErrProjectNotFound) {
			s.Warnf("Failed to get project for %s: %v", j.Purl, errProj)
		}

		// Fallback: when the component wasn't resolved, enrich PurlInfo from projects
		// and upgrade status to VersionNotFound
		if processedComponent.Status.StatusCode == domain.ComponentNotFound && errProj == nil {
			enrichPurlInfoFromProject(s, &processedComponent, project)
		}

		// Retrieve component versions on success or when version is not found
		if processedComponent.Status.StatusCode != domain.ComponentNotFound {
			componentVersions, errCompVersion := resolver.GetComponentVersions(ctx, processedComponent.Purl)
			if errCompVersion != nil {
				s.Errorf("Failed to get component versions: %s, %v", processedComponent.Purl, errCompVersion)
			} else {
				processedComponent.Versions = componentVersions.Versions
			}
		}

		if processedComponent.Status.StatusCode == domain.Success && errProj == nil {
			processedComponent.SourcePurl = buildSourcePurlFromProject(s, project, j.Purl)
		}

		results <- processedComponent
	}
}

// enrichPurlInfoFromProject rebuilds the component's PurlInfo from the
// project row's canonical purl_name/purl_type and promotes the status from
// ComponentNotFound to VersionNotFound. On parse failure the status is left
// as ComponentNotFound and the existing PurlInfo is preserved.
func enrichPurlInfoFromProject(s *zap.SugaredLogger, c *Component, project types.Project) {
	projectPurl := fmt.Sprintf("pkg:%s/%s", project.PurlType, project.PurlName)
	purlInfo, _, errBuild := buildPurlInfo(s, projectPurl)
	if errBuild != nil {
		s.Warnf("Failed to parse project PURL %q for %s: %v", projectPurl, c.Purl, errBuild)
		return
	}
	c.PurlInfo = purlInfo
	c.Status.StatusCode = domain.VersionNotFound
	c.Status.Message = "Component version not found"
}

// buildSourcePurlFromProject builds a SourcePurl from the project's linked
// source-mine fields. Returns nil when no source is linked. On parse failure
// returns a SourcePurl with a ComponentNotFound status so the caller can
// surface that the source data was malformed.
func buildSourcePurlFromProject(s *zap.SugaredLogger, project types.Project, parentPurl string) *SourcePurl {
	if project.SourceMineID == nil || project.SourcePurlName == nil {
		return nil
	}
	sourcePurlString := fmt.Sprintf("pkg:%s/%s", project.SourcePurlType, *project.SourcePurlName)
	sourceInfo, _, errBuild := buildPurlInfo(s, sourcePurlString)
	if errBuild != nil {
		s.Warnf("Failed to parse source PURL %q for %s: %v", sourcePurlString, parentPurl, errBuild)
		return &SourcePurl{
			Status: domain.ComponentStatus{
				Message:    "Source component not found",
				StatusCode: domain.ComponentNotFound,
			},
		}
	}
	if project.SourceRepositoryURL != "" {
		sourceInfo.URL = project.SourceRepositoryURL
	}
	return &SourcePurl{
		PurlInfo: sourceInfo,
		Status:   domain.ComponentStatus{StatusCode: domain.Success},
	}
}
