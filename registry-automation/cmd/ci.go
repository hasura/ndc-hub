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

type WhereClause struct {
	ConnectorName      string
	ConnectorNamespace string
}

func (wc WhereClause) MarshalJSON() ([]byte, error) {
	where := map[string]interface{}{
		"_and": []map[string]interface{}{
			{"name": map[string]string{"_eq": wc.ConnectorName}},
			{"namespace": map[string]string{"_eq": wc.ConnectorNamespace}},
		},
	}
	return json.Marshal(where)
}

type ConnectorOverviewUpdate struct {
	Set struct {
		Docs *string `json:"docs,omitempty"`
		Logo *string `json:"logo,omitempty"`
	} `json:"_set"`
	Where WhereClause `json:"where"`
}

type ConnectorOverviewUpdates struct {
	Updates []ConnectorOverviewUpdate `json:"updates"`
}

const (
	ManagedDockerBuild  = "ManagedDockerBuild"
	PrebuiltDockerImage = "PrebuiltDockerImage"
)

// Type to represent the metadata.json file
type ConnectorMetadata struct {
	Overview struct {
		Namespace     string   `json:"namespace"`
		Description   string   `json:"description"`
		Title         string   `json:"title"`
		Logo          string   `json:"logo"`
		Tags          []string `json:"tags"`
		LatestVersion string   `json:"latest_version"`
	} `json:"overview"`
	Author struct {
		SupportEmail string `json:"support_email"`
		Homepage     string `json:"homepage"`
		Name         string `json:"name"`
	} `json:"author"`

	IsVerified         bool `json:"is_verified"`
	IsHostedByHasura   bool `json:"is_hosted_by_hasura"`
	HasuraHubConnector struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	} `json:"hasura_hub_connector"`
	SourceCode struct {
		IsOpenSource bool   `json:"is_open_source"`
		Repository   string `json:"repository"`
	} `json:"source_code"`
}

// Make a struct with the fields expected in the command line arguments
type ConnectorRegistryArgs struct {
	ChangedFilesPath         string
	PublicationEnv           string
	ConnectorRegistryGQLUrl  string
	ConnectorPublicationKey  string
	GCPServiceAccountDetails string
	GCPBucketName            string
	CloudinaryUrl            string
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

type NewConnector struct {
	// Name of the connector, e.g. "mongodb"
	Name string
	// Namespace of the connector, e.g. "hasura"
	Namespace string
}

type MetadataFile string

type NewConnectors map[NewConnector]MetadataFile

type ProcessedChangedFiles struct {
	NewConnectorVersions NewConnectorVersions
	ModifiedLogos        ModifiedLogos
	ModifiedReadmes      ModifiedReadmes
	NewConnectors        NewConnectors
	NewLogos             NewLogos
	NewReadmes           NewReadmes
}

// processChangedFiles processes the files in the PR and extracts the connector name and version
// This function checks for the following things:
// 1. If a new connector version is added, it adds the connector version to the `newlyAddedConnectorVersions` map.
// 2. If the logo file is modified, it adds the connector name and the path to the modified logo to the `modifiedLogos` map.
// 3. If the README file is modified, it adds the connector name and the path to the modified README to the `modifiedReadmes` map.
func processChangedFiles(changedFiles ChangedFiles) ProcessedChangedFiles {

	newlyAddedConnectorVersions := make(map[Connector]map[string]string)
	modifiedLogos := make(map[Connector]string)
	modifiedReadmes := make(map[Connector]string)
	newConnectors := make(map[NewConnector]MetadataFile)
	newLogos := make(map[Connector]string)
	newReadmes := make(map[Connector]string)

	var connectorVersionPackageRegex = regexp.MustCompile(`^registry/([^/]+)/([^/]+)/releases/([^/]+)/connector-packaging\.json$`)
	var logoPngRegex = regexp.MustCompile(`^registry/([^/]+)/([^/]+)/logo\.(png|svg)$`)
	var readmeMdRegex = regexp.MustCompile(`^registry/([^/]+)/([^/]+)/README\.md$`)
	var connectorMetadataRegex = regexp.MustCompile(`^registry/([^/]+)/([^/]+)/metadata.json$`)

	for _, file := range changedFiles.Added {

		if connectorMetadataRegex.MatchString(file) {
			// Process the metadata file
			// print the name of the connector and the version
			matches := connectorMetadataRegex.FindStringSubmatch(file)
			if len(matches) == 3 {
				connectorNamespace := matches[1]
				connectorName := matches[2]
				connector := NewConnector{
					Name:      connectorName,
					Namespace: connectorNamespace,
				}
				newConnectors[connector] = MetadataFile(file)
				fmt.Printf("Processing metadata file for connector: %s\n", connectorName)
			}
		} else if logoPngRegex.MatchString(file) {
			// Process the logo file
			// print the name of the connector and the version
			matches := logoPngRegex.FindStringSubmatch(file)
			if len(matches) == 4 {

				connectorNamespace := matches[1]
				connectorName := matches[2]
				connector := Connector{
					Name:      connectorName,
					Namespace: connectorNamespace,
				}
				newLogos[connector] = file
				fmt.Printf("Processing logo file for connector: %s\n", connectorName)
			}

		} else if readmeMdRegex.MatchString(file) {
			// Process the README file
			// print the name of the connector and the version
			matches := readmeMdRegex.FindStringSubmatch(file)

			if len(matches) == 3 {

				connectorNamespace := matches[1]
				connectorName := matches[2]
				connector := Connector{
					Name:      connectorName,
					Namespace: connectorNamespace,
				}

				newReadmes[connector] = file

				fmt.Printf("Processing README file for connector: %s\n", connectorName)
			}
		} else if connectorVersionPackageRegex.MatchString(file) {

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
			fmt.Println("Skipping newly added file: ", file)
		}

	}

	for _, file := range changedFiles.Modified {
		if logoPngRegex.MatchString(file) {
			// Process the logo file
			// print the name of the connector and the version
			matches := logoPngRegex.FindStringSubmatch(file)
			if len(matches) == 4 {

				connectorNamespace := matches[1]
				connectorName := matches[2]
				connector := Connector{
					Name:      connectorName,
					Namespace: connectorNamespace,
				}
				modifiedLogos[connector] = file
				fmt.Printf("Processing logo file for connector: %s\n", connectorName)
			}

		} else if readmeMdRegex.MatchString(file) {
			// Process the README file
			// print the name of the connector and the version
			matches := readmeMdRegex.FindStringSubmatch(file)

			if len(matches) == 3 {

				connectorNamespace := matches[1]
				connectorName := matches[2]
				connector := Connector{
					Name:      connectorName,
					Namespace: connectorNamespace,
				}

				modifiedReadmes[connector] = file

				fmt.Printf("Processing README file for connector: %s\n", connectorName)
			}
		} else {
			fmt.Println("Skipping modified file: ", file)
		}

	}

	return ProcessedChangedFiles{
		NewConnectorVersions: newlyAddedConnectorVersions,
		ModifiedLogos:        modifiedLogos,
		ModifiedReadmes:      modifiedReadmes,
		NewConnectors:        newConnectors,
		NewLogos:             newLogos,
		NewReadmes:           newReadmes,
	}

}

func processNewConnector(ciCtx Context, connector NewConnector, metadataFile MetadataFile) {
	// Process the newly added connector
	// Get the string value from metadataFile

	connectorMetadata, err := readJSONFile[ConnectorMetadata](string(metadataFile))
	if err != nil {
		log.Fatalf("Failed to parse the connector metadata file: %v", err)
	}

	// Check if the connector already exists in the registry
	connectorInfo, err := getConnectorInfoFromRegistry(*ciCtx.RegistryGQLClient, connector.Name, connector.Namespace)
	if err != nil {
		log.Fatalf("Failed to get the connector info from the registry: %v", err)
	}

	if len(connectorInfo.HubRegistryConnector) > 0 {
		log.Fatalf("Attempting to create a new hub connector, but the connector already exists in the registry: %s/%s", connector.Namespace, connector.Name)
	}

	// Insert the connector in the registry
	err = insertConnectorInRegistry(*ciCtx.RegistryGQLClient, connectorMetadata, connector)
	if err != nil {
		log.Fatalf("Failed to insert the connector in the registry: %v", err)
	}

}

type HubRegistryConnectorInsertInput struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Namespace string `json:"namespace"`
}

func insertConnectorInRegistry(client graphql.Client, connectorMetadata ConnectorMetadata, connector NewConnector) error {
	var respData map[string]interface{}

	ctx := context.Background()

	// This if condition checks if the hub connector in the metadata file is the same as the connector in the PR, if yes, we proceed to
	// insert it as a hub connector in the registry. If the value is not the same, it means that the current connector is just an alias
	// to an existing connector in the registry and we skip inserting it as a new hub connector in the registry. Example: `postgres-cosmos` is
	// just an alias to the `postgres` connector in the registry.
	if (connectorMetadata.HasuraHubConnector.Namespace == connector.Namespace) && (connectorMetadata.HasuraHubConnector.Name == connector.Name) {

		req := graphql.NewRequest(`
mutation InsertHubRegistryConnector ($connector:hub_registry_connector_insert_input!){
  insert_hub_registry_connector_one(object: $connector) {
    name
    title
  }
}`)
		hubRegistryConnectorInsertInput := HubRegistryConnectorInsertInput{
			Name:      connector.Name,
			Title:     connectorMetadata.Overview.Title,
			Namespace: connector.Namespace,
		}

		req.Var("connector", hubRegistryConnectorInsertInput)
		req.Header.Set("x-hasura-role", "connector_publishing_automation")
		req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

		// Execute the GraphQL query and check the response.
		if err := client.Run(ctx, req, &respData); err != nil {
			return err
		} else {
			fmt.Printf("Successfully inserted the connector in the registry: %+v\n", respData)
		}

	}

	return nil

}

type Context struct {
	Env               string
	RegistryGQLClient *graphql.Client
	StorageClient     *storage.Client
	Cloudinary        *cloudinary.Cloudinary
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

	if len(newlyAddedConnectors) > 0 {
		fmt.Println("New connectors to be added to the registry: ", newlyAddedConnectors)

		for connector, metadataFile := range newlyAddedConnectors {
			processNewConnector(ctx, connector, metadataFile)
		}

	}

	// check if the map is empty
	if len(newlyAddedConnectorVersions) == 0 && len(modifiedLogos) == 0 && len(modifiedReadmes) == 0 {
		fmt.Println("No connectors to be added or modified in the registry")
		return
	} else {
		if len(newlyAddedConnectorVersions) > 0 {
			processNewlyAddedConnectorVersions(ctx, newlyAddedConnectorVersions)
		}

		if len(modifiedReadmes) > 0 {
			err := processModifiedReadmes(modifiedReadmes)
			if err != nil {
				log.Fatalf("Failed to process the modified READMEs: %v", err)
			}
			fmt.Println("Successfully updated the READMEs in the registry.")
		}

		if len(modifiedLogos) > 0 {
			err := processModifiedLogos(modifiedLogos)
			if err != nil {
				log.Fatalf("Failed to process the modified logos: %v", err)
			}
			fmt.Println("Successfully updated the logos in the registry.")
		}
	}

	fmt.Println("Successfully processed the changed files in the PR")
}

func processModifiedLogos(modifiedLogos ModifiedLogos) error {
	// Iterate over the modified logos and update the logos in the registry
	var connectorOverviewUpdates []ConnectorOverviewUpdate
	// upload the logo to cloudinary
	cloudinary, err := cloudinary.NewFromURL(ciCmdArgs.CloudinaryUrl)
	if err != nil {
		return err
	}

	for connector, logoPath := range modifiedLogos {
		// open the logo file
		logoContent, err := readFile(logoPath)
		if err != nil {
			fmt.Printf("Failed to read the logo file: %v", err)
			return err
		}

		imageReader := bytes.NewReader(logoContent)

		uploadResult, err := cloudinary.Upload.Upload(context.Background(), imageReader, uploader.UploadParams{
			PublicID: fmt.Sprintf("%s-%s", connector.Namespace, connector.Name),
			Format:   "png",
		})
		if err != nil {
			fmt.Printf("Failed to upload the logo to cloudinary for the connector: %s, Error: %v\n", connector.Name, err)
			return err
		} else {
			fmt.Printf("Successfully uploaded the logo to cloudinary for the connector: %s\n", connector.Name)
		}

		var connectorOverviewUpdate ConnectorOverviewUpdate

		if connectorOverviewUpdate.Set.Logo == nil {
			connectorOverviewUpdate.Set.Logo = new(string)
		} else {
			*connectorOverviewUpdate.Set.Logo = ""
		}

		*connectorOverviewUpdate.Set.Logo = string(uploadResult.SecureURL)

		connectorOverviewUpdate.Where.ConnectorName = connector.Name
		connectorOverviewUpdate.Where.ConnectorNamespace = connector.Namespace

		connectorOverviewUpdates = append(connectorOverviewUpdates, connectorOverviewUpdate)

	}

	return updateConnectorOverview(ConnectorOverviewUpdates{Updates: connectorOverviewUpdates})

}

func processModifiedReadmes(modifiedReadmes ModifiedReadmes) error {
	// Iterate over the modified READMEs and update the READMEs in the registry
	var connectorOverviewUpdates []ConnectorOverviewUpdate

	for connector, readmePath := range modifiedReadmes {
		// open the README file
		readmeContent, err := readFile(readmePath)
		if err != nil {
			return err

		}

		readMeContentString := string(readmeContent)

		var connectorOverviewUpdate ConnectorOverviewUpdate
		connectorOverviewUpdate.Set.Docs = &readMeContentString

		connectorOverviewUpdate.Where.ConnectorName = connector.Name
		connectorOverviewUpdate.Where.ConnectorNamespace = connector.Namespace

		connectorOverviewUpdates = append(connectorOverviewUpdates, connectorOverviewUpdate)

	}

	return updateConnectorOverview(ConnectorOverviewUpdates{Updates: connectorOverviewUpdates})

}

func processNewlyAddedConnectorVersions(ciCtx Context, newlyAddedConnectorVersions NewConnectorVersions) {
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

	} else {
		fmt.Printf("Connector versions to be added to the registry: %+v\n", connectorVersions)
		err := updateRegistryGQL(*ciCtx.RegistryGQLClient, connectorVersions)
		if err != nil {
			// attempt to cleanup the uploaded connector versions
			_ = cleanupUploadedConnectorVersions(ciCtx.StorageClient, connectorVersions) // ignore errors while cleaning up
			log.Fatalf("Failed to update the registry: %v", err)
		}
	}
	fmt.Println("Successfully added connector versions to the registry.")

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

// ModifiedLogos represents the modified logos in the PR, the key is the connector name and the value is the path to the modified logo
type ModifiedLogos map[Connector]string

// ModifiedReadmes represents the modified READMEs in the PR, the key is the connector name and the value is the path to the modified README
type ModifiedReadmes map[Connector]string

// ModifiedLogos represents the modified logos in the PR, the key is the connector name and the value is the path to the modified logo
type NewLogos map[Connector]string

// ModifiedReadmes represents the modified READMEs in the PR, the key is the connector name and the value is the path to the modified README
type NewReadmes map[Connector]string

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

func getConnectorInfoFromRegistry(client graphql.Client, connectorNamespace string, connectorName string) (GetConnectorInfoResponse, error) {
	var respData GetConnectorInfoResponse

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

func updateRegistryGQL(client graphql.Client, payload []ConnectorVersion) error {
	var respData map[string]interface{}

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

func updateConnectorOverview(updates ConnectorOverviewUpdates) error {
	var respData map[string]interface{}
	client := graphql.NewClient(ciCmdArgs.ConnectorRegistryGQLUrl)
	ctx := context.Background()

	req := graphql.NewRequest(`
mutation UpdateConnector ($updates: [connector_overview_updates!]!) {
  update_connector_overview_many(updates: $updates) {
    affected_rows
  }
}`)

	// add the payload to the request
	req.Var("updates", updates.Updates)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	} else {
		fmt.Printf("Successfully updated the connector overview: %+v\n", respData)
	}

	return nil
}
