package cmd

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/storage"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/machinebox/graphql"
)

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
		Docs          *string `json:"docs,omitempty"`
		Logo          *string `json:"logo,omitempty"`
		LatestVersion *string `json:"latest_version,omitempty"`
		Title         *string `json:"title,omitempty"`
		Description   *string `json:"description,omitempty"`
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

type MetadataFile string

type NewConnectors map[Connector]MetadataFile

type ProcessedChangedFiles struct {
	NewConnectorVersions NewConnectorVersions
	ModifiedLogos        ModifiedLogos
	ModifiedReadmes      ModifiedReadmes
	NewConnectors        NewConnectors
	NewLogos             NewLogos
	NewReadmes           NewReadmes
	ModifiedConnectors   ModifiedMetadata
}

type GraphQLClientInterface interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type StorageClientWrapper struct {
	*storage.Client
}

func (s *StorageClientWrapper) Bucket(name string) *storage.BucketHandle {
	return s.Client.Bucket(name)
}

type StorageClientInterface interface {
	Bucket(name string) *storage.BucketHandle
}

type CloudinaryInterface interface {
	Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error)
}

type CloudinaryWrapper struct {
	*cloudinary.Cloudinary
}

func (c *CloudinaryWrapper) Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error) {
	return c.Cloudinary.Upload.Upload(ctx, file, uploadParams)
}

//

type Context struct {
	Env               string
	RegistryGQLClient GraphQLClientInterface
	StorageClient     StorageClientInterface
	Cloudinary        CloudinaryInterface
}

// Type that uniquely identifies a connector
type Connector struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Logo struct {
	Path      string
	Extension LogoExtension
}

type LogoExtension string

const (
	PNG LogoExtension = "png"
	SVG LogoExtension = "svg"
)

type NewConnectorVersions map[Connector]map[string]string

// ModifiedMetadata represents the modified metadata in the PR, the key is the connector name and the value is the path to the modified metadata
type ModifiedMetadata map[Connector]MetadataFile

// ModifiedLogos represents the modified logos in the PR, the key is the connector name and the value is the path to the modified logo
type ModifiedLogos map[Connector]Logo

// ModifiedReadmes represents the modified READMEs in the PR, the key is the connector name and the value is the path to the modified README
type ModifiedReadmes map[Connector]string

// ModifiedLogos represents the modified logos in the PR, the key is the connector name and the value is the path to the modified logo
type NewLogos map[Connector]Logo

// ModifiedReadmes represents the modified READMEs in the PR, the key is the connector name and the value is the path to the modified README
type NewReadmes map[Connector]string
