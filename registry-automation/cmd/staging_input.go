package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var stagingPackagePath = regexp.MustCompile(`^registry/[a-z0-9][a-z0-9_-]{0,77}/[a-z0-9][a-z0-9_-]{0,77}/releases/v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.+-]+)?/connector-packaging\.json$`)

// The workflow verified archive bytes and metadata before acquiring credentials.
// The publisher reads files as data and never fetches or extracts a package.
func readStagingPackage(path string) (map[string]interface{}, string, error) {
	if !stagingPackagePath.MatchString(path) {
		return nil, "", errors.New("invalid staging package path")
	}
	dir := strings.TrimSuffix(strings.TrimPrefix(path, "registry/"), "/connector-packaging.json")
	prefix := filepath.Join("packages", filepath.FromSlash(dir))
	meta, err := checkedBundleFile(filepath.Join(prefix, "connector-metadata.yaml"), 2*1024*1024)
	if err != nil {
		return nil, "", err
	}
	archive, err := checkedBundleFile(filepath.Join(prefix, "package.tgz"), 32*1024*1024)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(meta)
	if err != nil {
		return nil, "", err
	}
	var m map[string]interface{}
	if err := yaml.Unmarshal(data, &m); err != nil || m == nil {
		return nil, "", errors.New("invalid staged connector metadata")
	}
	return m, archive, nil
}
