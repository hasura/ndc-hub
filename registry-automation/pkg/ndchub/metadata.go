package ndchub

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	MetadataJSON = "metadata.json"
)

// Type to represent the metadata.json file
type ConnectorMetadata struct {
	Overview struct {
		Namespace     string   `json:"namespace"`
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		Title         string   `json:"title"`
		Logo          string   `json:"logo"`
		Tags          []string `json:"tags"`
		LatestVersion string   `json:"latest_version"`
	} `json:"overview"`
	Author struct {
		SupportEmail string `json:"support_email"`
		Homepage     string `json:"homepage"`
		Name         string `json:"name"`
	} `json:"author"`
	IsVerified         bool `json:"is_verified"`
	IsHostedByHasura   bool `json:"is_hosted_by_hasura"`
	HasuraHubConnector struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	} `json:"hasura_hub_connector"`
	SourceCode struct {
		IsOpenSource bool   `json:"is_open_source"`
		Repository   string `json:"repository"`
	} `json:"source_code"`
}

func GetConnectorMetadata(path string) (*ConnectorMetadata, error) {
	if strings.Contains(path, "aliased_connectors") {
		// It should be safe to ignore aliased_connectors
		// as their slug is not used in the connector init process
		return nil, nil
	}

	connectorPackagingContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cm ConnectorMetadata
	err = json.Unmarshal(connectorPackagingContent, &cm)
	if err != nil {
		return nil, err
	}

	if cm.Overview.Name == "" {
		// path looks like this: /some/folder/ndc-hub/registry/hasura/turso/metadata.json
		connectorFolder := filepath.Dir(path)
		cm.Overview.Name = filepath.Base(connectorFolder)
	}

	return &cm, nil
}
