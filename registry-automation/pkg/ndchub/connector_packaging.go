package ndchub

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const (
	ConnectorPackagingJSON = "connector-packaging.json"
)

type Checksum struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Source struct {
	Hash string `json:"hash"`
}

type Test struct {
	TestConfigPath string `json:"test_config_path"`
}

func (cp *ConnectorPackaging) GetTestConfigPath() string {
	if cp.Test.TestConfigPath == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(cp.Path), cp.Test.TestConfigPath)
}

type ConnectorPackaging struct {
	Namespace string `json:"-"`
	Name      string `json:"-"`
	Path      string `json:"-"`

	Version  string   `json:"version"`
	URI      string   `json:"uri"`
	Checksum Checksum `json:"checksum"`
	Source   Source   `json:"source"`
	Test     Test     `json:"test"`
}

func GetConnectorPackaging(path string) (*ConnectorPackaging, error) {
	if strings.Contains(path, "aliased_connectors") {
		// It should be safe to ignore aliased_connectors
		// as their slug is not used in the connector init process
		return nil, nil
	}

	// path looks like this: /some/folder/ndc-hub/registry/hasura/turso/releases/v0.1.0/connector-packaging.json
	versionFolder := filepath.Dir(path)
	releasesFolder := filepath.Dir(versionFolder)
	connectorFolder := filepath.Dir(releasesFolder)
	namespaceFolder := filepath.Dir(connectorFolder)

	connectorPackagingContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var connectorPackaging ConnectorPackaging
	err = json.Unmarshal(connectorPackagingContent, &connectorPackaging)
	if err != nil {
		return nil, err
	}
	connectorPackaging.Namespace = filepath.Base(namespaceFolder)
	connectorPackaging.Name = filepath.Base(connectorFolder)
	connectorPackaging.Path = path

	return &connectorPackaging, nil
}
