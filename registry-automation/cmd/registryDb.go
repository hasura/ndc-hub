package cmd

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
)

type HubRegistryConnectorInsertInput struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Namespace string `json:"namespace"`
}

// struct to store the response of teh GetConnectorInfo query
type GetConnectorInfoResponse struct {
	HubRegistryConnector []struct {
		Name                 string `json:"name"`
		MultitenantConnector *struct {
			ID string `json:"id"`
		} `json:"multitenant_connector"`
	} `json:"hub_registry_connector"`
}

func insertHubRegistryConnector(client graphql.Client, connectorMetadata ConnectorMetadata, connector NewConnector) error {
	var respData map[string]interface{}

	ctx := context.Background()

	req := graphql.NewRequest(`
mutation InsertHubRegistryConnector ($connector:hub_registry_connector_insert_input!){
  insert_hub_registry_connector_one(object: $connector) {
    name
    title
  }
}`)
	hubRegistryConnectorInsertInput := HubRegistryConnectorInsertInput{
		Name:      connector.Name,
		Title:     connectorMetadata.Overview.Title,
		Namespace: connector.Namespace,
	}

	req.Var("connector", hubRegistryConnectorInsertInput)
	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	} else {
		fmt.Printf("Successfully inserted the connector in the registry: %+v\n", respData)
	}

	return nil
}

func getConnectorInfoFromRegistry(client graphql.Client, connectorNamespace string, connectorName string) (GetConnectorInfoResponse, error) {
	var respData GetConnectorInfoResponse

	ctx := context.Background()

	req := graphql.NewRequest(`
query GetConnectorInfo ($name: String!, $namespace: String!) {
  hub_registry_connector(where: {_and: [{name: {_eq: $name}}, {namespace: {_eq: $namespace}}]}) {
    name
    multitenant_connector {
      id
    }
  }
}`)
	req.Var("name", connectorName)
	req.Var("namespace", connectorNamespace)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return respData, err
	} else {
		if len(respData.HubRegistryConnector) == 0 {
			return respData, nil
		}
	}

	return respData, nil
}

func updateRegistryGQL(client graphql.Client, payload []ConnectorVersion) error {
	var respData map[string]interface{}

	ctx := context.Background()

	req := graphql.NewRequest(`
mutation InsertConnectorVersion($connectorVersion: [hub_registry_connector_version_insert_input!]!) {
  insert_hub_registry_connector_version(objects: $connectorVersion, on_conflict: {constraint: connector_version_namespace_name_version_key, update_columns: [image, package_definition_url, is_multitenant]}) {
    affected_rows
    returning {
      id
    }
  }
}`)
	// add the payload to the request
	req.Var("connectorVersion", payload)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil
}

func updateConnectorOverview(updates ConnectorOverviewUpdates) error {
	var respData map[string]interface{}
	client := graphql.NewClient(ciCmdArgs.ConnectorRegistryGQLUrl)
	ctx := context.Background()

	req := graphql.NewRequest(`
mutation UpdateConnector ($updates: [connector_overview_updates!]!) {
  update_connector_overview_many(updates: $updates) {
    affected_rows
  }
}`)

	// add the payload to the request
	req.Var("updates", updates.Updates)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	} else {
		fmt.Printf("Successfully updated the connector overview: %+v\n", respData)
	}

	return nil
}
