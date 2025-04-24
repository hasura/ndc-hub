package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/hasura/ndc-hub/registry-automation/pkg"

	"log"
	"os"
	"regexp"

	"cloud.google.com/go/storage"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
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
	var cloudinaryWrapper *CloudinaryWrapper
	var storageWrapper *StorageClientWrapper

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
		storageClient, err = storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(gcpServiceAccountDetails)))
		if err != nil {
			log.Fatalf("Failed to create Google bucket client: %v", err)
		}
		defer storageClient.Close()

		storageWrapper = &StorageClientWrapper{storageClient}

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
		cloudinaryWrapper = &CloudinaryWrapper{cloudinaryClient}

	}

	return Context{
		Env:               ciCmdArgs.PublicationEnv,
		RegistryGQLClient: registryGQLClient,
		StorageClient:     storageWrapper,
		Cloudinary:        cloudinaryWrapper,
	}

}

type fileProcessor struct {
	regex               *regexp.Regexp
	newFileHandler      func(matches []string, file string)
	modifiedFileHandler func(matches []string, file string) error
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
		ModifiedLogos:        make(map[Connector]Logo),
		ModifiedReadmes:      make(map[Connector]string),
		NewConnectors:        make(map[Connector]MetadataFile),
		NewLogos:             make(map[Connector]Logo),
		NewReadmes:           make(map[Connector]string),
		ModifiedConnectors:   make(map[Connector]MetadataFile),
	}

	processors := []fileProcessor{
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/metadata.json$`),
			newFileHandler: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.NewConnectors[connector] = MetadataFile(file)
				fmt.Fprintf(os.Stderr, "Processing metadata file for new connector: %s\n", connector.Name)
			},
			modifiedFileHandler: func(matches []string, file string) error {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.ModifiedConnectors[connector] = MetadataFile(file)
				fmt.Fprintf(os.Stderr, "Processing metadata file for modified connector: %s\n", connector.Name)
				return nil
			},
		},
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/logo\.(png|svg)$`),
			newFileHandler: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.NewLogos[connector] = Logo{Path: file, Extension: LogoExtension(matches[3])}
				fmt.Fprintf(os.Stderr, "Processing logo file for new connector: %s\n", connector.Name)
			},
			modifiedFileHandler: func(matches []string, file string) error {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.ModifiedLogos[connector] = Logo{Path: file, Extension: LogoExtension(matches[3])}
				fmt.Fprintf(os.Stderr, "Processing logo file for modified connector: %s\n", connector.Name)
				return nil
			},
		},
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/README\.md$`),
			newFileHandler: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.NewReadmes[connector] = file
				fmt.Fprintf(os.Stderr, "Processing README file for new connector: %s\n", connector.Name)
			},
			modifiedFileHandler: func(matches []string, file string) error {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				result.ModifiedReadmes[connector] = file
				fmt.Fprintf(os.Stderr, "Processing README file for modified connector: %s\n", connector.Name)
				return nil
			},
		},
		{
			regex: regexp.MustCompile(`^registry/([^/]+)/([^/]+)/releases/([^/]+)/connector-packaging\.json$`),
			newFileHandler: func(matches []string, file string) {
				connector := Connector{Name: matches[2], Namespace: matches[1]}
				version := matches[3]
				if _, exists := result.NewConnectorVersions[connector]; !exists {
					result.NewConnectorVersions[connector] = make(map[string]string)
				}
				result.NewConnectorVersions[connector][version] = file
			},
			modifiedFileHandler: func(matches []string, file string) error {
				return fmt.Errorf("Connector packaging files (%s) are immutable and should not be changed", file)
			},
		},
	}

	processFile := func(file string, isModified bool) {
		for _, processor := range processors {
			if matches := processor.regex.FindStringSubmatch(file); matches != nil {
				if isModified {
					processError := processor.modifiedFileHandler(matches, file)
					if processError != nil {
						fmt.Fprintf(os.Stderr, "Error processing modified %s file: %s: %v\n", matches[2], file, processError)
					}
				} else {
					processor.newFileHandler(matches, file)
				}
				return
			}
		}
		fmt.Fprintf(os.Stderr, "Skipping %s file: %s\n", map[bool]string{true: "modified", false: "newly added"}[isModified], file)
	}

	for _, file := range changedFiles.Added {
		processFile(file, false)
	}

	for _, file := range changedFiles.Modified {
		processFile(file, true)
	}

	return result
}

// processModifiedConnectors processes the modified connectors and updates the connector metadata in the registry
// This function updates the registry with the latest version, title, and description of the connector
func processModifiedConnector(metadataFile MetadataFile, connector Connector) (ConnectorOverviewUpdate, error) {
	// Iterate over the modified connectors and update the connectors in the registry
	var connectorOverviewUpdate ConnectorOverviewUpdate
	connectorMetadata, err := readJSONFile[ndchub.ConnectorMetadata](string(metadataFile))
	if err != nil {
		return connectorOverviewUpdate, fmt.Errorf("Failed to parse the connector metadata file: %v", err)
	}

	connectorOverviewUpdate = ConnectorOverviewUpdate{
		Set: struct {
			Docs          *string `json:"docs,omitempty"`
			Logo          *string `json:"logo,omitempty"`
			LatestVersion *string `json:"latest_version,omitempty"`
			Title         *string `json:"title,omitempty"`
			Description   *string `json:"description,omitempty"`
		}{
			LatestVersion: &connectorMetadata.Overview.LatestVersion,
			Title:         &connectorMetadata.Overview.Title,
			Description:   &connectorMetadata.Overview.Description,
		},
		Where: WhereClause{
			ConnectorName:      connector.Name,
			ConnectorNamespace: connector.Namespace,
		},
	}
	return connectorOverviewUpdate, nil
}

func processNewConnector(ciCtx Context, connector Connector, metadataFile MetadataFile, logoPath Logo) (ConnectorOverviewInsert, HubRegistryConnectorInsertInput, error) {
	// Process the newly added connector
	// Get the string value from metadataFile
	var connectorOverviewAndAuthor ConnectorOverviewInsert
	var hubRegistryConnectorInsertInput HubRegistryConnectorInsertInput

	connectorMetadata, err := readJSONFile[ndchub.ConnectorMetadata](string(metadataFile))
	if err != nil {
		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, fmt.Errorf("Failed to parse the connector metadata file: %v", err)
	}

	docs, err := readFile(fmt.Sprintf("registry/%s/%s/README.md", connector.Namespace, connector.Name))

	if err != nil {
		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, fmt.Errorf("Failed to read the README file of the connector: %s : %v", connector.Name, err)
	}

	uploadedLogoUrl, err := uploadLogoToCloudinary(ciCtx.Cloudinary, Connector{Name: connector.Name, Namespace: connector.Namespace}, logoPath)
	if err != nil {
		return connectorOverviewAndAuthor, hubRegistryConnectorInsertInput, err
	}

	// Get connector info from the registry
	connectorInfo, err := getConnectorInfoFromRegistry(ciCtx.RegistryGQLClient, connector.Name, connector.Namespace)
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
		Name:          connector.Name,
		Namespace:     connector.Namespace,
		Docs:          string(docs),
		Logo:          uploadedLogoUrl,
		Title:         connectorMetadata.Overview.Title,
		Description:   connectorMetadata.Overview.Description,
		IsVerified:    connectorMetadata.IsVerified,
		IsHosted:      connectorMetadata.IsHostedByHasura,
		LatestVersion: connectorMetadata.Overview.LatestVersion,
		Author: ConnectorAuthorNestedInsert{
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
	modifiedConnectors := processChangedFiles.ModifiedConnectors
	newLogos := processChangedFiles.NewLogos

	var newConnectorsToBeAdded NewConnectorsInsertInput
	newConnectorsToBeAdded.HubRegistryConnectors = make([]HubRegistryConnectorInsertInput, 0)
	newConnectorsToBeAdded.ConnectorOverviews = make([]ConnectorOverviewInsert, 0)
	newConnectorOverviewsToBeAdded := make([](ConnectorOverviewInsert), 0)
	hubRegistryConnectorsToBeAdded := make([](HubRegistryConnectorInsertInput), 0)
	connectorOverviewUpdates := make([]ConnectorOverviewUpdate, 0)
	newConnectorVersionsToBeAdded := make([]ConnectorVersion, 0)

	if len(newlyAddedConnectors) > 0 {
		fmt.Println("New connectors to be added to the registry: ", newlyAddedConnectors)

		for connector, metadataFile := range newlyAddedConnectors {
			// Find the logo corresponding to the connector from the newLogos map, throw error if not found
			logoPath := newLogos[connector]
			connectorOverviewAndAuthor, hubRegistryConnector, err := processNewConnector(ctx, connector, metadataFile, logoPath)

			if err != nil {
				log.Fatalf("Failed to process the new connector: %s/%s, Error: %v", connector.Namespace, connector.Name, err)
			}
			newConnectorOverviewsToBeAdded = append(newConnectorOverviewsToBeAdded, connectorOverviewAndAuthor)
			hubRegistryConnectorsToBeAdded = append(hubRegistryConnectorsToBeAdded, hubRegistryConnector)

		}

		newConnectorsToBeAdded.HubRegistryConnectors = hubRegistryConnectorsToBeAdded
		newConnectorsToBeAdded.ConnectorOverviews = newConnectorOverviewsToBeAdded

	}

	if len(modifiedConnectors) > 0 {
		fmt.Println("Modified connectors: ", modifiedConnectors)
		// Process the modified connectors
		for connector, metadataFile := range modifiedConnectors {
			connectorOverviewUpdate, err := processModifiedConnector(metadataFile, connector)
			if err != nil {
				log.Fatalf("Failed to process the modified connector: %s/%s, Error: %v", connector.Namespace, connector.Name, err)
			}
			connectorOverviewUpdates = append(connectorOverviewUpdates, connectorOverviewUpdate)
		}
	}

	if len(newlyAddedConnectorVersions) > 0 {
		newlyAddedConnectors := make(map[Connector]bool)
		for connector := range newlyAddedConnectorVersions {
			newlyAddedConnectors[connector] = true
		}
		newConnectorVersionsToBeAdded = processNewlyAddedConnectorVersions(ctx, newlyAddedConnectorVersions, newlyAddedConnectors)
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
		logoUpdates, err := processModifiedLogos(modifiedLogos, ctx.Cloudinary)
		if err != nil {
			log.Fatalf("Failed to process the modified logos: %v", err)
		}
		connectorOverviewUpdates = append(connectorOverviewUpdates, logoUpdates...)
		fmt.Println("Successfully updated the logos in the registry.")
	}

	if len(newConnectorVersionsToBeAdded) > 0 {
		var err error
		if ctx.Env == "production" {
			err = registryDbMutation(ctx.RegistryGQLClient, newConnectorsToBeAdded, connectorOverviewUpdates, newConnectorVersionsToBeAdded)

		} else if ctx.Env == "staging" {
			err = registryDbMutationStaging(ctx.RegistryGQLClient, newConnectorsToBeAdded, connectorOverviewUpdates, newConnectorVersionsToBeAdded)
		} else {
			log.Fatalf("Unexpected: invalid publication environment: %s", ctx.Env)
		}

		if err != nil {
			log.Fatalf("Failed to update the registry: %v", err)
		}

	}
	fmt.Println("Successfully processed the changed files in the PR")
}

func uploadLogoToCloudinary(cloudinary CloudinaryInterface, connector Connector, logo Logo) (string, error) {
	logoContent, err := readFile(logo.Path)
	if err != nil {
		fmt.Printf("Failed to read the logo file: %v", err)
		return "", err
	}

	imageReader := bytes.NewReader(logoContent)

	uploadResult, err := cloudinary.Upload(context.Background(), imageReader, uploader.UploadParams{
		PublicID: fmt.Sprintf("%s-%s", connector.Namespace, connector.Name),
		Format:   string(logo.Extension),
	})
	if err != nil {
		return "", fmt.Errorf("Failed to upload the logo to cloudinary for the connector: %s, Error: %v\n", connector.Name, err)
	}
	return uploadResult.SecureURL, nil
}

func processModifiedLogos(modifiedLogos ModifiedLogos, cloudinaryClient CloudinaryInterface) ([]ConnectorOverviewUpdate, error) {
	// Iterate over the modified logos and update the logos in the registry
	var connectorOverviewUpdates []ConnectorOverviewUpdate

	for connector, logoPath := range modifiedLogos {
		// open the logo file
		uploadedLogoUrl, err := uploadLogoToCloudinary(cloudinaryClient, connector, logoPath)
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

func processNewlyAddedConnectorVersions(ciCtx Context, newlyAddedConnectorVersions NewConnectorVersions, newConnectorsAdded map[Connector]bool) []ConnectorVersion {
	// Iterate over the added or modified connectors and upload the connector versions
	var connectorVersions []ConnectorVersion
	var uploadConnectorVersionErr error
	encounteredError := false

	for connectorName, versions := range newlyAddedConnectorVersions {

		for version, connectorVersionPath := range versions {
			var connectorVersion ConnectorVersion
			isNewConnector := newConnectorsAdded[connectorName]
			connectorVersion, uploadConnectorVersionErr = uploadConnectorVersionPackage(ciCtx, connectorName, version, connectorVersionPath, isNewConnector)

			if uploadConnectorVersionErr != nil {
				if errors.Is(uploadConnectorVersionErr, errV2Connector) {
					fmt.Printf("Skipping v2 connector upload: %s - %s", version, connectorName)
					continue
				}
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

	return connectorVersions

}

func cleanupUploadedConnectorVersions(client StorageClientInterface, connectorVersions []ConnectorVersion) error {
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

var errV2Connector = errors.New("v2 connectors are not required to be published")

// uploadConnectorVersionPackage uploads the connector version package to the registry
func uploadConnectorVersionPackage(ciCtx Context, connector Connector, version string, changedConnectorVersionPath string, isNewConnector bool) (ConnectorVersion, error) {

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

	connectorVersionMetadata, connectorMetadataTgzPath, err := pkg.GetConnectorVersionMetadata(tgzUrl,
		connector.Namespace, connector.Name, version)
	if err != nil {
		return connectorVersion, err
	}

	if connectorVersionMetadata["version"] != nil {
		packagingSpecVersion := connectorVersionMetadata["version"].(string)
		if packagingSpecVersion == "v2" {
			return connectorVersion, errV2Connector
		}
	}

	uploadedTgzUrl, err := uploadConnectorVersionDefinition(ciCtx, connector.Namespace, connector.Name, version, connectorMetadataTgzPath)
	if err != nil {
		return connectorVersion, fmt.Errorf("failed to upload the connector version definition - connector: %v version:%v - err: %v", connector.Name, version, err)
	} else {
		// print success message with the name of the connector and the version
		fmt.Printf("Successfully uploaded the connector version definition in google cloud registry for the connector: %v version: %v\n", connector.Name, version)
	}

	// Build payload for registry upsert
	return buildRegistryPayload(ciCtx, connector.Namespace, connector.Name, version, connectorVersionMetadata, uploadedTgzUrl, isNewConnector)
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

// buildRegistryPayload builds the payload for the registry upsert API
func buildRegistryPayload(
	ciCtx Context,
	connectorNamespace string,
	connectorName string,
	version string,
	connectorVersionMetadata map[string]interface{},
	uploadedConnectorDefinitionTgzUrl string,
	isNewConnector bool,
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

	connectorInfo, err := getConnectorInfoFromRegistry(ciCtx.RegistryGQLClient, connectorNamespace, connectorName)

	if err != nil {
		return connectorVersion, err
	}

	var isMultitenant bool

	// Check if the connector exists in the registry first
	if len(connectorInfo.HubRegistryConnector) == 0 {

		if isNewConnector {
			isMultitenant = false
		} else {
			return connectorVersion, fmt.Errorf("Unexpected: Couldn't get the connector info of the connector: %s", connectorName)

		}

	} else {
		if len(connectorInfo.HubRegistryConnector) == 1 {
			// check if the connector is multitenant
			isMultitenant = connectorInfo.HubRegistryConnector[0].MultitenantConnector != nil

		}

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
		IsMultitenant:        isMultitenant,
		Type:                 connectorVersionType,
	}

	return connectorVersion, nil
}
