# Connector Package Distribution RFC - Milestone 1

This is a Work-In-Progress document. Please provide any feedback you wish to contribute via Github comments and suggestions.


## Purpose

Connector API, definition and packaging are specified respectively by:

* [NDC specification](http://hasura.github.io/ndc-spec/)
* [Deployment Specification](https://github.com/hasura/ndc-hub/blob/main/rfcs/0000-deployment.md)
* [Packaging Specification (WIP)](https://github.com/hasura/ndc-hub/pull/89/files)
* [Umbrella Specification with the rest of the Roadmap](https://github.com/hasura/ndc-hub/pull/98)

This new distribution specification details the extensions to the connector registry metdata that are required for distribution of the new package definitions.

This proposal intends to allow the existing system to be extended to support new functionality and not require any breaking changes or downtime.

In addition, all "TODO" references should be replaced before finalization.


### Umbrella Delivery Roadmap

The delivery of a broader set of changes can be rolled out incrementally, and this can be paused or stopped at any stage without disruption to the current system.

#### Milestone 1 (this RFC) - Definition Links:

* Where there are currently git tag references:
	* Add link and checksum to package definition archive to DB Schema
	* Add link and checksum to package definition archive to Github metadata format

The initial change to the metadata format would look as follows for the [Postgres Connector](https://github.com/hasura/ndc-hub/blob/lyndon/distribution-rfc/registry/postgres/metadata.json):

```json
{
  "overview": { ... },
  "author": { ... },
  "is_verified": true,
  "is_hosted_by_hasura": true,
  // New stanza
  "packages": [
    {
      "version": "1.2.3",
      "uri": "https://foobar.com/releases/postgres-postgresql-v0.2.0-9283dh9283u...hd092ujdf2ued.tar.gz",
      "checksum": {
        "type": "sha256",
        "value": "9283dh9283u...hd092ujdf2ued"
      },
      // Optional link from package to source
      "source": {
        "hash": "98801634b0e1396c933188eef88178952f412a8c",
      }
    }
  ]
  "source_code": {
    "is_open_source": true,
    "repository": "https://github.com/hasura/ndc-postgres",
    "version": [
      {
        "tag": "v0.2.0",
        "hash": "98801634b0e1396c933188eef88178952f412a8c",
        "is_verified": true
      }
    ]
  }
}
```

While package definition releases can be hosted at any URL, some convenient locations could include:

* Under the github releases artefacts
* As a standalone github repository using the "download archive" feature
* In a Hasura Google cloud bucket (for Hasura authors)

We will establish conventions for this that make authoring as streamlined as possible.


#### Milestone 2 - Topic Tributary:

* Add a CI process to ingest connectors tagged with a topic in addition to registry connectors

#### Milestone 3 - User and Role updates:

* Extend the DB schema to include user/role information
* Set up auth to allow signup/login/token flows
* Extend the API permissions to allow role and resource based access to data
* Add validation actions to authoring flows

#### Milestone 4 - CLI:

* Create a new CLI plugin to manage interaction with the API


## Proposal

While the precursor specifications outline the structure and mechanisms of packaging, this RFC details how the packages are owned, distributed, and indexed. The solution is outlined from storage up to user-applications and how they can be leveraged by CI in order to automatically publish updated versions and scrape topics for community contribution discovery.

This solution enables the following UX scenarios:

* Packages are published by authors via Hasura CI
	* Deriving metadata and definitions from hub registry
	* Deriving metadata and definitions from Github topics
* Packages browsed and searched for by Hasura V3 users
* Packages are referenced in Hasura V3 projects
* Packages are fetched for local usage in Hasura V3 projects


### Ownership

For this milestone ownerships is restricted to Hasura, but PRs can be made to the hub registry, and repositories can use pre-defined topics to allow community contributions.


### Storage

Package definitions take the form described in the packaging spec. These need to be stored. The storage mechanism can be described abstractly:

* The storage system is philosophically idempotent and content addressable (hashes are included in assets).
* Stable read-only (fetch) URIs exist for the stored location of packages

While this abstract definition is useful for system-requirements, in practice our initial implementation will use Google Cloud Buckets for a centralised hasura publication, and Github releases for independent publication.

Storage conventions will be followed so that our system could initially predict the location of packages and we can incrementally transition to API based package access in service of rapid delivery.

Storage location convention should initially be: `ORG/PACKAGE/VERSION/SHA/org-packge-version-sha.tar.gz`, although this is not a system dependency.


### Database

The database backing the API provides all of the APIs state management capabilities outside of package archive storage (as described in "Storage").

The initial implementation of the Database will be an extension of the exiting hub registry Postgres instance.


### API

The API provides the user-interaction layer that mediates the database and storage components. No direct user interaction should occur with either the database, or storage, except for the case when the API delegates a storage interaction - such as providing a Hasura V3 project user a public storage URL for read access to a package definition.

The API will be implemented via a Hasura V2 instance.


### Indexing

The technical considerations were described in the "API" section, however, users should be able to leverage indexes on the following properties:

* Name
* Version
* Date of Publication
* Organisation
* Author
* Tags
* Checksum
* Category
* Free-text description
* Related packages
* Underlying DB (or Data solution)
* Any metadata in the Package description
* Verification status

as well as all the existing registry metadata, and all metadata included in the package definition.


### Checksums

All package definition references should be accompanied by a checksum in order to verify that the definition hasn't been changed in storage. Any definition fetch operation should verify the checksum.


### Versioning

Package Metadata will include a version. This should follow [Semantic Versioning](https://semver.org) practices.


### Discoverability

The discoverability component will simply hard-code various permutations of the "Indexing" criteria to provide the user browsable lists of packages.

These could include lists such as:

* Most popular packages
* Most recent packages
* Verified packages
* Etc.


### Access

Access is read-only by default, with only Hasura having write access to the registry, API, and database. Package authors may have write access to their definition storage such as Github releases.


### Publication

The publication of packages should be performed via PR to the ndc-hub registry.

Convenience interfaces could be developed to assist with this workflow.


### Verification

Verification status procedures and policies Organisation/Package should be defined further in subsequent specifications.

Verification information should describe what has been checked and to what it applies. A separate table should be kept for this information, not just a bit on existing resources.

This is for granularity and auditing purposes.


### System Abuse Scenarios and Mitigation

Any publicly accessible APIs with publication capabilities have the potential to be abused and as such we should attempt to predict and mitigate the scenarios that we can anticipate:

* Identity misappropriation
* Inappropriate content
* Leaked credentials
* Squatting
* Denial of service
* IP harvesting by third parties and competitors
* Incorrect application of verification
* Recycling of content
* Unintentional mistakes
* Spam / Reflection

