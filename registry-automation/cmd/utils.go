package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func generateGCPObjectName(namespace, connectorName, version string) string {
	return fmt.Sprintf("packages/%s/%s/%s/package.tgz", namespace, connectorName, version)
}

// Reads a JSON file and attempts to parse the content of the file
// into the type T.
// Note: The location is relative to the root of the repository
func readJSONFile[T any](location string) (T, error) {
	// Read the file
	var result T
	fileBytes, err := readFile(location)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(fileBytes, &result); err != nil {
		return result, fmt.Errorf("error parsing JSON: %v", err)
	}

	return result, nil
}

func checkedBundleFile(location string, maxSize int64) (string, error) {
	if !filepath.IsAbs(stagingInputDir) || !filepath.IsLocal(location) || filepath.Clean(location) != location {
		return "", errors.New("invalid staging input path")
	}
	path := stagingInputDir
	root, err := os.Lstat(path)
	if err != nil || !root.IsDir() || root.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("invalid staging input root")
	}
	var info os.FileInfo
	for _, component := range strings.Split(location, string(filepath.Separator)) {
		path = filepath.Join(path, component)
		info, err = os.Lstat(path)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("missing or invalid staging input")
		}
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || info.Size() > maxSize {
		return "", errors.New("invalid staging input size or file type")
	}
	return path, nil
}

// Note: The location is relative to the root of the repository.
func readFile(location string) ([]byte, error) {
	if stagingInputDir != "" {
		if !strings.HasPrefix(location, "registry/") {
			return nil, errors.New("only registry data can be read from the staging bundle")
		}
		path, err := checkedBundleFile(location, 2*1024*1024)
		if err != nil {
			return nil, err
		}
		return os.ReadFile(path)
	}
	// Read the file
	var fileBytes []byte
	var err error

	if filepath.IsAbs(location) {
		fileBytes, err = os.ReadFile(location)
	} else {
		fileBytes, err = os.ReadFile("../" + location) // previous behavior; written for backwards compatibility
	}
	if err != nil {
		return fileBytes, fmt.Errorf("error reading file at location: %s %v", location, err)
	}

	return fileBytes, nil
}
