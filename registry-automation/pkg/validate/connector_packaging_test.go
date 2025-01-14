package validate

import (
	"testing"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
)

func TestConnectorPackaging(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"Valid version", "v1.0.0", false},
		{"Valid version with pre-release", "v1.0.0-alpha.1", false},
		{"Valid version with build metadata", "v1.0.0+build.1", false},
		{"Missing v prefix", "1.0.0", true},
		{"Empty version", "", true},
		{"Invalid characters", "vabc.1.0", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cp := &ndchub.ConnectorPackaging{
				Version: tc.version,
			}

			err := ConnectorPackaging(cp)

			if tc.wantErr && err == nil {
				t.Errorf("ConnectorPackaging() error = nil, wantErr %v", tc.wantErr)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("ConnectorPackaging() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
