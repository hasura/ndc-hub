package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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

	type connectorMetadata struct {
		filepath string
		metadata *ndchub.ConnectorMetadata
	}
	var allConnectorMetadata []connectorMetadata

	type connectorPackaging struct {
		filePath         string
		connectorPackage *ndchub.ConnectorPackaging
	}
	var connectorPkgs []connectorPackaging

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
		if filepath.Base(path) == ndchub.MetadataJSON {
			cm, err := ndchub.GetConnectorMetadata(path)
			if err != nil {
				return err
			}
			if cm != nil {
				allConnectorMetadata = append(allConnectorMetadata, connectorMetadata{filepath: path, metadata: cm})
			}
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
	for _, cm := range allConnectorMetadata {
		var respectiveConnectorPkgs []ndchub.ConnectorPackaging
		for _, cp := range connectorPkgs {
			if cp.connectorPackage.Namespace == cm.metadata.Overview.Namespace && cp.connectorPackage.Name == cm.metadata.Overview.Name {
				respectiveConnectorPkgs = append(respectiveConnectorPkgs, *cp.connectorPackage)
			}
		}

		err := validate.Metadata(cm.metadata, respectiveConnectorPkgs)
		if err != nil {
			fmt.Println("error validating connector packaging", cm.filepath, err)
			hasError = true
		}
	}
	fmt.Println("Completed validating latest versions")

	if hasError {
		fmt.Println("Exiting with a non-zero error code due to the error(s) in validation")
		os.Exit(1)
	}
}
