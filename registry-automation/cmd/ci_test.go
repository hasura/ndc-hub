package cmd

import (
	"testing"
)

func TestProcessAddedOrModifiedConnectorVersions(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name                              string
		files                             []string
		expectedAddedOrModifiedConnectors map[string]map[string]string
	}{
		{
			name: "Test case 1",
			files: []string{
				"registry/hasura/releases/v1.0.0/connector-packaging.json",
				"registry/hasura/releases/v2.0.0/connector-packaging.json",
				"registry/other/releases/v1.0.0/connector-packaging.json",
			},
			expectedAddedOrModifiedConnectors: map[string]map[string]string{
				"hasura": {
					"v1.0.0": "registry/hasura/releases/v1.0.0/connector-packaging.json",
					"v2.0.0": "registry/hasura/releases/v2.0.0/connector-packaging.json",
				},
				"other": {
					"v1.0.0": "registry/other/releases/v1.0.0/connector-packaging.json",
				},
			},
		},
		{
			name: "Test case 2",
			files: []string{
				"registry/hasura/releases/v1.0.0/connector-packaging.json",
				"registry/hasura/releases/v1.0.0/other-file.json",
			},
			expectedAddedOrModifiedConnectors: map[string]map[string]string{
				"hasura": {
					"v1.0.0": "registry/hasura/releases/v1.0.0/connector-packaging.json",
				},
			},
		},
		{
			name: "Test case 3",
			files: []string{
				"registry/hasura/releases/v1.0.0/other-file.json",
				"registry/other/releases/v1.0.0/connector-packaging.json",
			},
			expectedAddedOrModifiedConnectors: map[string]map[string]string{
				"other": {
					"v1.0.0": "registry/other/releases/v1.0.0/connector-packaging.json",
				},
			},
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Initialize the map to store the added or modified connectors
			addedOrModifiedConnectorVersions := make(map[string]map[string]string)

			var changedFiles ChangedFiles

			changedFiles.Added = tc.files

			// Call the function under test
			processAddedOrModifiedConnectorVersions(changedFiles)

			// Compare the actual result with the expected result
			if len(addedOrModifiedConnectorVersions) != len(tc.expectedAddedOrModifiedConnectors) {
				t.Errorf("Unexpected number of connectors. Expected: %d, Got: %d", len(tc.expectedAddedOrModifiedConnectors), len(addedOrModifiedConnectorVersions))
			}

			for connectorName, versions := range addedOrModifiedConnectorVersions {
				expectedVersions, ok := tc.expectedAddedOrModifiedConnectors[connectorName]
				if !ok {
					t.Errorf("Unexpected connector name: %s", connectorName)
					continue
				}

				if len(versions) != len(expectedVersions) {
					t.Errorf("Unexpected number of versions for connector %s. Expected: %d, Got: %d", connectorName, len(expectedVersions), len(versions))
				}

				for version, connectorVersionPath := range versions {
					expectedPath, ok := expectedVersions[version]
					if !ok {
						t.Errorf("Unexpected version for connector %s: %s", connectorName, version)
						continue
					}

					if connectorVersionPath != expectedPath {
						t.Errorf("Unexpected connector version path for connector %s, version %s. Expected: %s, Got: %s", connectorName, version, expectedPath, connectorVersionPath)
					}
				}
			}
		})
	}
}
