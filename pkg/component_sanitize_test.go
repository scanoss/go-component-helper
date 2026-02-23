package pkg

import (
	"github.com/scanoss/go-grpc-helper/pkg/grpc/domain"
	"testing"
)

func TestSanitizeComponents(t *testing.T) {
	tests := []struct {
		name           string
		components     []ComponentDTO
		expectedStatus []domain.StatusCode
	}{
		{
			name: "All valid components",
			components: []ComponentDTO{
				{Purl: "pkg:npm/lodash@4.17.21"},
				{Purl: "pkg:github/scanoss/scanoss.js@1.0.0"},
			},
			expectedStatus: []domain.StatusCode{domain.Success, domain.Success},
		},
		{
			name: "Mixed valid and invalid components",
			components: []ComponentDTO{
				{Purl: "pkg:npm/lodash@4.17.21"},
				{Purl: "invalid-purl"},
			},
			expectedStatus: []domain.StatusCode{domain.Success, domain.InvalidPurl},
		},
		{
			name: "Component with empty requirement gets extracted from purl",
			components: []ComponentDTO{
				{Purl: "pkg:npm/lodash@4.17.21", Requirement: ""},
			},
			expectedStatus: []domain.StatusCode{domain.Success},
		},
		{
			name:           "All invalid components",
			components:     []ComponentDTO{{Purl: "invalid"}, {Purl: "also-invalid"}},
			expectedStatus: []domain.StatusCode{domain.InvalidPurl, domain.InvalidPurl},
		},
		{
			name:           "Invalid Purl with semver",
			components:     []ComponentDTO{{Purl: "pkg:npm/lodash@>=4.17.21"}},
			expectedStatus: []domain.StatusCode{domain.Success},
		},
		{
			name:           "Empty purl",
			components:     []ComponentDTO{{Purl: ""}},
			expectedStatus: []domain.StatusCode{domain.InvalidPurl},
		},
		{
			name:           "Empty components",
			components:     []ComponentDTO{},
			expectedStatus: []domain.StatusCode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeComponents(tt.components)

			if len(result) != len(tt.expectedStatus) {
				t.Fatalf("expected %d components, got %d", len(tt.expectedStatus), len(result))
			}

			for i, expectedCode := range tt.expectedStatus {
				if result[i].Status.StatusCode != expectedCode {
					t.Errorf("component[%d]: expected status %s, got %s", i, expectedCode, result[i].Status.StatusCode)
				}
			}
		})
	}
}
