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
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
)

<<<<<<< Updated upstream
// componentResolver abstracts the component lookup operations used by the worker.
type componentResolver interface {
	CheckPurl(ctx context.Context, purl string) (int, error)
	GetComponent(ctx context.Context, req types.ComponentRequest) (types.ComponentResponse, error)
}

=======
// Component represents the component entity used across all services.
>>>>>>> Stashed changes
type Component struct {
	Purl        string `json:"purl"`
	Requirement string `json:"requirement,omitempty"`
	Version     string `json:"version,omitempty"`
	Status      domain.ComponentStatus
}
