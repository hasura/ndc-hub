package validate

import (
	"fmt"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"golang.org/x/mod/semver"
)

func ConnectorPackaging(cp *ndchub.ConnectorPackaging) error {
	if !semver.IsValid(cp.Version) {
		return fmt.Errorf("invalid semantic version: %s", cp.Version)
	}

	return nil
}
