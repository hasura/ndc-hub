package cmd

import "fmt"

func generateGCPObjectName(connectorName, version string) string {
	return fmt.Sprintf("packages/%s/%s/package.tgz", connectorName, version)
}
