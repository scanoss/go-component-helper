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
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/scanoss/go-models/pkg/scanoss"
	"github.com/scanoss/go-models/pkg/services"
	"go.uber.org/zap"
)

// scanossResolver bundles the component and project services into a single
// componentResolver implementation by embedding both — each contributes its
// own methods (GetComponent / GetComponentVersions / GetSourcePurl from
// ComponentService, GetProject from ProjectService) without overlap.
type scanossResolver struct {
	*services.ComponentService
	*services.ProjectService
}

// ComponentVersionCfg holds the configuration for resolving component versions.
type ComponentVersionCfg struct {
	// MaxWorkers is the maximum number of concurrent goroutines used to resolve versions.
	// If <= 0, defaults to MaxWorkers (5).
	MaxWorkers int
	// Ctx is the context used for cancellation and deadline propagation.
	Ctx context.Context
	// S is the sugared logger for structured logging.
	S *zap.SugaredLogger
	// DB is the database connection used to query component data.
	DB *sqlx.DB
	// Input is the list of components whose versions need to be resolved.
	Input []ComponentDTO
}

const MaxWorkers = 5

// GetComponentsVersion resolves the concrete version for each component using a fan-out/fan-in
// concurrency pattern. It spawns up to MaxWorkers goroutines (capped by the number of components)
// to query versions in parallel, then collects and returns the results.
//
// Important: during sanitisation, if the input PURL contains a version (e.g., pkg:pypi/gtest@1.17.0),
// the version is extracted and moved to the Requirement field, overwriting any existing requirement.
// The PURL is then stored without the version (e.g., pkg:pypi/gtest).
func GetComponentsVersion(config ComponentVersionCfg) []Component {
	sanitisedComponents := sanitiseComponents(config.S, config.Input)
	numJobs := len(sanitisedComponents)
	jobs := make(chan Component, numJobs)
	results := make(chan Component, numJobs)
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = MaxWorkers
	}
	numWorkers := min(config.MaxWorkers, numJobs)
	sc := scanoss.New(config.DB)
	resolver := &scanossResolver{ComponentService: sc.Component, ProjectService: sc.Project}
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go componentVersionWorker(config.Ctx, config.S, resolver, jobs, results, &wg)
	}
	for _, c := range sanitisedComponents {
		jobs <- c
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	var processedComponents []Component
	for r := range results {
		processedComponents = append(processedComponents, r)
	}
	return processedComponents
}
