/*
Copyright © 2024 Hasura
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"cloud.google.com/go/storage"
	"github.com/machinebox/graphql"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run the CI workflow for hub registry publication",
	Run:   runCI,
}

type ChangedFiles struct {
	Added    []string `json:"added_files"`
	Modified []string `json:"modified_files"`
	Deleted  []string `json:"deleted_files"`
}

// ConnectorVersion represents a version of a connector, this type is
// used to insert a new version of a connector in the registry.
type ConnectorVersion struct {
	// Namespace of the connector, e.g. "hasura"
	Namespace string `graphql:"namespace"`
	// Name of the connector
	Name string `graphql:"name"`
	// Semantic version of the connector
	Version string `graphql:"version"`
	// Docker image of the connector version
	Image string `graphql:"image"`
	// URL to the connector's metadata
	PackageDefinitionURL string `graphql:"package_definition_url"`
	// Is the connector version multitenant?
	IsMultitenant bool `graphql:"is_multitenant"`
	// Type of the connector packaing `PreBuiltDockerImage`/`ManagedDockerBuild`
	Type string `graphql:"type"`
}

type ConnectionVersionMetadata struct {
	Type  string `yaml:"type"`
	Image string `yaml: "image", omitempty`
}

const (
	ManagedDockerBuild  = "ManagedDockerBuild"
	PrebuiltDockerImage = "PrebuiltDockerImage"
)

func init() {
	rootCmd.AddCommand(ciCmd)

	// Path for the changed files in the PR
	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH")
	ciCmd.PersistentFlags().String("changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		ciCmd.MarkPersistentFlagRequired("changed-files-path")
	}

	// Publication environment
	var publicationEnv = os.Getenv("PUBLICATION_ENV")
	ciCmd.PersistentFlags().String("publication-env", publicationEnv, "publication environment (staging/prod). Default: staging")
	// default publicationEnv to "staging"
	if publicationEnv == "" {
		ciCmd.PersistentFlags().Set("publication-env", "staging")
	}
}

// processAddedOrModifiedConnectorVersions processes the files in the PR and extracts the connector name and version
func processAddedOrModifiedConnectorVersions(files []string, addedOrModifiedConnectorVersions map[string]map[string]string, re *regexp.Regexp) {
	for _, file := range files {
		matches := re.FindStringSubmatch(file)
		if len(matches) == 3 {
			connectorName := matches[1]
			connectorVersion := matches[2]

			if _, exists := addedOrModifiedConnectorVersions[connectorName]; !exists {
				addedOrModifiedConnectorVersions[connectorName] = make(map[string]string)
			}

			addedOrModifiedConnectorVersions[connectorName][connectorVersion] = file
		}
	}

}

// runCI is the main function that runs the CI workflow
func runCI(cmd *cobra.Command, args []string) {
	var changed_files_path = cmd.PersistentFlags().Lookup("changed-files-path").Value.String()
	changedFilesContent, err := os.Open(changed_files_path)

	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", changed_files_path, err)
	}

	defer changedFilesContent.Close()

	// Read the changed file's contents. This file contains all the changed files in the PR
	changedFilesByteValue, err := io.ReadAll(changedFilesContent)
	if err != nil {
		log.Fatalf("Failed to read the changed files JSON file: %v", err)
	}

	var changedFiles ChangedFiles
	err = json.Unmarshal(changedFilesByteValue, &changedFiles)
	if err != nil {
		log.Fatalf("Failed to unmarshal the changed files content: %v", err)

	}

	log.Printf("Changed files: %+v", changedFiles)

	// Collect the added or modified connectors
	addedOrModifiedConnectorVersions := collectAddedOrModifiedConnectors(changedFiles)

	for connectorName, versions := range addedOrModifiedConnectorVersions {
		for version, connectorVersionPath := range versions {
			err := respondToAddedOrModifiedConnectorVersion(connectorName, version, connectorVersionPath)
			if err != nil {
				log.Fatalf("Error while processing version and connector: %s - %s, Error: %v", version, connectorName, err)
			}
		}
	}

	fmt.Printf("new connector versions: \n%+v\n", addedOrModifiedConnectorVersions)

}

func collectAddedOrModifiedConnectors(changedFiles ChangedFiles) map[string]map[string]string {
	const connectorVersionPackageRegex = `^registry/([^/]+)/releases/([^/]+)/connector-packaging\.json$`
	re := regexp.MustCompile(connectorVersionPackageRegex) //

	addedOrModifiedConnectorVersions := make(map[string]map[string]string)

	processAddedOrModifiedConnectorVersions(changedFiles.Added, addedOrModifiedConnectorVersions, re)
	processAddedOrModifiedConnectorVersions(changedFiles.Modified, addedOrModifiedConnectorVersions, re)

	return addedOrModifiedConnectorVersions
}

func respondToAddedOrModifiedConnectorVersion(connectorName string, connectorVersion string, changedConnectorVersionPath string) error {
	// // Detect status - added/modified/removed files
	// // for each added connector, create a stub in the registry db
	// // for each modified connector:
	// //   * Download tgz
	// //   * Re-upload tgz
	// //   * Extract
	// //   * Build payload for API
	// //   * PUT to API (gql)

	// Read the connector packaging file
	ctx := context.Background()

	// connector's `metadata.json`, `registry/mongodb/metadata.json`
	connectorMetadata, err := readJSONFile[map[string]interface{}](fmt.Sprintf("registry/%s/metadata.json", connectorName))
	if err != nil {
		return fmt.Errorf("failed to read the connector metadata file: %v", err)
	}

	// connector version's metadata, `registry/mongodb/releases/v1.0.0/connector-packaging.json`
	connectorVersionPackagingInfo, err := readJSONFile[map[string]interface{}](changedConnectorVersionPath) // Read metadata file
	if err != nil {
		return fmt.Errorf("failed to read the connector packaging file: %v", err)
	}

	// Fetch, parse, and reupload the TGZ
	tgzUrl, ok := connectorVersionPackagingInfo["uri"].(string)

	// Check if the TGZ URL is valid
	if !ok || tgzUrl == "" {
		return fmt.Errorf("invalid or undefined TGZ URL: %v", tgzUrl)
	}

	connectorVersionMetadata, connectorMetadataTgzPath, err := getConnectorVersionMetadata(err, tgzUrl, connectorName, connectorVersion)
	if err != nil {
		return fmt.Errorf("failed to get connector version metadata: %v", err)
	}

	uploadedTgzUrl, err := uploadConnectorVersionDefinition(ctx, connectorName, connectorVersion, connectorMetadataTgzPath)
	if err != nil {
		return fmt.Errorf("failed to upload the connector version definition - connector: %v version:%v - err: %v", connectorName, connectorVersion, err)
	}

	connectorMetadataType, ok := connectorVersionMetadata["type"].(string)
	if !ok && (connectorMetadataType == ManagedDockerBuild || connectorMetadataType == PrebuiltDockerImage) {
		return fmt.Errorf("invalid or undefined connector type: %v", connectorMetadataType)
	}

	// // Build payload for registry upsert
	// var logo_new_url = reuploadLogo(logo_path) // Is the logo hosted somewhere?
	registryPayload, err := buildRegistryPayload(connectorName, connectorVersion, connectorMetadataType, connectorMetadata, uploadedTgzUrl)

	if err != nil {
		return fmt.Errorf("failed to build registry payload	: %v", err)
	}

	fmt.Printf("\n\n ************************ Registry payload: %+v", registryPayload)

	// // Upsert
	// updateRegistryGQL(registry_payload)
	return nil
}

func uploadConnectorVersionDefinition(ctx context.Context, connectorName string, connectorVersion string, connectorMetadataTgzPath string) (string, error) {
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("gcp-service-account-detail.json"))
	if err != nil {

		return "", fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	bucketName := "dev-connector-platform-registry"

	var uploadedTgzUrl string

	objectName := fmt.Sprintf("packages/%s/%s/package.tgz", connectorName, connectorVersion)
	uploadedTgzUrl, err = uploadFile(ctx, client, bucketName, objectName, connectorMetadataTgzPath)

	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	} else {
		fmt.Println("Uploaded file to GCS:", uploadedTgzUrl)

	}
	return uploadedTgzUrl, nil
}

// Downloads the TGZ File from the URL specified by `tgzUrl`, extracts the TGZ file and returns the content of the
// connector-definition.yaml present in the .hasura-connector folder.
func getConnectorVersionMetadata(err error, tgzUrl string, connectorName string, connectorVersion string) (map[string]interface{}, string, error) {
	var connectorVersionMetadata map[string]interface{}
	tgzPath := getTempFilePath("extracted_tgz", ".tgz")

	err = downloadFile(tgzUrl, tgzPath, map[string]string{})
	if err != nil {
		return connectorVersionMetadata, "", fmt.Errorf("failed to download the connector version metadata file from the URL: %v - err: %v", tgzUrl, err)
	}

	extractedTgzFolderPath := "extracted_tgz"

	if _, err := os.Stat(extractedTgzFolderPath); os.IsNotExist(err) {
		err := os.Mkdir(extractedTgzFolderPath, 0755)
		if err != nil {
			return connectorVersionMetadata, "", fmt.Errorf("failed to read the connector version metadata file: %v", err)
		}
	}

	connectorVersionMetadataYamlFilePath, err := extractTarGz(tgzPath, extractedTgzFolderPath+"/"+connectorName+"/"+connectorVersion)
	if err != nil {
		return connectorVersionMetadata, "", fmt.Errorf("failed to read the connector version metadata file: %v", err)
	} else {
		fmt.Println("Extracted metadata file at :", connectorVersionMetadataYamlFilePath)
	}

	connectorVersionMetadata, err = readYAMLFile(connectorVersionMetadataYamlFilePath)
	if err != nil {
		return connectorVersionMetadata, "", fmt.Errorf("failed to read the connector version metadata file: %v", err)
	}
	return connectorVersionMetadata, tgzPath, nil
}

// Write a function that accepts a file path to a YAML file and returns
// the contents of the file as a map[string]interface{}.
// readYAMLFile accepts a file path to a YAML file and returns the contents of the file as a map[string]interface{}.
func readYAMLFile(filePath string) (map[string]interface{}, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Unmarshal the YAML contents into a map
	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return result, nil
}

func buildRegistryPayload(
	connectorName string,
	version string,
	packagingType string,
	connectorMetadata map[string]interface{},
	uploadedConnectorDefinitionTgzUrl string,
) (ConnectorVersion, error) {
	var connectorVersion ConnectorVersion
	connectorOverview, ok := connectorMetadata["overview"].(map[string]interface{})
	if !ok {
		return connectorVersion, fmt.Errorf("Could not find connector overview in the connector's metadata")
	}
	connectorNamespace, ok := connectorOverview["namespace"].(string)
	if !ok {
		return connectorVersion, fmt.Errorf("Could not find the 'namespace' of the connector in the connector's overview in the connector's metadata.json")
	}
	connectorVersion.Namespace = connectorNamespace
	connectorVersion.Name = connectorName
	connectorVersion.Version = version

	return connectorVersion, nil
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

	// add a new key value to req

	var respData map[string]interface{}

	if err := client.Run(ctx, req, &respData); err != nil {
		panic(err)
	}

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

// parseJSON accepts a JSON byte slice and returns the parsed value as the specified type.
func parseJSON[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return result, nil
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
func generateRandomFileName(fileExtension string) string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b) + fileExtension
}

// getTempFilePath generates a random file name in the specified directory.
func getTempFilePath(directory string, fileExtension string) string {
	// Ensure the directory exists
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		panic(fmt.Errorf("error creating directory: %v", err))
	}

	// Generate a random file name
	fileName := generateRandomFileName(fileExtension)

	// Create the file path
	filePath := filepath.Join(directory, fileName)

	// Check if the file already exists
	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		// File exists, generate a new name
		fileName = generateRandomFileName(fileExtension)
		filePath = filepath.Join(directory, fileName)
	}

	return filePath

}

// uploadFile uploads a file to Google Cloud Storage
// document this function with comments
func uploadFile(ctx context.Context, client *storage.Client, bucketName, objectName, filePath string) (string, error) {
	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	wc := object.NewWriter(ctx)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Return the public URL of the uploaded object.
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, objectName)

	fmt.Printf("File %s uploaded to bucket %s as %s and is available at %s.\n", filePath, bucketName, objectName, publicURL)
	return publicURL, nil
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

	log.Println("***** File path: ", filepath)

	return fmt.Sprintf("%s/.hasura-connector/connector-metadata.yaml", filepath), nil
}
