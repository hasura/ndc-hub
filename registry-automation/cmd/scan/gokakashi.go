package scan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hasura/ndc-hub/registry-automation/cmd"
	"github.com/hasura/ndc-hub/registry-automation/pkg/ndchub"
	"github.com/hasura/ndc-hub/registry-automation/pkg/vulnerabilityscan"
	"github.com/spf13/cobra"
)

var gokakashiCmd = &cobra.Command{
	Use:   "gokakashi",
	Short: "Scan the connector images using Gokakashi",
	Long:  `Scan the connector images for vulnerabilities and compliance issues using Gokakashi.`,
	Run:   runGokakashiCmd,
}

var gokakashiCmdArgs = struct {
	files                string
	server               string
	token                string
	policy               string
	cfAccessClientId     string
	cfAccessClientSecret string
	binaryPath           string
}{}

func init() {
	// Add the gokakashi command to the scan command
	scanCmd.AddCommand(gokakashiCmd)

	var files = os.Getenv("GOKAKASHI_FILES")                                     // this is the glob pattern to match files for scanning
	var serverEnv = os.Getenv("GOKAKASHI_SERVER")                                // this is the Gokakashi server URL
	var tokenEnv = os.Getenv("GOKAKASHI_TOKEN")                                  // this is the Gokakashi server token
	var policyEnv = os.Getenv("GOKAKASHI_POLICY")                                // this is the Gokakashi policy
	var cfAccessClientIdEnv = os.Getenv("GOKAKASHI_CF_ACCESS_CLIENT_ID")         // this is the Cloudflare Access Client ID
	var cfAccessClientSecretEnv = os.Getenv("GOKAKASHI_CF_ACCESS_CLIENT_SECRET") // this is the Cloudflare Access Client Secret
	var binaryPath = os.Getenv("GOKAKASHI_BINARY_PATH")                          // this is the path to the Gokakashi binary

	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.files, "files", files, "Glob pattern to match files for scanning (env: GOKAKASHI_FILES)")
	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.server, "server", serverEnv, "Server URL for Gokakashi (env: GOKAKASHI_SERVER)")
	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.token, "token", tokenEnv, "Token for Gokakashi server authentication (env: GOKAKASHI_TOKEN)")
	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.policy, "policy", policyEnv, "Policy for Gokakashi (env: GOKAKASHI_POLICY)")
	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.cfAccessClientId, "cf-access-client-id", cfAccessClientIdEnv, "Cloudflare Access Client ID for Gokakashi (env: GOKAKASHI_CF_ACCESS_CLIENT_ID)")
	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.cfAccessClientSecret, "cf-access-client-secret", cfAccessClientSecretEnv, "Cloudflare Access Client Secret for Gokakashi (env: GOKAKASHI_CF_ACCESS_CLIENT_SECRET)")
	gokakashiCmd.PersistentFlags().StringVar(&gokakashiCmdArgs.binaryPath, "binary-path", binaryPath, "(Optional) Path to the Gokakashi binary. Defaults to 'gokakashi' in the system PATH. (env: GOKAKASHI_BINARY_PATH)")

	if files == "" {
		gokakashiCmd.MarkPersistentFlagRequired("files")
	}
	if serverEnv == "" {
		gokakashiCmd.MarkPersistentFlagRequired("server")
	}
	if tokenEnv == "" {
		gokakashiCmd.MarkPersistentFlagRequired("token")
	}
	if policyEnv == "" {
		gokakashiCmd.MarkPersistentFlagRequired("policy")
	}
	if cfAccessClientIdEnv == "" {
		gokakashiCmd.MarkPersistentFlagRequired("cf-access-client-id")
	}
	if cfAccessClientSecretEnv == "" {
		gokakashiCmd.MarkPersistentFlagRequired("cf-access-client-secret")
	}
}

func runGokakashiCmd(cmd *cobra.Command, args []string) {
	files, err := getFiles(gokakashiCmdArgs.files)
	if err != nil {
		fmt.Printf("Error getting files: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("No files found matching the glob pattern. Please check the GOKAKASHI_FILES environment variable or --files flag.")
		os.Exit(1)
	}

	errors := []error{}
	shouldExitWithNonZero := false

	for _, file := range files {
		log.Printf("\n\n")
		artifacts, err := downloadArtifactsForGokakashi(file)
		if err != nil {
			err = fmt.Errorf("failed to download artifacts for file %s: %w. Skipping scan", file, err)
			log.Printf("%s", err)
			errors = append(errors, err)
			shouldExitWithNonZero = true
			continue
		}
		if len(artifacts) == 0 {
			// something weird happened, no artifacts downloaded
			err = fmt.Errorf("no artifacts downloaded for file %s. Skipping scan", file)
			log.Printf("%s", err)
			errors = append(errors, err)
			shouldExitWithNonZero = true
			continue
		}

		scanErrors, err := vulnerabilityscan.ScanDockerImageWithGokakashi(
			vulnerabilityscan.WithDockerImagesForScan(artifacts[0].DockerImages), // since we're scanning one file at a time, we can take the first artifact
			vulnerabilityscan.WithServer(gokakashiCmdArgs.server),
			vulnerabilityscan.WithToken(gokakashiCmdArgs.token),
			vulnerabilityscan.WithPolicy(gokakashiCmdArgs.policy),
			vulnerabilityscan.WithCfAccessClientId(gokakashiCmdArgs.cfAccessClientId),
			vulnerabilityscan.WithCfAccessClientSecret(gokakashiCmdArgs.cfAccessClientSecret),
			vulnerabilityscan.WithCustomBin(gokakashiCmdArgs.binaryPath),
		)
		if err != nil {
			fmt.Printf("error while running Gokakashi scan: %v\n", err)
			os.Exit(1)
		}
		if len(scanErrors) > 0 {
			shouldExitWithNonZero = true
			for _, err := range scanErrors {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\nErrors encountered during Gokakashi scan:\n\n")
		for _, err := range errors {
			fmt.Printf("%s\n\n", err.Error())
		}
	}

	if shouldExitWithNonZero {
		os.Exit(1)
	}
}

func downloadArtifactsForGokakashi(file string) ([]*ndchub.ConnectorArtifacts, error) {
	artifacts, err := cmd.DownloadArtifacts(cmd.WithSingleFile(file))
	if err != nil {
		return nil, fmt.Errorf("failed to download artifacts: %w", err)
	}
	return artifacts, nil
}

func getFiles(path string) ([]string, error) {
	matches, err := filepath.Glob(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve glob pattern: %w", err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no files matched pattern: %s", path)
	}
	absFiles := make([]string, 0, len(matches))
	for _, match := range matches {
		absPath, err := filepath.Abs(match)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for %s: %w", match, err)
		}
		absFiles = append(absFiles, absPath)
	}
	return absFiles, nil
}
