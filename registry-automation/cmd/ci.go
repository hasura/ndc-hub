package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"cloud.google.com/go/storage"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/machinebox/graphql"
	"github.com/spf13/cobra"
	"google.golang.org/api/option"
)

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run the CI workflow for hub registry publication",
	Run:   runCI,
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

func buildContext() Context {
	// Connector registry Hasura GraphQL URL
	registryGQLURL := os.Getenv("CONNECTOR_REGISTRY_GQL_URL")
	var registryGQLClient *graphql.Client
	var storageClient *storage.Client
	var cloudinaryClient *cloudinary.Cloudinary
	if registryGQLURL == "" {
		log.Fatalf("CONNECTOR_REGISTRY_GQL_URL is not set")
	} else {
		ciCmdArgs.ConnectorRegistryGQLUrl = registryGQLURL
		registryGQLClient = graphql.NewClient(registryGQLURL)
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
		var err error
		storageClient, err = storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(ciCmdArgs.GCPServiceAccountDetails)))
		if err != nil {
			log.Fatalf("Failed to create Google bucket client: %v", err)
		}
		defer storageClient.Close()

		ciCmdArgs.GCPServiceAccountDetails = gcpServiceAccountDetails
	}

	// GCP bucket name
	gcpBucketName := os.Getenv("GCP_BUCKET_NAME")
	if gcpBucketName == "" {
		log.Fatalf("GCP_BUCKET_NAME is not set")
	} else {
		ciCmdArgs.GCPBucketName = gcpBucketName
	}

	cloudinaryUrl := os.Getenv("CLOUDINARY_URL")

	if cloudinaryUrl == "" {
		log.Fatalf("CLOUDINARY_URL is not set")
	} else {
		var err error
		cloudinaryClient, err = cloudinary.NewFromURL(cloudinaryUrl)
		if err != nil {
			log.Fatalf("Failed to create cloudinary client: %v", err)

		}

	}

	return Context{
		Env:               ciCmdArgs.PublicationEnv,
		RegistryGQLClient: registryGQLClient,
		StorageClient:     storageClient,
		Cloudinary:        cloudinaryClient,
	}

}

type fileProcessor struct {
	regex   *regexp.Regexp
	process func(matches []string, file string)
}

// processChangedFiles categorizes changes in connector files within a registry system.
// It handles new and modified files including metadata, logos, READMEs, and connector versions.
//
// The function takes a ChangedFiles struct containing slices of added and modified filenames,
// and returns a ProcessedChangedFiles struct with categorized changes.
//
// Files are processed based on their path and type:
//   - metadata.json: New connectors
//   - logo.(png|svg): New or modified logos
//   - README.md: New or modified READMEs
//   - connector-packaging.json: New connector versions
//
// Any files not matching these patterns are logged as skipped.
//
// Example usage:
//
//	changedFiles := ChangedFiles{
//		Added: []string{"registry/namespace1/connector1/metadata.json"},
//		Modified: []string{"registry/namespace2/connector2/README.md"},
//	}
//	result := processChangedFiles(changedFiles)
func processChangedFiles(changedFiles ChangedFiles) ProcessedChangedFiles {
	result := ProcessedChangedFiles{
		NewConnectorVersions: make(map[Connector]map[string]string),
		ModifiedLogos:        make(map[Connector]string),
		ModifiedReadmes:      make(map[Connector]string),
		NewConnectors:        make(map[NewConnector]MetadataFile),
		NewLogos:             make(map[Connector]string),
		NewReadmes:           make(map[Connector]string),
	}

	processors := []fileProcessor{
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/metadata.json$`),
			process: func(matches []string, file string) {
				connector := NewConnector{Name: matches[2], Namespace: matches[1]}
				result.NewConnectors[connector] = MetadataFile(file)
				fmt.Printf("Processing metadata file for connector: %s\n", connector.Name)
			},
		},
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/logo\.(png|svg)$`),
			process: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.NewLogos[connector] = file
				fmt.Printf("Processing logo file for connector: %s\n", connector.Name)
			},
		},
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/README\.md$`),
			process: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.NewReadmes[connector] = file
				fmt.Printf("Processing README file for connector: %s\n", connector.Name)
			},
		},
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/releases/([^/]+)/connector-packaging\.json$`),
			process: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				version := matches[3]
				if _, exists := result.NewConnectorVersions[connector]; !exists {
					result.NewConnectorVersions[connector] = make(map[string]string)
				}
				result.NewConnectorVersions[connector][version] = file
			},
		},
	}

	processFile := func(file string, isModified bool) {
		for _, processor := range processors {
			if matches := processor.regex.FindStringSubmatch(file); matches != nil {
				if isModified {
					connector := Connector{Name: matches[2], Namespace: matches[1]}
					if processor.regex.String() == processors[1].regex.String() {
						result.ModifiedLogos[connector] = file
					} else if processor.regex.String() == processors[2].regex.String() {
						result.ModifiedReadmes[connector] = file
					}
				} else {
					processor.process(matches, file)
				}
				return
			}
		}
		fmt.Printf("Skipping %s file: %s\n", map[bool]string{true: "modified", false: "newly added"}[isModified], file)
	}

	for _, file := range changedFiles.Added {
		processFile(file, false)
	}

	for _, file := range changedFiles.Modified {
		processFile(file, true)
	}

	return result
}

func processNewConnector(ciCtx Context, connector NewConnector, metadataFile MetadataFile) (ConnectorOverviewInsert, HubRegistryConnectorInsertInput, error) {
	// Process the newly added connector
	// Get the string value from metadataFile
	var connectorOverviewAndAuthor ConnectorOverviewInsert
	var hubRegistryConnectorInsertInput HubRegistryConnectorInsertInput

	connectorMetadata, err := readJSONFile[ConnectorMetadata](string(metadataFile))
	if err != nil {
		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, fmt.Errorf("Failed to parse the connector metadata file: %v", err)
	}

	docs, err := readFile(fmt.Sprintf("registry/%s/%s/README.md", connector.Namespace, connector.Name))

	if err != nil {

		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, fmt.Errorf("Failed to read the README file of the connector: %s : %v", connector.Name, err)
	}

	logoPath := fmt.Sprintf("registry/%s/%s/logo.png", connector.Namespace, connector.Name)

	uploadedLogoUrl, err := uploadLogoToCloudinary(ciCtx.Cloudinary, Connector{Name: connector.Name, Namespace: connector.Namespace}, logoPath)
	if err != nil {
		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, err
	}

	// Get connector info from the registry
	connectorInfo, err := getConnectorInfoFromRegistry(*ciCtx.RegistryGQLClient, connector.Name, connector.Namespace)
	if err != nil {
		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput,
			fmt.Errorf("Failed to get the connector info from the registry: %v", err)
	}

	// Check if the connector already exists in the registry
	if len(connectorInfo.HubRegistryConnector) > 0 {
		if ciCtx.Env == "staging" {
			fmt.Printf("Connector already exists in the registry: %s/%s\n", connector.Namespace, connector.Name)
			fmt.Println("The connector is going to be overwritten in the registry.")

		} else {

			return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, fmt.Errorf("Attempting to create a new hub connector, but the connector already exists in the registry: %s/%s", connector.Namespace, connector.Name)
		}

	}

	hubRegistryConnectorInsertInput = HubRegistryConnectorInsertInput{
		Name:      connector.Name,
		Namespace: connector.Namespace,
		Title:     connectorMetadata.Overview.Title,
	}

	connectorOverviewAndAuthor = ConnectorOverviewInsert{
		Name:        connector.Name,
		Namespace:   connector.Namespace,
		Docs:        string(docs),
		Logo:        uploadedLogoUrl,
		Title:       connectorMetadata.Overview.Title,
		Description: connectorMetadata.Overview.Description,
		IsVerified:  connectorMetadata.IsVerified,
		IsHosted:    connectorMetadata.IsHostedByHasura,
		Author: struct {
			Data ConnectorAuthor `json:"data"`
		}{
			Data: ConnectorAuthor{
				Name:         connectorMetadata.Author.Name,
				SupportEmail: connectorMetadata.Author.SupportEmail,
				Website:      connectorMetadata.Author.Homepage,
			},
		},
	}

	return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, nil
}

// runCI is the main function that runs the CI workflow
func runCI(cmd *cobra.Command, args []string) {
	ctx := buildContext()
	changedFilesContent, err := os.Open(ciCmdArgs.ChangedFilesPath)
	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", ciCmdArgs.ChangedFilesPath, err)
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

	// Separate the modified files according to the type of file

	// Collect the added or modified connectors
	processChangedFiles := processChangedFiles(changedFiles)

	newlyAddedConnectorVersions := processChangedFiles.NewConnectorVersions
	modifiedLogos := processChangedFiles.ModifiedLogos
	modifiedReadmes := processChangedFiles.ModifiedReadmes

	newlyAddedConnectors := processChangedFiles.NewConnectors

	var newConnectorsToBeAdded NewConnectorsInsertInput
	var newConnectorVersionsToBeAdded []ConnectorVersion
	var connectorOverviewUpdates []ConnectorOverviewUpdate

	if len(newlyAddedConnectors) > 0 {
		fmt.Println("New connectors to be added to the registry: ", newlyAddedConnectors)
		newConnectorOverviewsToBeAdded := make([](ConnectorOverviewInsert), 0)
		hubRegistryConnectorsToBeAdded := make([](HubRegistryConnectorInsertInput), 0)

		for connector, metadataFile := range newlyAddedConnectors {
			connectorOverviewAndAuthor, hubRegistryConnector, err := processNewConnector(ctx, connector, metadataFile)

			if err != nil {
				log.Fatalf("Failed to process the new connector: %s/%s, Error: %v", connector.Namespace, connector.Name, err)
			}
			newConnectorOverviewsToBeAdded = append(newConnectorOverviewsToBeAdded, connectorOverviewAndAuthor)
			hubRegistryConnectorsToBeAdded = append(hubRegistryConnectorsToBeAdded, hubRegistryConnector)

		}

		newConnectorsToBeAdded.HubRegistryConnectors = hubRegistryConnectorsToBeAdded
		newConnectorsToBeAdded.ConnectorOverviews = newConnectorOverviewsToBeAdded

	}

	if len(newlyAddedConnectorVersions) > 0 {
		newConnectorVersionsToBeAdded = processNewlyAddedConnectorVersions(ctx, newlyAddedConnectorVersions)
	}

	if len(modifiedReadmes) > 0 {
		readMeUpdates, err := processModifiedReadmes(modifiedReadmes)
		if err != nil {
			log.Fatalf("Failed to process the modified READMEs: %v", err)
		}
		connectorOverviewUpdates = append(connectorOverviewUpdates, readMeUpdates...)
		fmt.Println("Successfully updated the READMEs in the registry.")
	}

	if len(modifiedLogos) > 0 {
		logoUpdates, err := processModifiedLogos(modifiedLogos)
		if err != nil {
			log.Fatalf("Failed to process the modified logos: %v", err)
		}
		connectorOverviewUpdates = append(connectorOverviewUpdates, logoUpdates...)
		fmt.Println("Successfully updated the logos in the registry.")
	}

	err = registryDbMutation(*ctx.RegistryGQLClient, newConnectorsToBeAdded, connectorOverviewUpdates, newConnectorVersionsToBeAdded)

	fmt.Println("Successfully processed the changed files in the PR")
}

func uploadLogoToCloudinary(cloudinary *cloudinary.Cloudinary, connector Connector, logoPath string) (string, error) {
	logoContent, err := readFile(logoPath)
	if err != nil {
		fmt.Printf("Failed to read the logo file: %v", err)
		return "", err
	}

	imageReader := bytes.NewReader(logoContent)

	uploadResult, err := cloudinary.Upload.Upload(context.Background(), imageReader, uploader.UploadParams{
		PublicID: fmt.Sprintf("%s-%s", connector.Namespace, connector.Name),
		Format:   "png",
	})
	if err != nil {
		return "", fmt.Errorf("Failed to upload the logo to cloudinary for the connector: %s, Error: %v\n", connector.Name, err)
	}
	return uploadResult.SecureURL, nil
}

func processModifiedLogos(modifiedLogos ModifiedLogos) ([]ConnectorOverviewUpdate, error) {
	// Iterate over the modified logos and update the logos in the registry
	var connectorOverviewUpdates []ConnectorOverviewUpdate
	// upload the logo to cloudinary
	cloudinary, err := cloudinary.NewFromURL(ciCmdArgs.CloudinaryUrl)
	if err != nil {
		return connectorOverviewUpdates, err
	}

	for connector, logoPath := range modifiedLogos {
		// open the logo file
		uploadedLogoUrl, err := uploadLogoToCloudinary(cloudinary, connector, logoPath)
		if err != nil {
			return connectorOverviewUpdates, err
		}

		var connectorOverviewUpdate ConnectorOverviewUpdate

		if connectorOverviewUpdate.Set.Logo == nil {
			connectorOverviewUpdate.Set.Logo = new(string)
		} else {
			*connectorOverviewUpdate.Set.Logo = ""
		}

		*connectorOverviewUpdate.Set.Logo = uploadedLogoUrl

		connectorOverviewUpdate.Where.ConnectorName = connector.Name
		connectorOverviewUpdate.Where.ConnectorNamespace = connector.Namespace

		connectorOverviewUpdates = append(connectorOverviewUpdates, connectorOverviewUpdate)

	}

	return connectorOverviewUpdates, nil

}

func processModifiedReadmes(modifiedReadmes ModifiedReadmes) ([]ConnectorOverviewUpdate, error) {
	// Iterate over the modified READMEs and update the READMEs in the registry
	var connectorOverviewUpdates []ConnectorOverviewUpdate

	for connector, readmePath := range modifiedReadmes {
		// open the README file
		readmeContent, err := readFile(readmePath)
		if err != nil {
			return connectorOverviewUpdates, err

		}

		readMeContentString := string(readmeContent)

		var connectorOverviewUpdate ConnectorOverviewUpdate
		connectorOverviewUpdate.Set.Docs = &readMeContentString

		connectorOverviewUpdate.Where.ConnectorName = connector.Name
		connectorOverviewUpdate.Where.ConnectorNamespace = connector.Namespace

		connectorOverviewUpdates = append(connectorOverviewUpdates, connectorOverviewUpdate)

	}

	return connectorOverviewUpdates, nil

}

func processNewlyAddedConnectorVersions(ciCtx Context, newlyAddedConnectorVersions NewConnectorVersions) []ConnectorVersion {
	// Iterate over the added or modified connectors and upload the connector versions
	var connectorVersions []ConnectorVersion
	var uploadConnectorVersionErr error
	encounteredError := false

	for connectorName, versions := range newlyAddedConnectorVersions {
		for version, connectorVersionPath := range versions {
			var connectorVersion ConnectorVersion
			connectorVersion, uploadConnectorVersionErr = uploadConnectorVersionPackage(ciCtx, connectorName, version, connectorVersionPath)

			if uploadConnectorVersionErr != nil {
				fmt.Printf("Error while processing version and connector: %s - %s, Error: %v", version, connectorName, uploadConnectorVersionErr)
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
		_ = cleanupUploadedConnectorVersions(ciCtx.StorageClient, connectorVersions) // ignore errors while cleaning up
		// delete the uploaded connector versions from the registry
		log.Fatalf("Failed to upload the connector version: %v", uploadConnectorVersionErr)
	}

	fmt.Println("Successfully added connector versions to the registry.")

	return connectorVersions

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

// uploadConnectorVersionPackage uploads the connector version package to the registry
func uploadConnectorVersionPackage(ciCtx Context, connector Connector, version string, changedConnectorVersionPath string) (ConnectorVersion, error) {

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

	uploadedTgzUrl, err := uploadConnectorVersionDefinition(ciCtx, connector.Namespace, connector.Name, version, connectorMetadataTgzPath)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to upload the connector version definition - connector: %v version:%v - err: %v", connector.Name, version, err)
	} else {
		// print success message with the name of the connector and the version
		fmt.Printf("Successfully uploaded the connector version definition in google cloud registry for the connector: %v version: %v\n", connector.Name, version)
	}

	// Build payload for registry upsert
	return buildRegistryPayload(ciCtx, connector.Namespace, connector.Name, version, connectorVersionMetadata, uploadedTgzUrl)
}

func uploadConnectorVersionDefinition(ciCtx Context, connectorNamespace, connectorName string, connectorVersion string, connectorMetadataTgzPath string) (string, error) {
	bucketName := ciCmdArgs.GCPBucketName
	objectName := generateGCPObjectName(connectorNamespace, connectorName, connectorVersion)
	uploadedTgzUrl, err := uploadFile(ciCtx.StorageClient, bucketName, objectName, connectorMetadataTgzPath)

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

// buildRegistryPayload builds the payload for the registry upsert API
func buildRegistryPayload(
	ciCtx Context,
	connectorNamespace string,
	connectorName string,
	version string,
	connectorVersionMetadata map[string]interface{},
	uploadedConnectorDefinitionTgzUrl string,
) (ConnectorVersion, error) {
	var connectorVersion ConnectorVersion
	var connectorVersionDockerImage string = ""
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

	connectorInfo, err := getConnectorInfoFromRegistry(*ciCtx.RegistryGQLClient, connectorNamespace, connectorName)

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

	var connectorVersionImage *string

	if connectorVersionDockerImage == "" {
		connectorVersionImage = nil
	} else {
		connectorVersionImage = &connectorVersionDockerImage
	}

	connectorVersion = ConnectorVersion{
		Namespace:            connectorNamespace,
		Name:                 connectorName,
		Version:              version,
		Image:                connectorVersionImage,
		PackageDefinitionURL: uploadedConnectorDefinitionTgzUrl,
		IsMultitenant:        connectorInfo.HubRegistryConnector[0].MultitenantConnector != nil,
		Type:                 connectorVersionType,
	}

	return connectorVersion, nil
}
