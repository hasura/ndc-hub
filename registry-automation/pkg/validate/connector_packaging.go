package validate

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"golang.org/x/mod/semver"
)

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
	if !semver.IsValid(version) {
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
