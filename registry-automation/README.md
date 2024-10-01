# Introduction

## Steps to runs

1. Consider the following `changed_files.json` file:
```json

{
    "added_files": [
        "registry/hasura/azure-cosmos/releases/v0.1.6/connector-packaging.json"
    ],
    "modified_files": [
        "registry/hasura/azure-cosmos/metadata.json"
    ],
    "deleted_files": []
}
```

2. You will require the following environment variables:

1. GCP_BUCKET_NAME
2. CLOUDINARY_URL
3. GCP_SERVICE_ACCOUNT_KEY
4. CONNECTOR_REGISTRY_GQL_URL
5. CONNECTOR_PUBLICATION_KEY
6. GCP_SERVICE_ACCOUNT_DETAILS



```bash


2. Run the following command from the `registry-automation` directory:


```bash
go run main.go ci --changed-files-path changed_files.json
```
