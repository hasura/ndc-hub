/*
Copyright © 2024 Hasura
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/machinebox/graphql"
	"github.com/spf13/cobra"
)

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run the CI workflow for hub registry publication",
	Run:   runCI,
}

type ChangedFiles struct {
	Added []string `json:"added_files"`
	Modified []string `json:"modified_files"`
	Deleted []string `json:"deleted_files"`
}

func init() {
	rootCmd.AddCommand(ciCmd)

	// Path for the changed files in the PR
	var changedFilesPathEnv = os.Getenv(`CHANGED_FILES_PATH`)
	ciCmd.PersistentFlags().String("changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		ciCmd.MarkPersistentFlagRequired("changed-files-path")
	}

	// Location of the registry files
	// var rfpE = os.Getenv(`CONNECTOR_HUB_DIRECTORY`)
	// ciCmd.PersistentFlags().String("connector-hub-directory", rfpE, "path to the connector hub checkout directory")
	// if rfpE == "" {
	// 	ciCmd.MarkPersistentFlagRequired("connector-hub-directory")
	// }

	// TODO: Check
	rand.Seed(time.Now().UnixNano())
}

func runCI(cmd *cobra.Command, args []string) {

	// For each connector where a change is detected...

	// var changed_files = map[string]bool{}
	// var changed_files_path = cmd.PersistentFlags().Lookup("changed-files-path").Value.String()
	// file, err := os.Open(changed_files_path)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()

	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	changed_files[filepath.Dir(scanner.Text())] = true // Just assume we'll treat each connector as if everything has changed.
	// }

	// if err := scanner.Err(); err != nil {
	// 	panic(err)
	// }

	// var hub_directory = cmd.PersistentFlags().Lookup("connector-hub-directory").Value.String()

	// for k := range changed_files {
	// 	respondToChangedConnector(path.Join(hub_directory, k))
	// }

	var changed_files_path = cmd.PersistentFlags().Lookup("changed-files-path").Value.String()
	changedFilesContent, err := os.Open(changed_files_path)

	if err!= nil {
		log.Fatalf("Failed to open the file: %v, err: %v", changed_files_path, err)
	}

	defer changedFilesContent.Close()

	// Read the file's contents
	changedFilesByteValue, err := ioutil.ReadAll(changedFilesContent)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	var changedFiles ChangedFiles
	err = json.Unmarshal(changedFilesByteValue, &changedFiles)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	fmt.Printf("Parsed JSON: \n%+v\n", changedFiles)



}

func respondToChangedConnector(changed_connector_path string) {
	// Detect status - added/modified/removed files
	// for each removed connector, remove from registry db
	// for each added connector, create a stub in the registry db
	// for each modified connector:
	//   * Download tgz
	//   * Re-upload tgz
	//   * Extract
	//   * Build payload for API
	//   * PUT to API (gql)

	const logo_file = "logo.png"
	const metadata_file = "metadata.json"
	const readme_file = "README.md"

	var logo_path = path.Join(changed_connector_path, logo_file)
	var metadata_path = path.Join(changed_connector_path, metadata_file)
	var readme_path = path.Join(changed_connector_path, readme_file)

	fmt.Println(changed_connector_path)
	fmt.Println(logo_path)
	fmt.Println(metadata_path)
	fmt.Println(readme_path)

	// Read the metadata
	var metadata_info = readJSONFile(metadata_path) // Read metadata file

	// Fetch, parse, and reupload the TGZ
	tgz_url := getStringFromPath([]string{"foo", "bar", "baz"}, metadata_info) // TODO // Get the url for the TGZ
	tgz_path := getTempFilePath("/tmp")
	downloadFile(tgz_url, tgz_path, map[string]string{}) // TODO: Auth headers
	extracted_tgz_path := "path/to/extract/directory"    // TODO
	err := extractTarGz(tgz_path, extracted_tgz_path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var tgz_info = readJSONFile(path.Join(extracted_tgz_path, "metadata.json")) // TODO: Check
	var tgz_new_url = "https://todo.com"
	uploadFile(tgz_path, tgz_new_url, map[string]string{})

	// Build payload for registry upsert
	// var logo_new_url = reuploadLogo(logo_path) // Is the logo hosted somewhere?
	var registry_payload = buildRegistryPayload(metadata_info, tgz_info, tgz_new_url)

	// Upsert
	updateRegistryGQL(registry_payload)
}

func buildRegistryPayload(metadata_info map[string]interface{}, tgz_info map[string]interface{}, tgz_url string) map[string]interface{} {
	return map[string]interface{}{} // TODO
}

func updateRegistryGQL(payload map[string]interface{}) {
	// Example: https://stackoverflow.com/questions/66931228/http-requests-golang-with-graphql

	client := graphql.NewClient("https://<GRAPHQL_API_HERE>")
	ctx := context.Background()

	req := graphql.NewRequest(`
    query ($key: String!) {
			items (id:$key) {
				field1
				field2
				field3
			}
    }
	`)

	req.Var("key", "value")

	var respData map[string]interface{}

	if err := client.Run(ctx, req, &respData); err != nil {
		panic(err)
	}

}

// Helper functions

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

func readJSONFile(location string) map[string]interface{} {
	// Read the file
	fileBytes, err := ioutil.ReadFile(location)
	if err != nil {
		panic(fmt.Errorf("error reading file: %v", err))
	}

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal(fileBytes, &data); err != nil {
		panic(fmt.Errorf("error parsing JSON: %v", err))
	}

	return data
}

func getStringFromPath(path []string, m map[string]interface{}) string {
	var current interface{} = m

	// Traverse the path
	for _, key := range path {
		// Check if current element is a map
		if currentMap, ok := current.(map[string]interface{}); ok {
			// Check if key exists in the current map
			if val, found := currentMap[key]; found {
				current = val
			} else {
				return "" // Key not found, return empty string
			}
		} else {
			return "" // Current element is not a map, return empty string
		}
	}

	// Check if the final value is a string
	if value, ok := current.(string); ok {
		return value
	}

	return "" // Final value is not a string, return empty string
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// generateRandomFileName generates a random file name based on the current time.
func generateRandomFileName() string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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

func uploadFile(sourceFilePath, destinationURL string, headers map[string]string) {
	// Open the file
	file, err := os.Open(sourceFilePath)
	if err != nil {
		panic(fmt.Errorf("error opening file: %v", err))
	}
	defer file.Close()

	// Create a new HTTP request with a POST method
	req, err := http.NewRequest("POST", destinationURL, file)
	if err != nil {
		panic(fmt.Errorf("error creating request: %v", err))
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Create an HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Errorf("error sending request: %v", err))
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		panic(fmt.Errorf("upload failed with status code: %d", resp.StatusCode))
	}
}

func extractTarGz(src, dest string) error {
	// Run the tar command with the -xzf options
	cmd := exec.Command("tar", "-xzf", src, "-C", dest)

	// Execute the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extracting tar.gz file: %v", err)
	}

	return nil
}
