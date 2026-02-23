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
	"go.uber.org/zap"
)

type ComponentVersionCfg struct {
	MaxWorkers int
	Ctx        context.Context
	S          *zap.SugaredLogger
	DB         *sqlx.DB
	Input      []ComponentDTO
}

const MaxWorkers = 5

// GetComponentsVersion resolves the concrete version for each component using a fan-out/fan-in
// concurrency pattern. It spawns up to MaxWorkers goroutines (capped by the number of components)
// to query versions in parallel, then collects and returns the results.
func GetComponentsVersion(config ComponentVersionCfg) []Component {
	sanitisedComponents := sanitiseComponents(config.Input)
	numJobs := len(sanitisedComponents)
	jobs := make(chan Component, numJobs)
	results := make(chan Component, numJobs)
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = MaxWorkers
	}
	numWorkers := min(config.MaxWorkers, numJobs)
	sc := scanoss.New(config.DB)
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go componentVersionWorker(config.Ctx, config.S, sc.Component, jobs, results, &wg)
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
