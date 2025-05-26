package scan

import (
	"github.com/hasura/ndc-hub/registry-automation/cmd"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the connector images",
	Long: `Scan the connector images for vulnerabilities and compliance issues.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},

}

func init() {
	cmd.RootCmd.AddCommand(scanCmd)
}
