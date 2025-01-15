package validate

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestCheckConnectorTarball(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("dummy content"))
	}))
	defer server.Close()

	tests := []struct {
		name        string
		cp          *ndchub.ConnectorPackaging
		expectError bool
	}{
		{
			name: "Valid tarball",
			cp: &ndchub.ConnectorPackaging{
				URI: server.URL,
				Checksum: ndchub.Checksum{
					Type:  "sha256",
					Value: fmt.Sprintf("%x", sha256.Sum256([]byte("dummy content"))),
				},
			},
			expectError: false,
		},
		{
			name: "Invalid checksum",
			cp: &ndchub.ConnectorPackaging{
				URI: server.URL,
				Checksum: ndchub.Checksum{
					Type:  "sha256",
					Value: "invalid_checksum",
				},
			},
			expectError: true,
		},
		{
			name: "Unsupported checksum type",
			cp: &ndchub.ConnectorPackaging{
				URI: server.URL,
				Checksum: ndchub.Checksum{
					Type:  "md5",
					Value: "some_value",
				},
			},
			expectError: true,
		},
		{
			name: "Invalid URI",
			cp: &ndchub.ConnectorPackaging{
				URI: "invalid_url",
				Checksum: ndchub.Checksum{
					Type:  "sha256",
					Value: "some_value",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkConnectorTarball(tt.cp)
			if (err != nil) != tt.expectError {
				t.Errorf("checkConnectorTarball() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
