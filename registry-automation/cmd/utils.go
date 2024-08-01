package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func generateGCPObjectName(namespace, connectorName, version string) string {
	return fmt.Sprintf("packages/%s/%s/%s/package.tgz", namespace, connectorName, version)
}

func downloadFile(sourceURL, destination string, headers map[string]string) error {
	// Create a new HTTP client
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Create the destination file
	outFile, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("error creating destination file: %v", err)
	}
	defer outFile.Close()

	// Write the response body to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

// Reads a JSON file and attempts to parse the content of the file
// into the type T.
// Note: The location is relative to the root of the repository
func readJSONFile[T any](location string) (T, error) {
	// Read the file
	var result T
	fileBytes, err := os.ReadFile("../" + location)
	if err != nil {
		return result, fmt.Errorf("error reading file at location: %s %v", location, err)
	}

	if err := json.Unmarshal(fileBytes, &result); err != nil {
		return result, fmt.Errorf("error parsing JSON: %v", err)
	}

	return result, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateRandomFileName generates a random file name based on the current time.
func generateRandomFileName() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b) + ".tar.gz"
}

// getTempFilePath generates a random file name in the specified directory.
func getTempFilePath(directory string) string {
	// Ensure the directory exists
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("error creating directory: %v", err))
	}

	// Generate a random file name
	fileName := generateRandomFileName()

	// Create the file path
	filePath := filepath.Join(directory, fileName)

	// Check if the file already exists
	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		// File exists, generate a new name
		fileName = generateRandomFileName()
		filePath = filepath.Join(directory, fileName)
	}
	return filePath

}

func extractTarGz(src, dest string) (string, error) {
	// Create the destination directory
	// Get the present working directory
	pwd := os.Getenv("PWD")
	filepath := pwd + "/" + dest

	if err := os.MkdirAll(filepath, 0755); err != nil {
		return "", fmt.Errorf("error creating destination directory: %v", err)
	}
	// Run the tar command with the -xvzf options
	cmd := exec.Command("tar", "-xvzf", src, "-C", dest)

	// Execute the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error extracting tar.gz file: %v", err)
	}

	return fmt.Sprintf("%s/.hasura-connector/connector-metadata.yaml", filepath), nil
}
