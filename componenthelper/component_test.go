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
	"testing"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
)

func TestGetComponentsVersion(t *testing.T) {
	err := zlog.NewSugaredDevLogger()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a sugared logger", err)
	}
	defer zlog.SyncZap()
	ctx := ctxzap.ToContext(context.Background(), zlog.L)
	s := ctxzap.Extract(ctx).Sugar()

	tests := []struct {
		name           string
		input          []ComponentDTO
		expectedLen    int
		expectedStatus []domain.StatusCode
	}{
		{
			name:           "Empty input returns empty result",
			input:          []ComponentDTO{},
			expectedLen:    0,
			expectedStatus: nil,
		},
		{
			name:           "Empty purl returns InvalidPurl",
			input:          []ComponentDTO{{Purl: ""}},
			expectedLen:    1,
			expectedStatus: []domain.StatusCode{domain.InvalidPurl},
		},
		{
			name:           "Invalid purls skip DB and return InvalidPurl",
			input:          []ComponentDTO{{Purl: "pg:scanoss/scanner.c"}, {Purl: "pkgscanoss/scanner.c"}},
			expectedLen:    2,
			expectedStatus: []domain.StatusCode{domain.InvalidPurl, domain.InvalidPurl},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetComponentsVersion(ComponentVersionCfg{
				MaxWorkers: 2,
				Ctx:        ctx,
				S:          s,
				DB:         nil,
				Input:      tt.input,
			})

			if len(result) != tt.expectedLen {
				t.Fatalf("expected %d components, got %d", tt.expectedLen, len(result))
			}

			for i, expectedCode := range tt.expectedStatus {
				if result[i].Status.StatusCode != expectedCode {
					t.Errorf("component[%d]: expected status %s, got %s", i, expectedCode, result[i].Status.StatusCode)
				}
			}
		})
	}
}
