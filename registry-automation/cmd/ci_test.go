package cmd

import (
	"context"

	"testing"

	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cloud.google.com/go/storage"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Mock structures
type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) Bucket(name string) *storage.BucketHandle {
	args := m.Called(name)
	return args.Get(0).(*storage.BucketHandle)
}

type MockCloudinaryUploader struct {
	mock.Mock
}

type MockCloudinary struct {
	mock.Mock
}

func (m *MockCloudinary) Upload(ctx context.Context, file interface{}, uploadParams uploader.UploadParams) (*uploader.UploadResult, error) {
	args := m.Called(ctx, file, uploadParams)
	return args.Get(0).(*uploader.UploadResult), args.Error(1)
}

type MockGraphQLClient struct {
	mock.Mock
}

func (m *MockGraphQLClient) Run(ctx context.Context, query *graphql.Request, resp interface{}) error {
	args := m.Called(ctx, query, resp)
	return args.Error(0)
}

func createTestContext() Context {
	return Context{
		Env:               "staging",
		RegistryGQLClient: &MockGraphQLClient{},
		StorageClient:     &MockStorageClient{},
		Cloudinary:        &MockCloudinary{},
	}
}

// Test processChangedFiles
func TestProcessChangedFiles(t *testing.T) {
	testCases := []struct {
		name         string
		changedFiles ChangedFiles
		expected     ProcessedChangedFiles
	}{
		{
			name: "New connector added",
			changedFiles: ChangedFiles{
				Added: []string{"registry/namespace1/connector1/metadata.json"},
			},
			expected: ProcessedChangedFiles{
				NewConnectorVersions: map[Connector]map[string]string{},
				ModifiedLogos:        map[Connector]string{},
				ModifiedReadmes:      map[Connector]string{},
				NewConnectors:        map[Connector]MetadataFile{{Name: "connector1", Namespace: "namespace1"}: "registry/namespace1/connector1/metadata.json"},
				NewLogos:             map[Connector]string{},
				NewReadmes:           map[Connector]string{},
			},
		},
		{
			name: "Modified logo and README",
			changedFiles: ChangedFiles{
				Modified: []string{
					"registry/namespace1/connector1/logo.png",
					"registry/namespace1/connector1/README.md",
				},
			},
			expected: ProcessedChangedFiles{
				NewConnectorVersions: map[Connector]map[string]string{},
				ModifiedLogos:        map[Connector]string{{Name: "connector1", Namespace: "namespace1"}: "registry/namespace1/connector1/logo.png"},
				ModifiedReadmes:      map[Connector]string{{Name: "connector1", Namespace: "namespace1"}: "registry/namespace1/connector1/README.md"},
				NewConnectors:        map[Connector]MetadataFile{},
				NewLogos:             map[Connector]string{},
				NewReadmes:           map[Connector]string{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processChangedFiles(tc.changedFiles)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// func TestProcessNewConnector(t *testing.T) {
// 	ctx := createTestContext()
// 	connector := Connector{Name: "testconnector", Namespace: "testnamespace"}

// 	// Create a temporary directory for our test files
// 	tempDir, err := os.MkdirTemp("", "connector-test")
// 	assert.NoError(t, err)
// 	defer os.RemoveAll(tempDir) // Clean up after the test

// 	// Set up the directory structure
// 	registryDir := filepath.Join(tempDir, "registry", connector.Namespace, connector.Name)
// 	err = os.MkdirAll(registryDir, 0755)
// 	assert.NoError(t, err)

// 	// Create the metadata file
// 	metadataFile := filepath.Join(registryDir, "metadata.json")
// 	tempMetadata := []byte(`{"overview": {"title": "Test Connector", "description": "A test connector"}, "isVerified": true, "isHostedByHasura": false, "author": {"name": "Test Author", "supportEmail": "support@test.com", "homepage": "https://test.com"}}`)
// 	err = os.WriteFile(metadataFile, tempMetadata, 0666)
// 	assert.NoError(t, err)

// 	// Create the README file
// 	readmeFile := filepath.Join(registryDir, "README.md")
// 	err = os.WriteFile(readmeFile, []byte("# Test Connector"), 0644)
// 	assert.NoError(t, err)

// 	// Mock the necessary functions and API calls
// 	mockCloudinaryUploader := &MockCloudinaryUploader{}
// 	mockCloudinaryUploader.On("Upload", mock.Anything, mock.Anything, mock.Anything).Return(&uploader.UploadResult{SecureURL: "https://res.cloudinary.com/demo/image/upload/logo.png"}, nil)

// 	mockGraphQLClient := ctx.RegistryGQLClient.(*MockGraphQLClient)
// 	mockGraphQLClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	// Run the function
// 	connectorOverviewInsert, hubRegistryConnectorInsert, err := processNewConnector(ctx, connector, MetadataFile(metadataFile))

// 	// Assert the results
// 	assert.NoError(t, err)
// 	assert.Equal(t, "testconnector", connectorOverviewInsert.Name)
// 	assert.Equal(t, "testnamespace", connectorOverviewInsert.Namespace)
// 	assert.Equal(t, "Test Connector", connectorOverviewInsert.Title)
// 	assert.Equal(t, "A test connector", connectorOverviewInsert.Description)
// 	assert.True(t, connectorOverviewInsert.IsVerified)
// 	assert.False(t, connectorOverviewInsert.IsHosted)
// 	assert.Equal(t, "Test Author", connectorOverviewInsert.Author.Data.Name)
// 	assert.Equal(t, "support@test.com", connectorOverviewInsert.Author.Data.SupportEmail)
// 	assert.Equal(t, "https://test.com", connectorOverviewInsert.Author.Data.Website)

// 	assert.Equal(t, "testconnector", hubRegistryConnectorInsert.Name)
// 	assert.Equal(t, "testnamespace", hubRegistryConnectorInsert.Namespace)
// 	assert.Equal(t, "Test Connector", hubRegistryConnectorInsert.Title)

// 	mockCloudinaryUploader.AssertExpectations(t)
// 	mockGraphQLClient.AssertExpectations(t)
// }

// // Test uploadConnectorVersionPackage
// func TestUploadConnectorVersionPackage(t *testing.T) {
// 	ctx := createTestContext()
// 	connector := Connector{Name: "testconnector", Namespace: "testnamespace"}
// 	version := "v1.0.0"
// 	changedConnectorVersionPath := "registry/testnamespace/testconnector/releases/v1.0.0/connector-packaging.json"
// 	isNewConnector := true

// 	// Mock necessary functions
// 	mockStorageClient := ctx.StorageClient.(*MockStorageClient)
// 	mockStorageClient.On("Bucket", mock.Anything).Return(&storage.BucketHandle{})

// 	mockGraphQLClient := ctx.RegistryGQLClient.(*MockGraphQLClient)
// 	mockGraphQLClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	// Create temporary files
// 	err := os.MkdirAll("registry/testnamespace/testconnector/releases/v1.0.0", 0755)
// 	assert.NoError(t, err)
// 	defer os.RemoveAll("registry/testnamespace/testconnector")

// 	packagingContent := []byte(`{"uri": "https://example.com/testconnector-v1.0.0.tgz"}`)
// 	err = os.WriteFile(changedConnectorVersionPath, packagingContent, 0644)
// 	assert.NoError(t, err)

// 	// Run the function
// 	connectorVersion, err := uploadConnectorVersionPackage(ctx, connector, version, changedConnectorVersionPath, isNewConnector)

// 	// Assert the results
// 	assert.NoError(t, err)
// 	assert.Equal(t, "testconnector", connectorVersion.Name)
// 	assert.Equal(t, "testnamespace", connectorVersion.Namespace)
// 	assert.Equal(t, "v1.0.0", connectorVersion.Version)

// 	mockStorageClient.AssertExpectations(t)
// 	mockGraphQLClient.AssertExpectations(t)
// }

// // Test buildRegistryPayload
// func TestBuildRegistryPayload(t *testing.T) {
// 	ctx := createTestContext()
// 	connectorNamespace := "testnamespace"
// 	connectorName := "testconnector"
// 	version := "v1.0.0"
// 	connectorVersionMetadata := map[string]interface{}{
// 		"packagingDefinition": map[string]interface{}{
// 			"type": "ManagedDockerBuild",
// 		},
// 	}
// 	uploadedConnectorDefinitionTgzUrl := "https://example.com/test.tgz"
// 	isNewConnector := true

// 	// Mock the GraphQL client
// 	mockGraphQLClient := ctx.RegistryGQLClient.(*MockGraphQLClient)
// 	mockGraphQLClient.On("Run", mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	// Run the function
// 	connectorVersion, err := buildRegistryPayload(ctx, connectorNamespace, connectorName, version, connectorVersionMetadata, uploadedConnectorDefinitionTgzUrl, isNewConnector)

// 	// Assert the results
// 	assert.NoError(t, err)
// 	assert.Equal(t, connectorNamespace, connectorVersion.Namespace)
// 	assert.Equal(t, connectorName, connectorVersion.Name)
// 	assert.Equal(t, version, connectorVersion.Version)
// 	assert.Equal(t, uploadedConnectorDefinitionTgzUrl, connectorVersion.PackageDefinitionURL)
// 	assert.Equal(t, "ManagedDockerBuild", connectorVersion.Type)
// 	assert.False(t, connectorVersion.IsMultitenant)
// 	assert.Nil(t, connectorVersion.Image)

// 	mockGraphQLClient.AssertExpectations(t)
// }
