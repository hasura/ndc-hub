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
	"os"
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
	// Name of the connector, e.g. "mongodb"
	Name string `json:"name"`
	// Semantic version of the connector version, e.g. "v1.0.0"
	Version string `json:"version"`
	// Docker image of the connector version (optional)
	// This field is only required if the connector version is of type `PrebuiltDockerImage`
	Image *string `json:"image,omitempty"`
	// URL to the connector's metadata
	PackageDefinitionURL string `json:"package_definition_url"`
	// Is the connector version multitenant?
	IsMultitenant bool `json:"is_multitenant"`
	// Type of the connector packaging `PreBuiltDockerImage`/`ManagedDockerBuild`
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
	ConnectorRegistryGQLUrl  string
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

}

func buildContext() {
	// Connector registry Hasura GraphQL URL
	registryGQLURL := os.Getenv("CONNECTOR_REGISTRY_GQL_URL")
	if registryGQLURL == "" {
		log.Fatalf("CONNECTOR_REGISTRY_GQL_URL is not set")
	} else {
		cmdArgs.ConnectorRegistryGQLUrl = registryGQLURL
	}

	// Connector publication key
	connectorPublicationKey := os.Getenv("CONNECTOR_PUBLICATION_KEY")
	if connectorPublicationKey == "" {
		log.Fatalf("CONNECTOR_PUBLICATION_KEY is not set")
	} else {
		cmdArgs.ConnectorPublicationKey = connectorPublicationKey
	}

	// GCP service account details
	gcpServiceAccountDetails := os.Getenv("GCP_SERVICE_ACCOUNT_DETAILS")
	if gcpServiceAccountDetails == "" {
		log.Fatalf("GCP_SERVICE_ACCOUNT_DETAILS is not set")
	} else {
		cmdArgs.GCPServiceAccountDetails = gcpServiceAccountDetails
	}

	// GCP bucket name
	gcpBucketName := os.Getenv("GCP_BUCKET_NAME")
	if gcpBucketName == "" {
		log.Fatalf("GCP_BUCKET_NAME is not set")
	} else {
		cmdArgs.GCPBucketName = gcpBucketName
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
	buildContext()
	changedFilesContent, err := os.Open(cmdArgs.ChangedFilesPath)

	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", cmdArgs.ChangedFilesPath, err)
	}

	defer changedFilesContent.Close()

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(cmdArgs.GCPServiceAccountDetails)))
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

	// check if the map is empty
	if len(addedOrModifiedConnectorVersions) == 0 {
		fmt.Println("No connector versions found in the changed files.")
		return
	} else {
		// Iterate over the added or modified connectors and upload the connector versions
		var connectorVersions []ConnectorVersion
		var uploadConnectorVersionErr error
		encounteredError := false

		for connectorName, versions := range addedOrModifiedConnectorVersions {
			for version, connectorVersionPath := range versions {
				var connectorVersion ConnectorVersion
				connectorVersion, uploadConnectorVersionErr = uploadConnectorVersionPackage(client, connectorName, version, connectorVersionPath)

				if uploadConnectorVersionErr != nil {
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

		} else {
			fmt.Printf("Connector versions to be added to the registry: %+v\n", connectorVersions)
			err = updateRegistryGQL(connectorVersions)
			if err != nil {
				// attempt to cleanup the uploaded connector versions
				_ = cleanupUploadedConnectorVersions(client, connectorVersions) // ignore errors while cleaning up
				log.Fatalf("Failed to update the registry: %v", err)
			}
		}

		fmt.Println("Successfully added connector versions to the registry.")
	}
}

func cleanupUploadedConnectorVersions(client *storage.Client, connectorVersions []ConnectorVersion) error {
	// Iterate over the connector versions and delete the uploaded files
	// from the google bucket
	fmt.Println("Cleaning up the uploaded connector versions")

	for _, connectorVersion := range connectorVersions {
		objectName := generateGCPObjectName(connectorVersion.Namespace, connectorVersion.Name, connectorVersion.Version)
		err := deleteFile(client, cmdArgs.GCPBucketName, objectName)
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
		return connectorVersion, err
	}

	connectorNamespace, err := getConnectorNamespace(connectorMetadata)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to get the connector namespace: %v", err)
	}

	uploadedTgzUrl, err := uploadConnectorVersionDefinition(client, connectorNamespace, connectorName, version, connectorMetadataTgzPath)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to upload the connector version definition - connector: %v version:%v - err: %v", connectorName, version, err)
	} else {
		// print success message with the name of the connector and the version
		fmt.Printf("Successfully uploaded the connector version definition in google cloud registry for the connector: %v version: %v\n", connectorName, version)
	}

	// Build payload for registry upsert
	return buildRegistryPayload(connectorNamespace, connectorName, version, connectorVersionMetadata, uploadedTgzUrl)
}

func uploadConnectorVersionDefinition(client *storage.Client, connectorNamespace, connectorName string, connectorVersion string, connectorMetadataTgzPath string) (string, error) {
	bucketName := cmdArgs.GCPBucketName
	objectName := generateGCPObjectName(connectorNamespace, connectorName, connectorVersion)
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
	tgzPath := getTempFilePath("extracted_tgz")

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

func getConnectorNamespace(connectorMetadata map[string]interface{}) (string, error) {
	connectorOverview, ok := connectorMetadata["overview"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("could not find connector overview in the connector's metadata")
	}
	connectorNamespace, ok := connectorOverview["namespace"].(string)
	if !ok {
		return "", fmt.Errorf("could not find the 'namespace' of the connector in the connector's overview in the connector's metadata.json")
	}
	return connectorNamespace, nil
}

// struct to store the response of teh GetConnectorInfo query
type GetConnectorInfoResponse struct {
	HubRegistryConnector []struct {
		Name                 string `json:"name"`
		MultitenantConnector *struct {
			ID string `json:"id"`
		} `json:"multitenant_connector"`
	} `json:"hub_registry_connector"`
}

func getConnectorInfoFromRegistry(connectorNamespace string, connectorName string) (GetConnectorInfoResponse, error) {
	var respData GetConnectorInfoResponse
	client := graphql.NewClient(cmdArgs.ConnectorRegistryGQLUrl)
	ctx := context.Background()

	req := graphql.NewRequest(`
query GetConnectorInfo ($name: String!, $namespace: String!) {
  hub_registry_connector(where: {_and: [{name: {_eq: $name}}, {namespace: {_eq: $namespace}}]}) {
    name
    multitenant_connector {
      id
    }
  }
}`)
	req.Var("name", connectorName)
	req.Var("namespace", connectorNamespace)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", cmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, err
	} else {
		if len(respData.HubRegistryConnector) == 0 {
			return respData, nil
		}
	}

	return respData, nil
}

// buildRegistryPayload builds the payload for the registry upsert API
func buildRegistryPayload(
	connectorNamespace string,
	connectorName string,
	version string,
	connectorVersionMetadata map[string]interface{},
	uploadedConnectorDefinitionTgzUrl string,
) (ConnectorVersion, error) {
	var connectorVersion ConnectorVersion
	var connectorVersionDockerImage string
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

	// query GetConnectorInfo ($name: String!, $namespace: String!) {
	// 	hub_registry_connector(where: {_and: [{name: {_eq: $name}}, {namespace: {_eq: $namespace}}]}) {
	// 		name
	// 		multitenant_connector {
	// 			id
	// 		}
	// 	}
	// }

	connectorInfo, err := getConnectorInfoFromRegistry(connectorNamespace, connectorName)

	if err != nil {
		return connectorVersion, err
	}

	// Check if the connector exists in the registry first
	if len(connectorInfo.HubRegistryConnector) == 0 {
		return connectorVersion, fmt.Errorf("Inserting a new connector is not supported yet")
	}

	connectorVersion = ConnectorVersion{
		Namespace:            connectorNamespace,
		Name:                 connectorName,
		Version:              version,
		Image:                &connectorVersionDockerImage,
		PackageDefinitionURL: uploadedConnectorDefinitionTgzUrl,
		IsMultitenant:        connectorInfo.HubRegistryConnector[0].MultitenantConnector != nil,
		Type:                 connectorVersionPackagingType,
	}

	return connectorVersion, nil
}

func updateRegistryGQL(payload []ConnectorVersion) error {
	var respData map[string]interface{}
	client := graphql.NewClient(cmdArgs.ConnectorRegistryGQLUrl)
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

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", cmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil
}
