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
			name:          "versions with pre-release tags should compare against stable version",
			latestVersion: "1.0.0",
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
		{
			name:          "stable version with newer beta version should not error",
			latestVersion: "1.2.16",
			allVersions:   []string{"1.2.15", "1.2.16", "1.2.17-beta.1"},
			wantErr:       false,
		},
		{
			name:          "stable version with newer alpha version should not error",
			latestVersion: "1.1.3",
			allVersions:   []string{"1.1.2", "1.1.3", "1.1.4-alpha.1"},
			wantErr:       false,
		},
		{
			name:          "only pre-release versions should not error",
			latestVersion: "1.0.0-beta.1",
			allVersions:   []string{"1.0.0-alpha.1", "1.0.0-beta.1", "1.0.0-rc.1"},
			wantErr:       false,
		},
		{
			name:          "mixed stable and pre-release versions",
			latestVersion: "2.0.0",
			allVersions:   []string{"1.9.0", "2.0.0", "2.1.0-beta.1", "2.0.1-alpha.1"},
			wantErr:       false,
		},
		{
			name:          "postgres-promptql case: v1.2.16 with v1.2.17-beta.1",
			latestVersion: "v1.2.16",
			allVersions:   []string{"v1.2.15", "v1.2.16", "v1.2.17-beta.1"},
			wantErr:       false,
		},
		{
			name:          "snowflake case: v1.1.3 with v1.1.4-beta.1",
			latestVersion: "v1.1.3",
			allVersions:   []string{"v1.1.2", "v1.1.3", "v1.1.4-beta.1"},
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
