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

package utils

import (
	"math"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// operatorRegex matches semver range operators (>=, <=, ~, ^, =, >, <) at the start of a string.
var operatorRegex = regexp.MustCompile(`^(>=|<=|~|\^|=|>|<)`)

// zeroVersion is used as a fallback for versions that cannot be parsed as semver.
var zeroVersion, _ = semver.NewVersion("v0.0.0")

// HasSemverOperator reports whether the given version string starts with a semver range operator.
func HasSemverOperator(version string) bool {
	return operatorRegex.MatchString(version)
}

// FindNearestVersion finds the version from candidates that is closest to the target requirement.
// It strips any semver operators from the requirement before comparing.
// Candidates that are not valid semver are treated as v0.0.0 (following the pickOneUrl approach).
// If the requirement itself is not valid semver, it returns an empty string.
func FindNearestVersion(requirement string, candidates []string) string {
	target := stripOperators(requirement)
	targetVer, err := semver.NewVersion(target)
	if err != nil {
		return ""
	}

	var nearest *semver.Version
	var nearestOriginal string
	minDistance := math.MaxFloat64

	for _, c := range candidates {
		raw := strings.TrimSpace(c)
		v, errSemver := semver.NewVersion(raw)
		if errSemver != nil {
			v = zeroVersion
		}
		d := semverDistance(targetVer, v)
		if d < minDistance || (d == minDistance && nearest != nil && v.GreaterThan(nearest)) {
			minDistance = d
			nearest = v
			nearestOriginal = raw
		}
	}

	if nearest == nil {
		return ""
	}
	return nearestOriginal
}

// stripOperators removes leading semver range operators and surrounding whitespace from a version string.
func stripOperators(version string) string {
	return strings.TrimSpace(operatorRegex.ReplaceAllString(strings.TrimSpace(version), ""))
}

// semverDistance computes a weighted numeric distance between two semver versions.
// Major differences are weighted most heavily, then minor, then patch.
func semverDistance(a, b *semver.Version) float64 {
	majorDiff := math.Abs(float64(a.Major()) - float64(b.Major()))
	minorDiff := math.Abs(float64(a.Minor()) - float64(b.Minor()))
	patchDiff := math.Abs(float64(a.Patch()) - float64(b.Patch()))
	return majorDiff*1_000_000 + minorDiff*1_000 + patchDiff
}
