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
	Namespace string `json:"namespace"`
	// Name of the connector
	Name string `json:"name"`
	// Semantic version of the connector
	Version string `json:"version"`
	// Docker image of the connector version
	Image *string `json:"image,omitempty"`
	// URL to the connector's metadata
	PackageDefinitionURL string `json:"package_definition_url"`
	// Is the connector version multitenant?
	IsMultitenant bool `json:"is_multitenant"`
	// Type of the connector packaing `PreBuiltDockerImage`/`ManagedDockerBuild`
	Type string `json:"type"`
}

// Create a struct with the following fields:
// type string
// image *string (optional)
type ConnectionVersionMetadata struct {
	Type  string  `yaml:"type"`
	Image *string `yaml:"image,omitempty"`
}

const (
	ManagedDockerBuild  = "ManagedDockerBuild"
	PrebuiltDockerImage = "PrebuiltDockerImage"
)

// Make a struct with the fields expected in the command line arguments
type ConnectorRegistryArgs struct {
	ChangedFilesPath         string
	PublicationEnv           string
	ConnectorRegistryGQL     string
	ConnectorPublicationKey  string
	GCPServiceAccountDetails string
	GCPBucketName            string
}

var cmdArgs ConnectorRegistryArgs

func init() {
	rootCmd.AddCommand(ciCmd)

	// Path for the changed files in the PR
	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH")
	ciCmd.PersistentFlags().StringVar(&cmdArgs.ChangedFilesPath, "changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		ciCmd.MarkPersistentFlagRequired("changed-files-path")
	}

	// Publication environment
	var publicationEnv = os.Getenv("PUBLICATION_ENV")
	ciCmd.PersistentFlags().StringVar(&cmdArgs.PublicationEnv, "publication-env", publicationEnv, "publication environment (staging/prod). Default: staging")
	// default publicationEnv to "staging"
	if publicationEnv == "" {
		ciCmd.PersistentFlags().Set("publication-env", "staging")
	}

	// Connector registry Hasura GraphQL URL
	registryGQLURL := os.Getenv("CONNECTOR_REGISTRY_GQL_URL")
	ciCmd.PersistentFlags().StringVar(&cmdArgs.ConnectorRegistryGQL, "connector-registry-gql-url", registryGQLURL, "Hasura GraphQL URL for the connector registry")
	if registryGQLURL == "" {
		ciCmd.MarkPersistentFlagRequired("connector-registry-gql-url")
	}

	// Connector publication key
	connectorPublicationKey := os.Getenv("CONNECTOR_PUBLICATION_KEY")
	ciCmd.PersistentFlags().StringVar(&cmdArgs.ConnectorPublicationKey, "connector-publication-key", connectorPublicationKey, "Connector publication key used for authentication with the registry GraphQL API")
	if connectorPublicationKey == "" {
		ciCmd.MarkPersistentFlagRequired("connector-publication-key")
	}

	// GCP service account details
	gcpServiceAccountDetails := os.Getenv("GCP_SERVICE_ACCOUNT_DETAILS")
	ciCmd.PersistentFlags().StringVar(&cmdArgs.GCPServiceAccountDetails, "gcp-service-account-details", gcpServiceAccountDetails, "GCP service account details file path")
	if gcpServiceAccountDetails == "" {
		ciCmd.MarkPersistentFlagRequired("gcp-service-account-details")
	}

	// GCP bucket name
	gcpBucketName := os.Getenv("GCP_BUCKET_NAME")
	ciCmd.PersistentFlags().StringVar(&cmdArgs.GCPBucketName, "gcp-bucket-name", gcpBucketName, "GCP bucket name")
	if gcpBucketName == "" {
		ciCmd.MarkPersistentFlagRequired("gcp-bucket-name")
	}

}

// processAddedOrModifiedConnectorVersions processes the files in the PR and extracts the connector name and version
func processAddedOrModifiedConnectorVersions(files []string, addedOrModifiedConnectorVersions map[string]map[string]string) {
	const connectorVersionPackageRegex = `^registry/([^/]+)/releases/([^/]+)/connector-packaging\.json$`
	re := regexp.MustCompile(connectorVersionPackageRegex)

	for _, file := range files {
		// Extract the connector name and version from the file path

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
	ctx := context.Background()
	changedFilesContent, err := os.Open(cmdArgs.ChangedFilesPath)

	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", cmdArgs.ChangedFilesPath, err)
	}

	defer changedFilesContent.Close()

	client, err := storage.NewClient(ctx, option.WithCredentialsFile("gcp-service-account-detail.json"))
	if err != nil {
		log.Fatalf("Failed to create Google bucket client: %v", err)
	}
	defer client.Close()

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

	// Collect the added or modified connectors
	addedOrModifiedConnectorVersions := collectAddedOrModifiedConnectors(changedFiles)

	var connectorVersions []ConnectorVersion
	var uploadConnectorVersionErr error
	encounteredError := false

	for connectorName, versions := range addedOrModifiedConnectorVersions {
		for version, connectorVersionPath := range versions {
			var connectorVersion ConnectorVersion
			connectorVersion, uploadConnectorVersionErr = uploadConnectorVersionPackage(client, connectorName, version, connectorVersionPath)
			if err != nil {
				fmt.Printf("Error while processing version and connector: %s - %s, Error: %v", version, connectorName, err)
				encounteredError = true
				break
			}
			connectorVersions = append(connectorVersions, connectorVersion)
		}
		if encounteredError {
			break
		}
	}

	if encounteredError {
		// attempt to cleanup the uploaded connector versions
		_ = cleanupUploadedConnectorVersions(client, connectorVersions) // ignore errors while cleaning up
		// delete the uploaded connector versions from the registry
		log.Fatalf("Failed to upload the connector version: %v", uploadConnectorVersionErr)
		// TODO: Delete the uploaded connector versions from the registry

	} else {
		fmt.Println("Successfully uploaded the connector versions to the registry")
		err = updateRegistryGQL(connectorVersions)
		if err != nil {
			// attempt to cleanup the uploaded connector versions
			_ = cleanupUploadedConnectorVersions(client, connectorVersions) // ignore errors while cleaning up

			log.Fatalf("Failed to update the registry: %v", err)
		}
	}
}

func cleanupUploadedConnectorVersions(client *storage.Client, connectorVersions []ConnectorVersion) error {
	// Iterate over the connector versions and delete the uploaded files
	// from the google bucket

	for _, connectorVersion := range connectorVersions {
		objectName := generateGCPObjectName(connectorVersion.Name, connectorVersion.Version)
		err := deleteFile(client, "dev-connector-platform-registry", objectName)
		if err != nil {
			return err
		}
	}
	return nil
}

// collectAddedOrModifiedConnectors collects the added or modified connectors from the changed files
func collectAddedOrModifiedConnectors(changedFiles ChangedFiles) map[string]map[string]string {

	addedOrModifiedConnectorVersions := make(map[string]map[string]string)

	processAddedOrModifiedConnectorVersions(changedFiles.Added, addedOrModifiedConnectorVersions)

	// Not sure if we need to process the modified files as well, because it is very unlikely
	// that an existing connector version will be modified.

	// processAddedOrModifiedConnectorVersions(changedFiles.Modified, addedOrModifiedConnectorVersions)

	return addedOrModifiedConnectorVersions
}

// uploadConnectorVersionPackage uploads the connector version package to the registry
func uploadConnectorVersionPackage(client *storage.Client, connectorName string, version string, changedConnectorVersionPath string) (ConnectorVersion, error) {

	var connectorVersion ConnectorVersion

	// Read the connector's metadata and the connector version's metadata

	// connector's `metadata.json`, `registry/mongodb/metadata.json`
	connectorMetadata, err := readJSONFile[map[string]interface{}](fmt.Sprintf("registry/%s/metadata.json", connectorName))
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to read the connector metadata file: %v", err)
	}

	// connector version's metadata, `registry/mongodb/releases/v1.0.0/connector-packaging.json`
	connectorVersionPackagingInfo, err := readJSONFile[map[string]interface{}](changedConnectorVersionPath) // Read metadata file
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to read the connector packaging file: %v", err)
	}
	// Fetch, parse, and reupload the TGZ
	tgzUrl, ok := connectorVersionPackagingInfo["uri"].(string)

	// Check if the TGZ URL is valid
	if !ok || tgzUrl == "" {
		return connectorVersion, fmt.Errorf("invalid or undefined TGZ URL: %v", tgzUrl)
	}

	connectorVersionMetadata, connectorMetadataTgzPath, err := getConnectorVersionMetadata(err, tgzUrl, connectorName, version)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to get connector version metadata: %v", err)
	}

	uploadedTgzUrl, err := uploadConnectorVersionDefinition(client, connectorName, version, connectorMetadataTgzPath)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to upload the connector version definition - connector: %v version:%v - err: %v", connectorName, version, err)
	} else {
		// print success message with the name of the connector and the version
		fmt.Printf("Successfully uploaded the connector version definition in google cloud registry for the connector: %v version: %v\n", connectorName, version)
	}

	// Build payload for registry upsert
	return buildRegistryPayload(connectorName, version, connectorVersionMetadata, connectorMetadata, uploadedTgzUrl)
}

func uploadConnectorVersionDefinition(client *storage.Client, connectorName string, connectorVersion string, connectorMetadataTgzPath string) (string, error) {
	bucketName := "dev-connector-platform-registry"
	objectName := generateGCPObjectName(connectorName, connectorVersion)
	uploadedTgzUrl, err := uploadFile(client, bucketName, objectName, connectorMetadataTgzPath)

	if err != nil {
		return "", err
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

// buildRegistryPayload builds the payload for the registry upsert API
func buildRegistryPayload(
	connectorName string,
	version string,
	connectorVersionMetadata map[string]interface{},
	connectorMetadata map[string]interface{},
	uploadedConnectorDefinitionTgzUrl string,
) (ConnectorVersion, error) {
	var connectorVersion ConnectorVersion
	var connectorVersionDockerImage string

	connectorOverview, ok := connectorMetadata["overview"].(map[string]interface{})
	if !ok {
		return connectorVersion, fmt.Errorf("could not find connector overview in the connector's metadata")
	}
	connectorNamespace, ok := connectorOverview["namespace"].(string)
	if !ok {
		return connectorVersion, fmt.Errorf("could not find the 'namespace' of the connector in the connector's overview in the connector's metadata.json")
	}

	connectorVersionPackagingDefinition, ok := connectorVersionMetadata["packagingDefinition"].(map[interface{}]interface{})
	if !ok {
		return connectorVersion, fmt.Errorf("could not find the 'packagingDefinition' of the connector %s version %s in the connector's metadata", connectorName, version)
	}
	connectorVersionPackagingType, ok := connectorVersionPackagingDefinition["type"].(string)

	if !ok && (connectorVersionPackagingType == ManagedDockerBuild || connectorVersionPackagingType == PrebuiltDockerImage) {
		return connectorVersion, fmt.Errorf("invalid or undefined connector type: %v", connectorVersionPackagingDefinition)
	} else if connectorVersionPackagingType == PrebuiltDockerImage {
		connectorVersionDockerImage, ok = connectorVersionPackagingDefinition["dockerImage"].(string)
		if !ok {
			return connectorVersion, fmt.Errorf("could not find the 'dockerImage' of the PrebuiltDockerImage connector %s version %s in the connector's metadata", connectorName, version)
		}
	}

	// TODO: Make a query to the registry to check if the connector already exists,
	// if not, insert the connector first and then insert the connector version.
	// Also, fetch the is_multitenant value from the registry.
	connectorVersion = ConnectorVersion{
		Namespace:            connectorNamespace,
		Name:                 connectorName,
		Version:              version,
		Image:                &connectorVersionDockerImage,
		PackageDefinitionURL: uploadedConnectorDefinitionTgzUrl,
		IsMultitenant:        false, // TODO(KC): Figure this out.
		Type:                 connectorVersionPackagingType,
	}

	return connectorVersion, nil
}

func updateRegistryGQL(payload []ConnectorVersion) error {
	client := graphql.NewClient("http://localhost:8081/v1/graphql")
	ctx := context.Background()

	req := graphql.NewRequest(`
mutation InsertConnectorVersion($connectorVersion: [hub_registry_connector_version_insert_input!]!) {
  insert_hub_registry_connector_version(objects: $connectorVersion, on_conflict: {constraint: connector_version_namespace_name_version_key, update_columns: [image, package_definition_url]}) {
    affected_rows
    returning {
      id
    }
  }
}`)
	// add the payload to the request
	req.Var("connectorVersion", payload)

	// add a new key value to req

	var respData map[string]interface{}

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", "usnEu*pYp8wiUjbzv3g4iruemTzDgfi@") // TODO: The value of the header should be fetched from the environment variable CONNECTOR_PUBLICATION_KEY

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	// print the respData
	fmt.Println("Response from the API: ", respData)

	return nil

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
