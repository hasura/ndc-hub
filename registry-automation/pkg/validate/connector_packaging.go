package validate

import (
	"fmt"
	"strings"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"golang.org/x/mod/semver"
)

func ConnectorPackaging(cp *ndchub.ConnectorPackaging) error {
	// validate version field
	if !strings.HasPrefix(cp.Version, "v") {
		return fmt.Errorf("version must start with 'v': but got %s", cp.Version)
	}
	if !semver.IsValid(cp.Version) {
		return fmt.Errorf("invalid semantic version: %s", cp.Version)
	}

	return nil
}
