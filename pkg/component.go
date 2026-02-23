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
	"go.uber.org/zap"
	"sync"
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
	sanitizedComponents := sanitizeComponents(config.Input)
	numJobs := len(sanitizedComponents)
	jobs := make(chan Component, numJobs)
	results := make(chan Component, numJobs)
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = MaxWorkers
	}
	numWorkers := min(config.MaxWorkers, numJobs)
	wg := sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go componentVersionWorker(config.Ctx, config.S, config.DB, jobs, results, &wg)
	}
	for _, c := range sanitizedComponents {
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
