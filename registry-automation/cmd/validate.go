package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"github.com/hasura/ndc-hub/registry-automation/pkg/validate"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the contents of ndc-hub",
	Run:   executeValidateCmd,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func executeValidateCmd(cmd *cobra.Command, args []string) {
	ndcHubGitRepoFilePath := os.Getenv("NDC_HUB_GIT_REPO_FILE_PATH")
	if ndcHubGitRepoFilePath == "" {
		fmt.Println("please set a value for NDC_HUB_GIT_REPO_FILE_PATH env var")
		os.Exit(1)
		return
	}

	registryFolder := filepath.Join(ndcHubGitRepoFilePath, "registry")
	_, err := os.Stat(registryFolder)
	if err != nil {
		fmt.Println("error while finding the registry folder", err)
		os.Exit(1)
		return
	}
	if os.IsNotExist(err) {
		fmt.Println("registry folder does not exist")
		os.Exit(1)
		return
	}

	type connectorPackaging struct {
		filePath         string
		connectorPackage *ndchub.ConnectorPackaging
	}
	var connectorPkgs []connectorPackaging

	// Track all connectors to validate their versions
	type connectorMetadataWithVersion struct {
		connector     Connector
		latestVersion string
	}
	var connectorsToValidate []connectorMetadataWithVersion

	err = filepath.WalkDir(registryFolder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Base(path) == ndchub.ConnectorPackagingJSON {
			cp, err := ndchub.GetConnectorPackaging(path)
			if err != nil {
				return err
			}
			if cp != nil {
				connectorPkgs = append(connectorPkgs, connectorPackaging{filePath: path, connectorPackage: cp})
			}
		}

		// Check for metadata.json files
		if filepath.Base(path) == "metadata.json" && !strings.Contains(path, "aliased_connectors") {
			metadata, err := os.ReadFile(path)
			var connectorMetadata ConnectorMetadata
			err = json.Unmarshal(metadata, &connectorMetadata)
			if err != nil {
				return fmt.Errorf("failed to read metadata.json at %s: %v", path, err)
			}

			if err != nil {
				return fmt.Errorf("failed to read metadata.json at %s: %v", path, err)
			}

			// Get namespace and name from path
			connectorFolder := filepath.Dir(path)
			namespaceFolder := filepath.Dir(connectorFolder)

			connectorsToValidate = append(connectorsToValidate, connectorMetadataWithVersion{
				connector: Connector{
					Name:      filepath.Base(connectorFolder),
					Namespace: filepath.Base(namespaceFolder),
				},
				latestVersion: connectorMetadata.Overview.LatestVersion,
			})
		}

		return nil
	})
	if err != nil {
		fmt.Println("error while walking the registry folder", err)
		os.Exit(1)
		return
	}

	hasError := false

	fmt.Println("Validating `connector-packaging.json` contents")
	for _, cp := range connectorPkgs {
		println("validating connector packaging for", cp.connectorPackage.Namespace, cp.connectorPackage.Name, "with version", cp.connectorPackage.Version)
		err := validate.ConnectorPackaging(cp.connectorPackage)
		if err != nil {
			fmt.Println("error validating connector packaging", cp.filePath, err)
			hasError = true
		}
	}
	fmt.Println("Completed validating `connector-packaging.json` contents")

	fmt.Println("Validating latest versions in metadata.json")
	for _, cm := range connectorsToValidate {
		println("validating latest version for", cm.connector.Namespace, cm.connector.Name, "with version", cm.latestVersion)
		err := validate.ValidateLatestVersion(cm.connector.Name, cm.connector.Namespace, cm.latestVersion)
		if err != nil {
			fmt.Printf("error validating latest version for %s/%s: %v\n",
				cm.connector.Namespace, cm.connector.Name, err)
			hasError = true
		}
	}
	fmt.Println("Completed validating latest versions")

	if hasError {
		fmt.Println("Exiting with a non-zero error code due to the error(s) in validation")
		os.Exit(1)
	}
}
