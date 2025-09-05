package validate

import (
	"fmt"
	"sort"

	semver "github.com/Masterminds/semver/v3"
	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
)

func Metadata(cm *ndchub.ConnectorMetadata, connPkgs []ndchub.ConnectorPackaging) error {
	var connectorVersions []string
	for _, connPkg := range connPkgs {
		if connPkg.Namespace == cm.Overview.Namespace && connPkg.Name == cm.Overview.Name {
			connectorVersions = append(connectorVersions, connPkg.Version)
		}
	}
	if err := validateLatestVersion(cm.Overview.LatestVersion, connectorVersions); err != nil {
		return err
	}
	return nil
}

func validateLatestVersion(latestVersion string, allVersions []string) error {
	var stableVersions semver.Collection
	for _, version := range allVersions {
		v, err := semver.NewVersion(version)
		if err != nil {
			return fmt.Errorf("invalid semver format for version: %v", err)
		}
		// Only include stable versions (exclude pre-release versions like beta, alpha, rc, etc.)
		if v.Prerelease() == "" {
			stableVersions = append(stableVersions, v)
		}
	}

	// If no stable versions exist, return early without error
	if len(stableVersions) == 0 {
		return nil
	}

	sort.Sort(stableVersions)
	actualLatestVersion := stableVersions[len(stableVersions)-1]

	// Parse the declared version for comparison
	declaredVersion, err := semver.NewVersion(latestVersion)
	if err != nil {
		return fmt.Errorf("invalid semver format for latest_version in metadata.json: %v", err)
	}

	if !actualLatestVersion.Equal(declaredVersion) {
		return fmt.Errorf("latest_version in metadata.json (%s) does not match actual latest version (%s)",
			latestVersion, actualLatestVersion)
	}
	return nil
}
