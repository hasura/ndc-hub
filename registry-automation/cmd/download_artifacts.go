package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"github.com/spf13/cobra"
)

var downloadArtifactsCmd = &cobra.Command{
	Use:   "download-artifacts",
	Short: "Downloads the artifacts from the connector registry",
	Run:   runDownloadArtifactsCmd,
}

var downloadArtifactsCmdArgs ConnectorRegistryArgs

func init() {
	RootCmd.AddCommand(downloadArtifactsCmd)
	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH") // this file contains the list of changed files
	downloadArtifactsCmd.PersistentFlags().StringVar(&downloadArtifactsCmdArgs.ChangedFilesPath, "changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		downloadArtifactsCmd.MarkPersistentFlagRequired("changed-files-path")
	}
}

type ArtifactDownloadOptions struct {
	ChangedFilesPath string
	SingleFilePath   string
}

type ArtifactOption func(*ArtifactDownloadOptions)

func WithChangedFilesPath(path string) ArtifactOption {
	return func(o *ArtifactDownloadOptions) {
		o.ChangedFilesPath = path
	}
}

func WithSingleFile(path string) ArtifactOption {
	return func(o *ArtifactDownloadOptions) {
		o.SingleFilePath = path
	}
}

func DownloadArtifacts(opts ...ArtifactOption) ([]*ndchub.ConnectorArtifacts, error) {
	options := &ArtifactDownloadOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if options.ChangedFilesPath == "" && options.SingleFilePath == "" {
		return nil, fmt.Errorf("at least one of ChangedFilesPath or SingleFilePath must be provided")
	}

	if options.ChangedFilesPath != "" {
		connectorPackagingFiles, err := getConnectorPackagingFilesFromChangedFiles(options.ChangedFilesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get connector packaging files from changed files: %w", err)
		}
		return downloadArtifacts(connectorPackagingFiles)
	} else {
		// If a single file path is provided, we can directly download the artifacts for that file
		artifacts, err := downloadArtifactsUtil(options.SingleFilePath)
		artifactsArr := []*ndchub.ConnectorArtifacts{artifacts}
		return artifactsArr, err
	}
}

func getChangedFiles(filepath string) (*ChangedFiles, error) {
	changedFilesContent, err := os.Open(filepath)
	if err != nil {
		// log.Fatalf("Failed to open the file: %v, err: %v", ciCmdArgs.ChangedFilesPath, err)
		return nil, fmt.Errorf("failed to open the file: %v, err: %w", downloadArtifactsCmdArgs.ChangedFilesPath, err)
	}
	defer changedFilesContent.Close()

	// Read the changed file's contents. This file contains all the changed files in the PR
	changedFilesByteValue, err := io.ReadAll(changedFilesContent)
	if err != nil {
		// log.Fatalf("Failed to read the changed files JSON file: %v", err)
		return nil, fmt.Errorf("failed to read the changed files JSON file: %w", err)
	}

	var changedFiles *ChangedFiles = &ChangedFiles{}
	err = json.Unmarshal(changedFilesByteValue, changedFiles)
	if err != nil {
		// log.Fatalf("Failed to unmarshal the changed files content: %v", err)
		return nil, fmt.Errorf("failed to unmarshal the changed files content: %w", err)
	}

	return changedFiles, nil
}

// for now, since we're only adding support for newly added connectors, we will only check for the added files
// in the added files, we only need files that end with connector-packaging.json
func filterConnectorPackagingFiles(changedFiles *ChangedFiles) []string {
	// Filter the changed files to only include the ones that are in the connector registry
	var filteredChangedFiles []string = make([]string, 0)
	for _, file := range changedFiles.Added {
		if isConnectorPackagingFile(file) {
			filteredChangedFiles = append(filteredChangedFiles, file)
		}
	}

	log.Printf("Filtered connector packaging files: %v", filteredChangedFiles)
	return filteredChangedFiles
}

func isConnectorPackagingFile(path string) bool {
	return filepath.Base(path) == "connector-packaging.json"
}

func getConnectorPackaging(filePath string) (*ndchub.ConnectorPackaging, error) {
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join("../", filePath) // the filepaths returned by the changed files action are relative to the root of the repository. Therefore, we need to prepend "../" to the path to get the correct path
	}
	ndcHubConnectorPackaging, err := ndchub.GetConnectorPackaging(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get the connector packaging for file %s: %w", filePath, err)
	}
	return ndcHubConnectorPackaging, nil
}

func getConnectorPackagingFilesFromChangedFiles(changedFilesPath string) ([]string, error) {
	// Get the changed files from the PR
	changedFiles, err := getChangedFiles(changedFilesPath)
	if err != nil {
		return nil, err
	}

	// Get the connector packaging files (connector-packaging.json) from the changed files
	connectorPackagingFiles := filterConnectorPackagingFiles(changedFiles)
	return connectorPackagingFiles, nil
}

func downloadArtifacts(connectorPackagingFiles []string) ([]*ndchub.ConnectorArtifacts, error) {
	artifactList := make([]*ndchub.ConnectorArtifacts, 0)
	for _, file := range connectorPackagingFiles {
		log.Printf("\n\n") // for more readable logs
		artifacts, err := downloadArtifactsUtil(file)
		if err != nil {
			return nil, fmt.Errorf("failed to download artifacts for file %s: %w", file, err)
		}
		artifactList = append(artifactList, artifacts)
	}

	return artifactList, nil
}

func downloadArtifactsUtil(connectorPackagingFilePath string) (*ndchub.ConnectorArtifacts, error) {
	connectorPackaging, err := getConnectorPackaging(connectorPackagingFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get the connector packaging: %w", err)
	}

	connectorMetadata, _, extractedTgzPath, err := ndchub.GetPackagingSpec(connectorPackaging.URI, 
		connectorPackaging.Namespace,
		connectorPackaging.Name, 
		connectorPackaging.Version,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get connector metadata for %s/%s:%s: %w",
			connectorPackaging.Namespace, connectorPackaging.Name, connectorPackaging.Version, err)
	}

	artifacts, err := connectorMetadata.GetArtifacts(extractedTgzPath)	
	if err != nil {
		return nil, fmt.Errorf("failed to get the artifacts for %s/%s:%s: %w",
			connectorPackaging.Namespace, connectorPackaging.Name, connectorPackaging.Version, err)
	}
	
	return artifacts, nil
}

func runDownloadArtifactsCmd(cmd *cobra.Command, args []string) {
	artifacts, err := DownloadArtifacts(WithChangedFilesPath(downloadArtifactsCmdArgs.ChangedFilesPath))
	if err != nil {
		fmt.Printf("Failed to download artifacts: %v\n", err)
		os.Exit(1)
	}

	// Print artifactList as JSON
	artifactListJSON, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal artifact list to JSON: %v\n", err)
		return
	}
	fmt.Println(string(artifactListJSON))
}
