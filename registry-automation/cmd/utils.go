package cmd

import (
	"encoding/json"
	"fmt"
	"os"
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

// Note: The location is relative to the root of the repository
func readFile(location string) ([]byte, error) {
	// Read the file

	fileBytes, err := os.ReadFile("../" + location)
	if err != nil {
		return fileBytes, fmt.Errorf("error reading file at location: %s %v", location, err)
	}

	return fileBytes, nil
}
