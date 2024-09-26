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

type NewConnectorsInsertInput struct {
	HubRegistryConnectors []HubRegistryConnectorInsertInput `json:"hub_registry_connectors"`
	ConnectorOverviews    []ConnectorOverviewInsert         `json:"connector_overviews"`
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

func insertHubRegistryConnector(client graphql.Client, newConnectors NewConnectorsInsertInput) error {
	var respData map[string]interface{}

	ctx := context.Background()

	req := graphql.NewRequest(`
mutation InsertHubRegistryConnector ($hub_registry_connectors:[hub_registry_connector_insert_input!]!, $connector_overview_objects: [connector_overview_insert_input!]!){

  insert_hub_registry_connector(objects: $hub_registry_connectors) {
affected_rows
  }
  insert_connector_overview(objects: $connector_overview_objects) {
    affected_rows
  }
}
`)
	// add the payload to the request
	req.Var("hub_registry_connectors", newConnectors.HubRegistryConnectors)
	req.Var("connectors_overviews", newConnectors.ConnectorOverviews)

	// set the headers
	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	} else {
		connectorNames := make([]string, 0)
		for _, connector := range newConnectors.HubRegistryConnectors {
			connectorNames = append(connectorNames, fmt.Sprintf("%s/%s", connector.Namespace, connector.Name))
		}
		fmt.Printf("Successfully inserted the following connectors in the registry: %+v\n", connectorNames)
	}

	return nil
}

func getConnectorInfoFromRegistry(client GraphQLClientInterface, connectorNamespace string, connectorName string) (GetConnectorInfoResponse, error) {
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

type ConnectorOverviewInsert struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Logo        string `json:"logo"`
	Docs        string `json:"docs"`
	IsVerified  bool   `json:"is_verified"`
	IsHosted    bool   `json:"is_hosted_by_hasura"`
	Author      struct {
		Data ConnectorAuthor `json:"data"`
	} `json:"author"`
}

type ConnectorAuthor struct {
	Name         string `json:"name"`
	SupportEmail string `json:"support_email"`
	Website      string `json:"website"`
}

// registryDbMutation is a function to insert data into the registry database, all the mutations are done in a single transaction.
func registryDbMutation(client GraphQLClientInterface, newConnectors NewConnectorsInsertInput, connectorOverviewUpdates []ConnectorOverviewUpdate, connectorVersionInserts []ConnectorVersion) error {
	var respData map[string]interface{}
	ctx := context.Background()
	mutationQuery := `
mutation HubRegistryMutationRequest (
  $hub_registry_connectors:[hub_registry_connector_insert_input!]!,
  $connector_overview_inserts: [connector_overview_insert_input!]!,
  $connector_overview_updates: [connector_overview_updates!]!,
  $connector_version_inserts: [hub_registry_connector_version_insert_input!]!
){

  insert_hub_registry_connector(objects: $hub_registry_connectors) {
affected_rows
  }
  insert_connector_overview(objects: $connector_overview_inserts) {
    affected_rows
  }
  insert_hub_registry_connector_version(objects: $connector_version_inserts, on_conflict: {constraint: connector_version_namespace_name_version_key, update_columns: [image, package_definition_url, is_multitenant]}) {
    affected_rows
  }

  update_connector_overview_many(updates: $connector_overview_updates) {
    affected_rows
  }
}
`
	req := graphql.NewRequest(mutationQuery)
	req.Var("hub_registry_connectors", newConnectors.HubRegistryConnectors)
	req.Var("connector_overview_inserts", newConnectors.ConnectorOverviews)
	req.Var("connector_overview_updates", connectorOverviewUpdates)
	req.Var("connector_version_inserts", connectorVersionInserts)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil

}

// registryDbMutation is a function to insert data into the registry database, all the mutations are done in a single transaction.
func registryDbMutationStaging(client GraphQLClientInterface, newConnectors NewConnectorsInsertInput, connectorOverviewUpdates []ConnectorOverviewUpdate, connectorVersionInserts []ConnectorVersion) error {
	var respData map[string]interface{}
	ctx := context.Background()
	mutationQuery := `
mutation HubRegistryMutationRequest (
  $hub_registry_connectors:[hub_registry_connector_insert_input!]!,
  $connector_overview_inserts: [connector_overview_insert_input!]!,
  $connector_overview_updates: [connector_overview_updates!]!,
  $connector_version_inserts: [hub_registry_connector_version_insert_input!]!
){

  insert_hub_registry_connector(objects: $hub_registry_connectors, on_conflict: {constraint: connector_pkey}) {
affected_rows
  }
  insert_connector_overview(objects: $connector_overview_inserts, on_conflict: {constraint: connector_overview_pkey}) {
    affected_rows
  }
  insert_hub_registry_connector_version(objects: $connector_version_inserts, on_conflict: {constraint: connector_version_namespace_name_version_key, update_columns: [image, package_definition_url, is_multitenant]}) {
    affected_rows
  }

  update_connector_overview_many(updates: $connector_overview_updates) {
    affected_rows
  }
}
`
	req := graphql.NewRequest(mutationQuery)
	req.Var("hub_registry_connectors", newConnectors.HubRegistryConnectors)
	req.Var("connector_overview_inserts", newConnectors.ConnectorOverviews)
	req.Var("connector_overview_updates", connectorOverviewUpdates)
	req.Var("connector_version_inserts", connectorVersionInserts)

	req.Header.Set("x-hasura-role", "connector_publishing_automation")
	req.Header.Set("x-connector-publication-key", ciCmdArgs.ConnectorPublicationKey)

	// Execute the GraphQL query and check the response.
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil

}
