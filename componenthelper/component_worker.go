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
	"github.com/scanoss/go-models/pkg/services"
	"sync"

	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"github.com/scanoss/go-models/pkg/types"
	"go.uber.org/zap"
)

// componentResolver abstracts the component lookup operations used by the worker.
type componentResolver interface {
	GetComponent(ctx context.Context, req types.ComponentRequest) (types.ComponentResponse, error)
}

// componentVersionWorker processes components from the jobs channel, resolving each component's
// version by querying the SCANOSS API. If a resolved version is found, it replaces the original;
// otherwise the existing version is preserved.
func componentVersionWorker(ctx context.Context, s *zap.SugaredLogger, resolver componentResolver, jobs chan Component, results chan Component, wg *sync.WaitGroup) {
	defer wg.Done()
	for j := range jobs {

		processedComponent := Component{
			Purl:           j.Purl,
			Requirement:    j.Requirement,
			Version:        j.Version,
			Name:           j.Name,
			Status:         j.Status,
			PurlType:       j.PurlType,
			PurlSubpath:    j.PurlSubpath,
			PurlNamespace:  j.PurlNamespace,
			PurlQualifiers: j.PurlQualifiers,
			PurlName:       j.PurlName,
			URL:            j.URL,
		}

		if processedComponent.Status.StatusCode != domain.Success {
			results <- processedComponent
			continue
		}

		// Set by default version = requirement
		var component types.ComponentResponse
		component, err := resolver.GetComponent(ctx, types.ComponentRequest{
			Purl:        j.Purl,
			Requirement: j.Requirement,
		})

		if err != nil {
			if errors.Is(err, services.ErrComponentNotFound) {
				s.Warnf("Purl doesn't exist: %s, %v", j.Purl, err)
				processedComponent.Status.StatusCode = domain.ComponentNotFound
				processedComponent.Status.Message = "Component not found"
				results <- processedComponent
				continue
			}
			if errors.Is(err, services.ErrVersionNotFound) {
				s.Warnf("Failed to get component version: %s, %s", j.Purl, j.Requirement)
				processedComponent.Status.StatusCode = domain.VersionNotFound
				processedComponent.Status.Message = "Component version not found"
				results <- processedComponent
				continue
			}
			s.Errorf("Failed to get component: %s, %s, %v", j.Purl, j.Requirement, err)
			processedComponent.Status.StatusCode = domain.ComponentNotFound
			processedComponent.Status.Message = "Component not found"
			results <- processedComponent
			continue
		}
		if component.Version != "" {
			processedComponent.Version = component.Version
		}
		results <- processedComponent
	}
}
