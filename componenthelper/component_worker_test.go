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
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"github.com/scanoss/go-models/pkg/types"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

type mockComponentResolver struct {
	getComponentFn func(ctx context.Context, req types.ComponentRequest) (types.ComponentResponse, error)
}

func (m *mockComponentResolver) GetComponent(ctx context.Context, req types.ComponentRequest) (types.ComponentResponse, error) {
	return m.getComponentFn(ctx, req)
}

func TestComponentVersionWorker_NonSuccessStatusPassthrough(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	tests := []struct {
		name                string
		input               Component
		expectedStatus      domain.StatusCode
		expectedMessage     string
		expectedPurl        string
		expectedRequirement string
	}{
		{
			name: "InvalidPurl status is passed through unchanged",
			input: Component{
				Purl:        "invalid-purl",
				Requirement: "",
				Status: domain.ComponentStatus{
					Message:    "Invalid Purl",
					StatusCode: domain.InvalidPurl,
				},
			},
			expectedStatus:      domain.InvalidPurl,
			expectedMessage:     "Invalid Purl",
			expectedPurl:        "invalid-purl",
			expectedRequirement: "",
		},
		{
			name: "Empty purl with InvalidPurl status is passed through",
			input: Component{
				Purl:        "",
				Requirement: "1.0.0",
				Status: domain.ComponentStatus{
					Message:    "Empty Purl",
					StatusCode: domain.InvalidPurl,
				},
			},
			expectedStatus:      domain.InvalidPurl,
			expectedMessage:     "Empty Purl",
			expectedPurl:        "",
			expectedRequirement: "1.0.0",
		},
		{
			name: "ComponentNotFound status is passed through",
			input: Component{
				Purl:        "pkg:npm/nonexistent",
				Requirement: "1.0.0",
				Status: domain.ComponentStatus{
					Message:    "Component not found",
					StatusCode: domain.ComponentNotFound,
				},
			},
			expectedStatus:      domain.ComponentNotFound,
			expectedMessage:     "Component not found",
			expectedPurl:        "pkg:npm/nonexistent",
			expectedRequirement: "1.0.0",
		},
		{
			name: "VersionNotFound status is passed through",
			input: Component{
				Purl:        "pkg:npm/lodash",
				Requirement: ">=99.0.0",
				Status: domain.ComponentStatus{
					Message:    "Component version not found",
					StatusCode: domain.VersionNotFound,
				},
			},
			expectedStatus:      domain.VersionNotFound,
			expectedMessage:     "Component version not found",
			expectedPurl:        "pkg:npm/lodash",
			expectedRequirement: ">=99.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan Component, 1)
			results := make(chan Component, 1)
			wg := sync.WaitGroup{}
			wg.Add(1)

			go componentVersionWorker(ctx, s, nil, jobs, results, &wg)

			jobs <- tt.input
			close(jobs)

			wg.Wait()
			close(results)

			result := <-results
			if result.Status.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %s, got %s", tt.expectedStatus, result.Status.StatusCode)
			}
			if result.Status.Message != tt.expectedMessage {
				t.Errorf("expected message %q, got %q", tt.expectedMessage, result.Status.Message)
			}
			if result.Purl != tt.expectedPurl {
				t.Errorf("expected purl %q, got %q", tt.expectedPurl, result.Purl)
			}
			if result.Requirement != tt.expectedRequirement {
				t.Errorf("expected requirement %q, got %q", tt.expectedRequirement, result.Requirement)
			}
		})
	}
}

func TestComponentVersionWorker_PreservesVersion(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	input := Component{
		Purl:        "pkg:npm/lodash",
		Requirement: "4.17.21",
		Version:     "4.17.21",
		Status: domain.ComponentStatus{
			Message:    "Component version not found",
			StatusCode: domain.VersionNotFound,
		},
	}

	jobs := make(chan Component, 1)
	results := make(chan Component, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go componentVersionWorker(ctx, s, nil, jobs, results, &wg)

	jobs <- input
	close(jobs)

	wg.Wait()
	close(results)

	result := <-results
	if result.Version != "4.17.21" {
		t.Errorf("expected version %q to be preserved, got %q", "4.17.21", result.Version)
	}
}

func TestComponentVersionWorker_MultipleJobs(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	inputs := []Component{
		{
			Purl: "invalid1",
			Status: domain.ComponentStatus{
				Message:    "Invalid Purl",
				StatusCode: domain.InvalidPurl,
			},
		},
		{
			Purl: "",
			Status: domain.ComponentStatus{
				Message:    "Empty Purl",
				StatusCode: domain.InvalidPurl,
			},
		},
		{
			Purl:        "invalid2",
			Requirement: "1.0.0",
			Status: domain.ComponentStatus{
				Message:    "Invalid Purl",
				StatusCode: domain.InvalidPurl,
			},
		},
	}

	jobs := make(chan Component, len(inputs))
	results := make(chan Component, len(inputs))
	wg := sync.WaitGroup{}
	wg.Add(1)

	go componentVersionWorker(ctx, s, nil, jobs, results, &wg)

	for _, input := range inputs {
		jobs <- input
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var collected []Component
	for r := range results {
		collected = append(collected, r)
	}

	if len(collected) != len(inputs) {
		t.Fatalf("expected %d results, got %d", len(inputs), len(collected))
	}

	for _, c := range collected {
		if c.Status.StatusCode != domain.InvalidPurl {
			t.Errorf("expected all components to have InvalidPurl status, got %s for purl %q", c.Status.StatusCode, c.Purl)
		}
	}
}

func TestComponentVersionWorker_ComponentResolverSuccess(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	tests := []struct {
		name            string
		input           Component
		mock            *mockComponentResolver
		expectedStatus  domain.StatusCode
		expectedMessage string
		expectedVersion string
	}{
		{
			name: "GetComponent success replaces version",
			input: Component{
				Purl:        "pkg:npm/lodash",
				Requirement: "^4.0.0",
				Version:     "4.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{Version: "4.17.21"}, nil
				},
			},
			expectedStatus:  domain.Success,
			expectedVersion: "4.17.21",
		},
		{
			name: "GetComponent success with empty version preserves original",
			input: Component{
				Purl:        "pkg:npm/lodash",
				Requirement: "^4.0.0",
				Version:     "4.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{Version: ""}, nil
				},
			},
			expectedStatus:  domain.Success,
			expectedVersion: "4.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan Component, 1)
			results := make(chan Component, 1)
			wg := sync.WaitGroup{}
			wg.Add(1)

			go componentVersionWorker(ctx, s, tt.mock, jobs, results, &wg)

			jobs <- tt.input
			close(jobs)

			wg.Wait()
			close(results)

			result := <-results
			if result.Status.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, result.Status.StatusCode)
			}
			if tt.expectedMessage != "" && result.Status.Message != tt.expectedMessage {
				t.Errorf("expected message %q, got %q", tt.expectedMessage, result.Status.Message)
			}
			if tt.expectedVersion != "" && result.Version != tt.expectedVersion {
				t.Errorf("expected version %q, got %q", tt.expectedVersion, result.Version)
			}
		})
	}
}

func TestComponentVersionWorker_ComponentResolverInvalid(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	tests := []struct {
		name            string
		input           Component
		mock            *mockComponentResolver
		expectedStatus  domain.StatusCode
		expectedMessage string
		expectedVersion string
	}{
		{
			name: "Purl not found returns ComponentNotFound",
			input: Component{
				Purl:        "pkg:npm/unknown",
				Requirement: "1.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{}, errors.New("component not found")
				},
			},
			expectedStatus:  domain.ComponentNotFound,
			expectedMessage: "Component not found",
		},
		{
			name: "CheckPurl error returns ComponentNotFound",
			input: Component{
				Purl:        "pkg:npm/error-pkg",
				Requirement: "1.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{}, errors.New("component not found")
				},
			},
			expectedStatus:  domain.ComponentNotFound,
			expectedMessage: "Component not found",
		},
		{
			name: "GetComponent error returns VersionNotFound",
			input: Component{
				Purl:        "pkg:npm/lodash",
				Requirement: ">=99.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{}, services.ErrVersionNotFound
				},
			},
			expectedStatus:  domain.VersionNotFound,
			expectedMessage: "Component version not found",
		},
		{
			name: "Negative CheckPurl count returns ComponentNotFound",
			input: Component{
				Purl:        "pkg:npm/negative-count",
				Requirement: "1.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{}, errors.New("component not found")
				},
			},
			expectedStatus:  domain.ComponentNotFound,
			expectedMessage: "Component not found",
		},
		{
			name: "GetComponent returns empty response preserves empty version",
			input: Component{
				Purl:        "pkg:npm/empty-response",
				Requirement: "^1.0.0",
				Status:      domain.ComponentStatus{StatusCode: domain.Success},
			},
			mock: &mockComponentResolver{
				getComponentFn: func(_ context.Context, _ types.ComponentRequest) (types.ComponentResponse, error) {
					return types.ComponentResponse{}, nil
				},
			},
			expectedStatus:  domain.Success,
			expectedVersion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan Component, 1)
			results := make(chan Component, 1)
			wg := sync.WaitGroup{}
			wg.Add(1)

			go componentVersionWorker(ctx, s, tt.mock, jobs, results, &wg)

			jobs <- tt.input
			close(jobs)

			wg.Wait()
			close(results)

			result := <-results
			if result.Status.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, result.Status.StatusCode)
			}
			if tt.expectedMessage != "" && result.Status.Message != tt.expectedMessage {
				t.Errorf("expected message %q, got %q", tt.expectedMessage, result.Status.Message)
			}
			if tt.expectedVersion != "" && result.Version != tt.expectedVersion {
				t.Errorf("expected version %q, got %q", tt.expectedVersion, result.Version)
			}
		})
	}
}
