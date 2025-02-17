package validate

import (
	"strings"
	"testing"
)

func TestValidateLatestVersion(t *testing.T) {
	tests := []struct {
		name          string
		latestVersion string
		allVersions   []string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "valid versions with correct latest",
			latestVersion: "1.2.3",
			allVersions:   []string{"1.0.0", "1.1.0", "1.2.3"},
			wantErr:       false,
		},
		{
			name:          "valid versions with incorrect latest",
			latestVersion: "1.1.0",
			allVersions:   []string{"1.0.0", "1.1.0", "1.2.3"},
			wantErr:       true,
			errContains:   "latest_version in metadata.json (1.1.0) does not match actual latest version (1.2.3)",
		},
		{
			name:          "valid versions with latest version not in allVersions",
			latestVersion: "1.2.4",
			allVersions:   []string{"1.0.0", "1.1.0", "1.2.3"},
			wantErr:       true,
			errContains:   "latest_version in metadata.json (1.2.4) does not match actual latest version (1.2.3)",
		},

		{
			name:          "invalid latest version format",
			latestVersion: "invalid.version",
			allVersions:   []string{"1.0.0", "1.1.0"},
			wantErr:       true,
			errContains:   "invalid semver format for latest_version in metadata.json",
		},
		{
			name:          "invalid version in allVersions",
			latestVersion: "1.0.0",
			allVersions:   []string{"1.0.0", "invalid.version"},
			wantErr:       true,
			errContains:   "invalid semver format for version",
		},
		{
			name:          "versions with pre-release tags",
			latestVersion: "2.0.0-beta",
			allVersions:   []string{"1.0.0", "2.0.0-alpha", "2.0.0-beta"},
			wantErr:       false,
		},
		{
			name:          "single version",
			latestVersion: "1.0.0",
			allVersions:   []string{"1.0.0"},
			wantErr:       false,
		},
		{
			name:          "versions with build metadata",
			latestVersion: "1.2.3+build.123",
			allVersions:   []string{"1.0.0", "1.1.0+build.111", "1.2.3+build.123"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLatestVersion(tt.latestVersion, tt.allVersions)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
