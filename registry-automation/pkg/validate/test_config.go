package validate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
)

func TestConfig(tc *ndchub.TestConfig) error {
	if tc.HubID == "" {
		return fmt.Errorf("hubID is required in test-config.json %v", tc.Path)
	}
	if tc.SnapshotsDir == "" {
		return fmt.Errorf("snapshots_dir is required in test-config.json %v", tc.Path)
	}
	tcDir := filepath.Dir(tc.Path)
	snapshotsDir := filepath.Join(tcDir, tc.SnapshotsDir)
	snapshotsInfo, err := os.Stat(snapshotsDir)
	if err != nil {
		return fmt.Errorf("error reading snapshots directory: %w", err)
	}
	if !snapshotsInfo.IsDir() {
		return fmt.Errorf("snapshots_dir %q must be a directory", snapshotsDir)
	}
	shapshots, err := os.ReadDir(snapshotsDir)
	if err != nil {
		return fmt.Errorf("error reading snapshots directory: %w", err)
	}
	if len(shapshots) == 0 {
		return fmt.Errorf("snapshots_dir must contain at least one snapshot")
	}
	for _, snapshot := range shapshots {
		if !snapshot.IsDir() {
			return fmt.Errorf("snapshots_dir must contain only directories")
		}
		snapshotPath := filepath.Join(snapshotsDir, snapshot.Name())
		testFiles, err := os.ReadDir(snapshotPath)
		if err != nil {
			return fmt.Errorf("error reading snapshot directory: %w", err)
		}
		isRequestFilePresent, isResponseFilePresent := false, false
		for _, file := range testFiles {
			if file.Name() == "request.graphql" {
				isRequestFilePresent = true
			}
			if file.Name() == "response.json" {
				isResponseFilePresent = true
			}
		}
		if !isRequestFilePresent || !isResponseFilePresent {
			return fmt.Errorf("snapshot %q must contain request.graphql and response.json files", snapshotPath)
		}
	}
	if tc.SetupComposeFilePath != nil {
		setupComposeFile := filepath.Join(tc.Path, *tc.SetupComposeFilePath)
		setupComposeInfo, err := os.Stat(setupComposeFile)
		if err != nil {
			return fmt.Errorf("error reading setup compose file %q: %w", setupComposeFile, err)
		}
		if setupComposeInfo.IsDir() {
			return fmt.Errorf("setup compose file %q must be a file, not a directory", setupComposeFile)
		}
	}
	return nil
}
