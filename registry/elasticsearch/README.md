# Hasura Elasticsearch Connector

## Get started
This guide will help you perform the setup of this connector in local

### Clone Repository
Clone connector from [ndc-elasticsearch](https://github.com/hasura/ndc-elasticsearch/) github repository
```
git clone https://github.com/hasura/ndc-elasticsearch.git
```

### Set Environment Variables
Set the following environment variables for connector
```
ELASTICSEARCH_URL=${YOUR_ELASTICSEARCH_URL}
ELASTICSEARCH_USERNAME=${YOUR_ELASTICSEARCH_USERNAME}
ELASTICSEARCH_PASSWORD=${YOUR_ELASTICSEARCH_PASSWORD}
ELASTICSEARCH_API_KEY=${YOUR_ELASTICSEARCH_API_KEY}
ELASTICSEARCH_CA_CERT_PATH=${YOUR_ELASTICSEARCH_CA_CERT_PATH}
ELASTICSEARCH_INDEX_PATTERN=${REGEX_PATTERN_OF_INDICES}
```
Note: One can set either username and password or api key in env variables for authentication.

### Build connector executable file
```go
go build
```

### Update configuration
Run the update cli command to update the `configuration.json` file

```
./ndc-elasticsearch update
```

### Run the connector locally
Run the following command to start your connector

```
./ndc-elasticsearch serve
```
Connector server will be up and running at http://localhost:8080

### Verify
```
curl http://localhost:8080/schema
```
Send request to `query` endpoint for queries
