package main

import (
	"github.com/hasura/ndc-hub/registry-automation/cmd"
	_ "github.com/hasura/ndc-hub/registry-automation/cmd/scan" // Import scan package to register its commands=
)

func main() {
	cmd.Execute()
}
