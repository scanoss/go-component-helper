package pkg

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	zlog "github.com/scanoss/zap-logging-helper/pkg/logger"
	"testing"
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
