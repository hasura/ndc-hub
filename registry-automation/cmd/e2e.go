package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

func e2eInput(cmd *cobra.Command, args []string) {
	changedFilesContent, err := os.Open(e2eCmdArgs.ChangedFilesPath)
	if err != nil {
		log.Fatalf("Failed to open the file: %v, err: %v", e2eCmdArgs.ChangedFilesPath, err)
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
		for version, _ := range versions {
			out = append(out, E2EOutput{
				SelectorPattern:  fmt.Sprintf("%s/%s", connector.Namespace, connector.Name),
				ConnectorVersion: version,
			})
		}
	}
	outBytes, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal e2e outoput: %v", err)
	}
	fmt.Fprintln(os.Stdout, string(outBytes))
}

var e2eCmd = &cobra.Command{
	Use:   "e2e",
	Short: "Provides end-to-end testing for the connector release",
	Run:   e2eInput,
}

var e2eCmdArgs ConnectorRegistryArgs

func init() {
	rootCmd.AddCommand(e2eCmd)

	// Path for the changed files in the PR
	var changedFilesPathEnv = os.Getenv("CHANGED_FILES_PATH")
	e2eCmd.PersistentFlags().StringVar(&e2eCmdArgs.ChangedFilesPath, "changed-files-path", changedFilesPathEnv, "path to a line-separated list of changed files in the PR")
	if changedFilesPathEnv == "" {
		e2eCmd.MarkPersistentFlagRequired("changed-files-path")
	}
}
