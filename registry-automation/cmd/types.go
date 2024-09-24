package cmd

import (
	"cloud.google.com/go/storage"
	"encoding/json"
	"github.com/cloudinary/cloudinary-go/v2"
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

type Context struct {
	Env               string
	RegistryGQLClient *graphql.Client
	StorageClient     *storage.Client
	Cloudinary        *cloudinary.Cloudinary
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
