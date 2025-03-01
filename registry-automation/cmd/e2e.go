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

var e2eCmd = &cobra.Command{
	Use:   "e2e",
	Short: "Provides e2e testing for the connector release",
}

var e2eChangedCmdArgs ConnectorRegistryArgs
var e2eChangedCmd = &cobra.Command{
	Use:   "changed",
	Short: "Outputs the changed connector releases to test",
	Run:   e2eChanged,
}

var e2eLatestRegistryDirArg string
var e2eLatest = &cobra.Command{
	Use:   "latest",
	Short: "Outputs the latest connector releases to test",
	Run:   e2eLatestFunc,
}

var e2eAllRegistryDirArg string
var e2eAll = &cobra.Command{
	Use:   "all",
	Short: "Outputs all connector releases to test",
	Run:   e2eAllFunc,
}

func init() {

	// Path for the changed files in the PR
	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH")
	e2eChangedCmd.PersistentFlags().StringVar(&e2eChangedCmdArgs.ChangedFilesPath, "changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		e2eChangedCmd.MarkPersistentFlagRequired("changed-files-path")
	}

	// Path to the registry directory
	registryDirectoryEnv := os.Getenv("REGISTRY_DIRECTORY")
	e2eLatest.PersistentFlags().StringVar(&e2eLatestRegistryDirArg, "registry-directory",
		registryDirectoryEnv, "path to the ndc-hub registry directory")
	if registryDirectoryEnv == "" {
		e2eLatest.MarkPersistentFlagRequired("registry-directory")
	}
	e2eAll.PersistentFlags().StringVar(&e2eAllRegistryDirArg, "registry-directory",
		registryDirectoryEnv, "path to the ndc-hub registry directory")
	if registryDirectoryEnv == "" {
		e2eAll.MarkPersistentFlagRequired("registry-directory")
	}

	e2eCmd.AddCommand(e2eChangedCmd, e2eLatest, e2eAll)

	rootCmd.AddCommand(e2eCmd)
}

func e2eChanged(cmd *cobra.Command, args []string) {
	changedFilesContent, err := os.Open(e2eChangedCmdArgs.ChangedFilesPath)
	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", e2eChangedCmdArgs.ChangedFilesPath, err)
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

	// Collect the added or modified connectors
	processChangedFiles := processChangedFiles(changedFiles)
	out := make([]E2EOutput, 0)
	for connector, versions := range processChangedFiles.NewConnectorVersions {
		for version, connectorPackagingPath := range versions {
			testConfigPath := getTestConfigPath(connectorPackagingPath)
			if testConfigPath == "" {
				log.Printf("test config path is empty for %v, ignoring", connectorPackagingPath)
				continue
			}
			out = append(out, E2EOutput{
				Namespace:          connector.Namespace,
				ConnectorName:      connector.Name,
				ConnectorVersion:   version,
				TestConfigFilePath: testConfigPath,
			})
		}
	}
	printE2EOutput(out)
}

func e2eAllFunc(cmd *cobra.Command, args []string) {
	registryDir, err := filepath.Abs(e2eAllRegistryDirArg)
	if err != nil {
		log.Fatalf("Failed to get the absolute path for the registry directory: %v", err)
	}
	out := make([]E2EOutput, 0)
	if err := filepath.WalkDir(registryDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == ndchub.ConnectorPackagingJSON {
			e2eOutput := getE2EOutput(path)
			if e2eOutput != nil {
				out = append(out, *e2eOutput)
			}
		}
		return nil
	}); err != nil {
		log.Fatalf("Failed to walk the registry directory: %v", err)
	}
	printE2EOutput(out)
}

func e2eLatestFunc(cmd *cobra.Command, args []string) {
	registryDir, err := filepath.Abs(e2eLatestRegistryDirArg)
	if err != nil {
		log.Fatalf("Failed to get the absolute path for the registry directory: %v", err)
	}
	out := make([]E2EOutput, 0)
	registries, err := os.ReadDir(registryDir)
	if err != nil {
		log.Fatalf("Failed to read the registry root directory: %v", err)
	}
	for _, registry := range registries {
		if !registry.IsDir() {
			continue
		}
		registryDir = filepath.Join(registryDir, registry.Name())
		connectors, err := os.ReadDir(registryDir)
		if err != nil {
			log.Fatalf("Failed to read the registry directory: %v", err)
		}
		for _, connector := range connectors {
			if !connector.IsDir() {
				continue
			}
			connectorDir := filepath.Join(registryDir, connector.Name())
			metadataPath := filepath.Join(connectorDir, ndchub.MetadataJSON)
			cm, err := ndchub.GetConnectorMetadata(metadataPath)
			if err != nil {
				log.Fatalf("Failed to get connector metadata: %v", err)
			}
			if cm == nil {
				log.Printf("Connector metadata is nil for %v", metadataPath)
				continue
			}
			latestVersion := cm.Overview.LatestVersion
			latestVersionCPPath := filepath.Join(connectorDir, "releases", latestVersion, ndchub.ConnectorPackagingJSON)
			e2eOutput := getE2EOutput(latestVersionCPPath)
			if e2eOutput != nil {
				out = append(out, *e2eOutput)
			}

		}
	}
	printE2EOutput(out)
}

func printE2EOutput(out []E2EOutput) {
	outBytes, err := json.Marshal(out)
	if err != nil {
		log.Fatalf("Failed to marshal e2e outoput: %v", err)
	}
	fmt.Fprintln(os.Stdout, string(outBytes))
}

func getTestConfigPath(connectorPackagingPath string) string {
	cp, err := ndchub.GetConnectorPackaging(connectorPackagingPath)
	if err != nil {
		log.Fatalf("Failed to get connector packaging: %v", err)
	}
	if cp == nil {
		log.Printf("connector packaging is nil for %v, ignoring", connectorPackagingPath)
		return ""
	}
	return cp.GetTestConfigPath()
}

func getE2EOutput(connectorPackagingPath string) *E2EOutput {

	testConfigPath := getTestConfigPath(connectorPackagingPath)
	if testConfigPath == "" {
		log.Printf("test config path is empty for %v, ignoring", connectorPackagingPath)
		return nil
	}
	// path looks like this: /some/folder/ndc-hub/registry/hasura/turso/releases/v0.1.0/connector-packaging.json
	versionFolder := filepath.Dir(connectorPackagingPath)
	releasesFolder := filepath.Dir(versionFolder)
	connectorFolder := filepath.Dir(releasesFolder)
	namespaceFolder := filepath.Dir(connectorFolder)

	return &E2EOutput{
		Namespace:          filepath.Base(namespaceFolder),
		ConnectorName:      filepath.Base(connectorFolder),
		ConnectorVersion:   filepath.Base(versionFolder),
		TestConfigFilePath: testConfigPath,
	}
}
