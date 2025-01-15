package validate

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	// validate uri and checksum fields
	connectorTgzFile, err := os.CreateTemp("", "connector-*.tgz")
	if err != nil {
		return err
	}
	defer connectorTgzFile.Close()
	err = downloadFile(cp.URI, connectorTgzFile)
	if err != nil {
		return err
	}
	computeChecksum, ok := checksumFuncs[cp.Checksum.Type]
	if !ok {
		return fmt.Errorf("unsupported checksum type: %s", cp.Checksum.Type)
	}
	checksum, err := computeChecksum(connectorTgzFile)
	if err != nil {
		return err
	}
	if checksum != cp.Checksum.Value {
		return fmt.Errorf("checksum mismatch: checksum of downloaded file: %s, but checksum in connector-packaging.json: %s", checksum, cp.Checksum.Value)
	}

	return nil
}

func downloadFile(uri string, destFile *os.File) error {
	var err error

	log.Println("starting download: ", uri)
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading: status code %d", resp.StatusCode)
	}

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

var checksumFuncs map[string]func(*os.File) (string, error) = map[string]func(*os.File) (string, error){
	"sha256": getSHA256,
}

func getSHA256(file *os.File) (string, error) {
	hash := sha256.New()

	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	checksum := hash.Sum(nil)
	return fmt.Sprintf("%x", checksum), nil
}
