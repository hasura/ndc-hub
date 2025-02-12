package validate

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"os"
	"path/filepath"
	"sort"
)

type Connector struct {
	Namespace string
	Name      string
}

// validateLatestVersion checks if the latest version in metadata.json matches
// the actual latest version from the releases directory
func ValidateLatestVersion(connectorName string, connectorNamespace string, declaredLatestVersion string) error {
	// Check if releases directory exists
	releasesPath := fmt.Sprintf("../registry/%s/%s/releases", connectorNamespace, connectorName)

	// Read all version directories
	entries, err := os.ReadDir(releasesPath)
	if err != nil {
		return fmt.Errorf("failed to read releases directory for connector %s/%s: %v",
			connectorNamespace, connectorName, err)
	}

	var versions semver.Collection
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if connector-packaging.json exists for this version
		packagingPath := filepath.Join(releasesPath, entry.Name(), "connector-packaging.json")
		if _, err := os.Stat(packagingPath); os.IsNotExist(err) {
			continue // Skip directories without connector-packaging.json
		}

		// Parse the version string, skipping invalid semver strings
		version, err := semver.NewVersion(entry.Name())
		if err != nil {
			continue // Skip invalid semver versions
		}
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no valid versions found in releases directory for connector %s/%s",
			connectorNamespace, connectorName)
	}

	// Sort versions
	sort.Sort(versions)
	actualLatestVersion := versions[len(versions)-1]

	// Parse the declared version for comparison
	declaredVersion, err := semver.NewVersion(declaredLatestVersion)
	if err != nil {
		return fmt.Errorf("invalid semver format for latest_version in metadata.json: %v", err)
	}

	if !actualLatestVersion.Equal(declaredVersion) {
		return fmt.Errorf("latest_version in metadata.json (%s) does not match actual latest version (%s) for connector %s/%s",
			declaredLatestVersion, actualLatestVersion, connectorNamespace, connectorName)
	}
	return nil
}

func ConnectorPackaging(cp *ndchub.ConnectorPackaging) error {
	// validate version field
	if err := checkVersion(cp.Version); err != nil {
		return err
	}

	// validate uri and checksum fields
	if err := checkConnectorTarball(cp); err != nil {
		return err
	}

	return nil
}

func checkVersion(version string) error {
	if !strings.HasPrefix(version, "v") {
		return fmt.Errorf("version must start with 'v': but got %s", version)
	}
	_, err := semver.NewVersion(version)
	if err != nil {
		return fmt.Errorf("invalid semantic version: %s", version)
	}
	return nil
}

func checkConnectorTarball(cp *ndchub.ConnectorPackaging) error {
	var checksumFuncs map[string]hash.Hash = map[string]hash.Hash{
		"sha256": sha256.New(),
	}

	fileContents, err := downloadFile(cp.URI)
	if err != nil {
		return err
	}

	hashFunc, ok := checksumFuncs[cp.Checksum.Type]
	if !ok {
		return fmt.Errorf("unsupported checksum type: %s", cp.Checksum.Type)
	}

	_, err = io.Copy(hashFunc, fileContents)
	if err != nil {
		return err
	}
	defer fileContents.Close()

	checksum := fmt.Sprintf("%x", hashFunc.Sum(nil))
	if checksum != cp.Checksum.Value {
		return fmt.Errorf("checksum mismatch: checksum of downloaded file: %s, but checksum in connector-packaging.json: %s", checksum, cp.Checksum.Value)
	}

	return nil
}

func downloadFile(uri string) (io.ReadCloser, error) {
	var err error

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading: status code %d", resp.StatusCode)
	}

	return resp.Body, nil
}
