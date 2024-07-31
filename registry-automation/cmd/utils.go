package cmd

import "fmt"

func generateGCPObjectName(namespace, connectorName, version string) string {
	return fmt.Sprintf("packages/%s/%s/%s/package.tgz", namespace, connectorName, version)
}
