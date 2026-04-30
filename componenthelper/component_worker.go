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
	// GetSourcePurl returns the raw source-mine data used to build a source
	// PURL for the given component PURL. Implementations should return
	// services.ErrSourcePurlNotFound when no source PURL exists.
	GetSourcePurl(ctx context.Context, purl string) (types.SourcePurl, error)
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

		// Retrieve component versions on success or when version is not found
		if !errors.Is(err, services.ErrComponentNotFound) && processedComponent.Status.StatusCode != domain.ComponentNotFound {
			componentVersions, errCompVersion := resolver.GetComponentVersions(ctx, j.Purl)
			if errCompVersion != nil {
				s.Errorf("Failed to get component versions: %s, %v", j.Purl, errCompVersion)
			} else {
				processedComponent.Versions = componentVersions.Versions
			}
		}

		// Resolve the source PURL (e.g. upstream source-mine equivalent)
		// only when the component itself was resolved successfully. A missing
		// source PURL is normal and should not affect the component status.
		if processedComponent.Status.StatusCode == domain.Success {
			sourcePurl, errSrc := resolver.GetSourcePurl(ctx, j.Purl)
			switch {
			case errors.Is(errSrc, services.ErrSourcePurlNotFound):
				// No source PURL for this component — leave nil.
			case errSrc != nil:
				s.Warnf("Failed to get source PURL for %s: %v", j.Purl, errSrc)
			default:
				sourcePurlString := buildSourcePurlString(sourcePurl)
				sourceInfo, _, errBuild := buildPurlInfo(s, sourcePurlString)
				switch {
				case errBuild != nil:
					s.Warnf("Failed to parse source PURL %q for %s: %v", sourcePurlString, j.Purl, errBuild)
					processedComponent.SourcePurl = &SourcePurl{
						Status: domain.ComponentStatus{
							Message:    "Invalid Source Purl",
							StatusCode: domain.InvalidPurl,
						},
					}
				case sourceInfo.PurlType == processedComponent.PurlType && sourceInfo.Name == processedComponent.Name:
					// Source PURL resolves to the same component — leave nil.
				default:
					if sourcePurl.RepositoryURL != "" {
						sourceInfo.URL = sourcePurl.RepositoryURL
					}
					processedComponent.SourcePurl = &SourcePurl{
						PurlInfo: sourceInfo,
						Status: domain.ComponentStatus{
							StatusCode: domain.Success,
						},
					}
				}
			}
		}

		results <- processedComponent
	}
}
