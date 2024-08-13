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
	// Type of the connector packaging `PrebuiltDockerImage`/`ManagedDockerBuild`
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

var ciCmdArgs ConnectorRegistryArgs

func init() {
	rootCmd.AddCommand(ciCmd)

	// Path for the changed files in the PR
	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH")
	ciCmd.PersistentFlags().StringVar(&ciCmdArgs.ChangedFilesPath, "changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		ciCmd.MarkPersistentFlagRequired("changed-files-path")
	}

	// Publication environment
	var publicationEnv = os.Getenv("PUBLICATION_ENV")
	ciCmd.PersistentFlags().StringVar(&ciCmdArgs.PublicationEnv, "publication-env", publicationEnv, "publication environment (staging/prod). Default: staging")
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
		ciCmdArgs.ConnectorRegistryGQLUrl = registryGQLURL
	}

	// Connector publication key
	connectorPublicationKey := os.Getenv("CONNECTOR_PUBLICATION_KEY")
	if connectorPublicationKey == "" {
		log.Fatalf("CONNECTOR_PUBLICATION_KEY is not set")
	} else {
		ciCmdArgs.ConnectorPublicationKey = connectorPublicationKey
	}

	// GCP service account details
	gcpServiceAccountDetails := os.Getenv("GCP_SERVICE_ACCOUNT_DETAILS")
	if gcpServiceAccountDetails == "" {
		log.Fatalf("GCP_SERVICE_ACCOUNT_DETAILS is not set")
	} else {
		ciCmdArgs.GCPServiceAccountDetails = gcpServiceAccountDetails
	}

	// GCP bucket name
	gcpBucketName := os.Getenv("GCP_BUCKET_NAME")
	if gcpBucketName == "" {
		log.Fatalf("GCP_BUCKET_NAME is not set")
	} else {
		ciCmdArgs.GCPBucketName = gcpBucketName
	}
}

// processChangedFiles processes the files in the PR and extracts the connector name and version
// This function checks for the following things:
// 1. If a new connector version is added, it adds the connector version to the `newlyAddedConnectorVersions` map.
// 2. If the logo file is modified, it adds the connector name and the path to the modified logo to the `modifiedLogos` map.
// 3. If the README file is modified, it adds the connector name and the path to the modified README to the `modifiedReadmes` map.
func processChangedFiles(changedFiles ChangedFiles) NewConnectorVersions {

	newlyAddedConnectorVersions := make(map[Connector]map[string]string)

	var connectorVersionPackageRegex = regexp.MustCompile(`^registry/([^/]+)/([^/]+)/releases/([^/]+)/connector-packaging\.json$`)

	files := append(changedFiles.Added, changedFiles.Modified...)

	for _, file := range files {
		// Extract the connector name and version from the file path
		if connectorVersionPackageRegex.MatchString(file) {

			matches := connectorVersionPackageRegex.FindStringSubmatch(file)
			if len(matches) == 4 {
				connectorNamespace := matches[1]
				connectorName := matches[2]
				connectorVersion := matches[3]

				connector := Connector{
					Name:      connectorName,
					Namespace: connectorNamespace,
				}

				if _, exists := newlyAddedConnectorVersions[connector]; !exists {
					newlyAddedConnectorVersions[connector] = make(map[string]string)
				}

				newlyAddedConnectorVersions[connector][connectorVersion] = file
			}

		} else {
			fmt.Println("Skipping file: ", file)
		}
	}

	return newlyAddedConnectorVersions
}

// runCI is the main function that runs the CI workflow
func runCI(cmd *cobra.Command, args []string) {
	buildContext()
	changedFilesContent, err := os.Open(ciCmdArgs.ChangedFilesPath)
	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", ciCmdArgs.ChangedFilesPath, err)
	}
	defer changedFilesContent.Close()

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(ciCmdArgs.GCPServiceAccountDetails)))
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
	addedOrModifiedConnectorVersions := processChangedFiles(changedFiles)
	// check if the map is empty
	if len(addedOrModifiedConnectorVersions) == 0 {
		fmt.Println("No connector versions found in the changed files.")
		return
	} else {
		processNewlyAddedConnectorVersions(client, addedOrModifiedConnectorVersions)
	}
}

func processNewlyAddedConnectorVersions(client *storage.Client, newlyAddedConnectorVersions NewConnectorVersions) {
	// Iterate over the added or modified connectors and upload the connector versions
	var connectorVersions []ConnectorVersion
	var uploadConnectorVersionErr error
	encounteredError := false

	for connectorName, versions := range newlyAddedConnectorVersions {
		for version, connectorVersionPath := range versions {
			var connectorVersion ConnectorVersion
			connectorVersion, uploadConnectorVersionErr = uploadConnectorVersionPackage(client, connectorName, version, connectorVersionPath)

			if uploadConnectorVersionErr != nil {
				encounteredError = true
				break

			} else {
				connectorVersions = append(connectorVersions, connectorVersion)
			}

		}

		if encounteredError {
			// attempt to cleanup the uploaded connector versions
			_ = cleanupUploadedConnectorVersions(client, connectorVersions) // ignore errors while cleaning up
			// delete the uploaded connector versions from the registry
			log.Fatalf("Failed to upload the connector version: %v", uploadConnectorVersionErr)

		} else {
			fmt.Printf("Connector versions to be added to the registry: %+v\n", connectorVersions)
			err := updateRegistryGQL(connectorVersions)
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
		err := deleteFile(client, ciCmdArgs.GCPBucketName, objectName)
		if err != nil {
			return err
		}
	}
	return nil
}

// Type that uniquely identifies a connector
type Connector struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type NewConnectorVersions map[Connector]map[string]string

// uploadConnectorVersionPackage uploads the connector version package to the registry
func uploadConnectorVersionPackage(client *storage.Client, connector Connector, version string, changedConnectorVersionPath string) (ConnectorVersion, error) {

	var connectorVersion ConnectorVersion

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

	connectorVersionMetadata, connectorMetadataTgzPath, err := getConnectorVersionMetadata(tgzUrl, connector, version)
	if err != nil {
		return connectorVersion, err
	}

	uploadedTgzUrl, err := uploadConnectorVersionDefinition(client, connector.Name, connector.Namespace, version, connectorMetadataTgzPath)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to upload the connector version definition - connector: %v version:%v - err: %v", connector.Name, version, err)
	} else {
		// print success message with the name of the connector and the version
		fmt.Printf("Successfully uploaded the connector version definition in google cloud registry for the connector: %v version: %v\n", connector.Name, version)
	}

	// Build payload for registry upsert
	return buildRegistryPayload(connector.Namespace, connector.Name, version, connectorVersionMetadata, uploadedTgzUrl)
}

func uploadConnectorVersionDefinition(client *storage.Client, connectorNamespace, connectorName string, connectorVersion string, connectorMetadataTgzPath string) (string, error) {
	bucketName := ciCmdArgs.GCPBucketName
	objectName := generateGCPObjectName(connectorNamespace, connectorName, connectorVersion)
	uploadedTgzUrl, err := uploadFile(client, bucketName, objectName, connectorMetadataTgzPath)

	if err != nil {
		return "", err
	}
	return uploadedTgzUrl, nil
}

// Downloads the TGZ File from the URL specified by `tgzUrl`, extracts the TGZ file and returns the content of the
// connector-definition.yaml present in the .hasura-connector folder.
func getConnectorVersionMetadata(tgzUrl string, connector Connector, connectorVersion string) (map[string]interface{}, string, error) {
	var connectorVersionMetadata map[string]interface{}
	tgzPath, err := getTempFilePath("extracted_tgz")
	if err != nil {
		return connectorVersionMetadata, "", fmt.Errorf("failed to get the temp file path: %v", err)
	}
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

	connectorVersionMetadataYamlFilePath, err := extractTarGz(tgzPath, extractedTgzFolderPath+"/"+connector.Namespace+"/"+connector.Name+"/"+connectorVersion)
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
	client := graphql.NewClient(ciCmdArgs.ConnectorRegistryGQLUrl)
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
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

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

	connectorInfo, err := getConnectorInfoFromRegistry(connectorNamespace, connectorName)

	if err != nil {
		return connectorVersion, err
	}

	// Check if the connector exists in the registry first
	if len(connectorInfo.HubRegistryConnector) == 0 {
		return connectorVersion, fmt.Errorf("Inserting a new connector is not supported yet")
	}

	var connectorVersionType string

	if connectorVersionPackagingType == PrebuiltDockerImage {
		// Note: The connector version type is set to `PreBuiltDockerImage` if the connector version is of type `PrebuiltDockerImage`, this is a HACK because this value might be removed in the future and we might not even need to insert new connector versions in the `hub_registry_connector_version` table.
		connectorVersionType = "PreBuiltDockerImage"
	} else {
		connectorVersionType = ManagedDockerBuild
	}

	connectorVersion = ConnectorVersion{
		Namespace:            connectorNamespace,
		Name:                 connectorName,
		Version:              version,
		Image:                &connectorVersionDockerImage,
		PackageDefinitionURL: uploadedConnectorDefinitionTgzUrl,
		IsMultitenant:        connectorInfo.HubRegistryConnector[0].MultitenantConnector != nil,
		Type:                 connectorVersionType,
	}

	return connectorVersion, nil
}

func updateRegistryGQL(payload []ConnectorVersion) error {
	var respData map[string]interface{}
	client := graphql.NewClient(ciCmdArgs.ConnectorRegistryGQLUrl)
	ctx := context.Background()

	req := graphql.NewRequest(`
mutation InsertConnectorVersion($connectorVersion: [hub_registry_connector_version_insert_input!]!) {
  insert_hub_registry_connector_version(objects: $connectorVersion, on_conflict: {constraint: connector_version_namespace_name_version_key, update_columns: [image, package_definition_url, is_multitenant]}) {
    affected_rows
    returning {
      id
    }
  }
}`)
	// add the payload to the request
	req.Var("connectorVersion", payload)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil
}
