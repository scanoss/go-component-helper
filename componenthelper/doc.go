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

// Package componenthelper provides utilities for resolving component versions
// from package URLs (purls) and version requirement strings.
//
// It uses a fan-out/fan-in concurrency pattern to query a database for matching
// versions in parallel, controlled by a configurable number of worker goroutines.
//
// # Usage
//
// Create a [ComponentVersionCfg] with the desired configuration and call
// [GetComponentsVersion] to resolve versions for a list of components:
//
//	components := componenthelper.GetComponentsVersion(componenthelper.ComponentVersionCfg{
//		MaxWorkers: 10,
//		Ctx:        ctx,
//		S:          sugar,
//		DB:         db,
//		Input: []componenthelper.ComponentDTO{
//			{Purl: "pkg:github/scanoss/scanner.c", Requirement: ">=1.0.0"},
//		},
//	})
//
// Each returned [Component] contains the resolved version and a status indicating
// whether the resolution succeeded.
package componenthelper
