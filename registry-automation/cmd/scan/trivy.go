package scan

import (
	"fmt"
	"os"

	"github.com/hasura/ndc-hub/registry-automation/cmd"
	"github.com/hasura/ndc-hub/registry-automation/pkg/vulnerabilityscan"
	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"github.com/spf13/cobra"
)

var trivyCmd = &cobra.Command{
	Use:   "trivy",
	Short: "Scan the connector images using Trivy",
	Long:  `Scan the connector images for vulnerabilities and compliance issues using Trivy.`,
	Run: runTrivyCmd,
}

var trivyCmdArgs = struct {
	ChangedFilesPath string
}{}

func init() {
	// Add the trivy command to the scan command
	scanCmd.AddCommand(trivyCmd)

	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH") // this file contains the list of changed files
	trivyCmd.PersistentFlags().StringVar(&trivyCmdArgs.ChangedFilesPath, "changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		trivyCmd.MarkPersistentFlagRequired("changed-files-path")
	}
}

func downloadArtifacts(changedFilesPath string) ([]*ndchub.ConnectorArtifacts, error)  {
	artifacts, err := cmd.DownloadArtifacts(cmd.WithChangedFilesPath(changedFilesPath))
	if err != nil {
		return nil, fmt.Errorf("failed to download artifacts: %w", err)
	}
	return artifacts, nil
}

func scanConnectorImages(artifacts []*ndchub.ConnectorArtifacts) {
	for _, artifact := range artifacts {
		vulnerabilityscan.ScanArtifacts(artifact)
	}
}

func runTrivyCmd(cmd *cobra.Command, args []string) {
	changedFilesPath := trivyCmdArgs.ChangedFilesPath
	if changedFilesPath == "" {
		fmt.Println("No changed files path provided. Please set the CHANGED_FILES_PATH environment variable or use the --changed-files-path flag.")
		os.Exit(1)
	}

	// Download artifacts based on changed files
	artifacts, err := downloadArtifacts(changedFilesPath)
	if err != nil {
		fmt.Printf("Error downloading artifacts: %v\n", err)
		os.Exit(1)
	}

	// Scan each connector image for vulnerabilities
	scanConnectorImages(artifacts)

	fmt.Println("All connector images scanned successfully.")
}