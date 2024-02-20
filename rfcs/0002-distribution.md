# Connector Package Distribution RFC

This is a Work-In-Progress document. Please provide any feedback you wish to contribute via Github comments and suggestions.


## Purpose

Connector API, definition and packaging are specified respectively by:

* [NDC specification](http://hasura.github.io/ndc-spec/)
* [Deployment Specification](https://github.com/hasura/ndc-hub/blob/main/rfcs/0000-deployment.md)
* [Packaging Specification (WIP)](https://github.com/hasura/ndc-hub/pull/89/files)

This new distribution specification details how connector packages are intended to be owned, stored, indexed, searched, fetched and automatically published.

The intuition for this system is inspired by other package management systems such as NPM, Cabal, etc.

There was a previous implementation of these concepts as described (TODO: Get docs links from Shraddha)

This proposal intends to allow the existing system to be extended to support new functionality and not require a migration.


### Items Outstanding in this Specification (TODO)

The following items are intended to be fleshed-out in this specification prior to approval:

* Data-formats
* URI locations
* Identifiers
* Assignment of implementation
* Dependencies
* Revocation concerns

In addition, all "TODO" references should be replaced before finalization.


### Delivery Roadmap

The delivery of the changes outlined in this RFC can be rolled out incrementally, and this can be paused or stopped at any stage without disruption to the current system.

#### Milestone 1 - Definition Links:

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


### Follow-Up Work

After implementation of this system, follow-up actions should be performed:

* Publication of package definition archives for all connectors
* Publication of new versions of all available connectors linking to archives


### Out of Scope for this RFC

The following are not described in this specification:

* The format of the metadata used for indexing, etc
* Authentication mechanisms
* Verification policies and procedures


## Motivation and Why and How this Differs from Existing Solutions

First, where are the current and proposed system different?

| Layer | Current System | Proposed System | Difference |
| ---   | ---            | ---             | ---        |
| Package Definition Storage | Github Tag | .tar.gz | Currently the package definitions are stored in a directory hierarchy on a github branch/tag  |
| Database | Postgres | Postgres | Same infrastructure with a new schema |
| API | Hasura | Hasura | Same infrastructure with different roles and actions |
| CLI | N/A | Hasura V3 CLI Plugin | No current CLI interactions available |
| Registry | ndc-hub/registry | No central component in proposal | There is no central definition source outside of the DB/Storage layer in the proposal |
| Third-party CI | No current component | Usage via CLI/API | Third-parties are able to integrate into their CI via API |
| Topic Scraping | No current component | Scheduled/Webhook trigger of topic ingestion of community connectors | Arbitrary ingestion is possible via the API/CLI |

The current system's GraphQL: https://data.pro.arusah.com/ - Looking at all the graphql queries starting with connector_ should give all the things that we have so far.

Google cloud function (https://github.com/hasura/connectors-cloud-integration/tree/main/sync_connector_hub_data) runs every 24 hours and scrapes registry and inserts into DB.

Issues with the current system:

* Pull-based ingestion of registry - Changes should ideally propagate instantaneously
* No ability to treat connectors independently - Need to PR to central registry
* Artefacts are not distributable outside of Github references - Tied to Github
* TODO


## Proposal

While the precursor specifications outline the structure and mechanisms of packaging, this RFC details how the packages are owned, distributed, and indexed. A layered solution is outlined from storage up to user-applications and how they can be leveraged by CI in order to automatically publish updated versions and scrape topics for community contribution discovery.

This solution enables the following UX scenarios:

* Authors are granted system credentials and API tokens
* Packages are manually published via CLI by authors along with metadata
* Packages are automatically published by authors via CI and API in connector repositories
* Packages browsed and searched for by Hasura V3 users
* Packages are referenced in Hasura V3 projects
* Packages are fetched for local usage in Hasura V3 projects


### Ownership

Ownership is granted on an Organisation -> Package -> Author -> Version hierarchy.

Users will be granted roles that authorize operations withing this hierarchy.

An initial draft of roles (from most, to least privileged) is:

* Operations - Global Access for System Operations
* Auditor - Global Read-Only Access
* Admin - Organisation Administrator - Create Packages
* Owner - Package Administrator
* Author - Package Contributor via Releases
* Public - General Public (default)

All roles (except auditor) can grant lesser role privileges to users in their domains.

Only Operations can grant the Operations role.

All changes within the system are logged and can be viewed by the auditor role.


### Storage

Package definitions take the form described in the packaging spec. These need to be stored. The storage mechanism can be described abstractly:

* The storage system is philosophically idempotent and content addressable (hashes are included in assets).
* Upload is available to the system internally, and able to be delegated to authors via pre-authorized URIs.
* Stable read-only (fetch) URIs exist for the stored location of packages
* Indexes are maintained outside of storage, but minimal metadata is maintained for system-administration purposes

While this abstract definition is useful for system-requirements, in practice our initial implementation will use Google Cloud Buckets.

Storage conventions will be followed so that our system could initially predict the location of packages and we can incrementally transition to API based package access in service of rapid delivery.

Storage location convention will initially be: `ORG/PACKAGE/VERSION/SHA/org-packge-version-sha.tar.gz`


### Database

The database backing the API provides all of the APIs state management capabilities outside of package archive storage (as described in "Storage").

The initial implementation of the Database will be Postgres.


### API

The API provides the user-interaction layer that mediates the database and storage components. No direct user interaction should occur with either the database, or storage, except for the case when the API delegates a storage interaction - such as providing an author a pre-authorized URL for publication, or providing a Hasura V3 project user a public storage URL for a package definition.

The API will be implemented via a Hasura V3 instance. (TODO: Check if V3 has the capabilities to implement this yet, or if we should start with a V2 instance for stability reasons)

The various functions of the API are described as follows:

#### Operation

* System health monitoring
* Restarts
* Resource allocation

#### Administration

* Creation of organisations
* Creation of users
* Creation of packages
* Assignment of roles
* Verification
* Revocation of content
* Redirection of resources

#### Authoring

* Creation of packages
* Publication of new versions of packages
* Association of new metadata with organisation/package/versions
* Request for verification
* Revocation (requests?) of package versions

#### Discoverability

* Search for package by metadata filters

#### Acquisition

* Request for package download URI


### Applications / CLI

The Hasura V3 CLI will provide a consistent and convenient interface to interacting with the API.

The CLI commands will include the following:

* hasura3 package create
* hasura3 package publish
* hasura3 package revoke
* hasura3 package search
* hasura3 package fetch


### CI

Contributors may leverage the API or CLI in their CI workflows in order to automate the publication of new versions of their packages.


### Community Topic Consumption

Hasura may automatically crawl pre-defined locations (such as Github) in order to collect third-party community contributions without authors needing to explicitly create new pull-requests, etc.

For example: Github Topics can be use to search for e.g. "#hasura-v3-packge" and if the repository contains valid package definitions, consume these and index them.

Please see previous work: TODO: Previous topic collection proof-of-concept

*Open-Question: How would a community crawled package transition to an explicitly managed package? How would ownership be established and transitioned, etc?*


### Indexing

The technical considerations were described in the "API" section, however, users should be able to leverage indexes on the following properties:

* Name
* Version
* Date of Publication
* Organisation
* Author
* Tags
* Category
* Free-text description
* Related packages
* Underlying DB (or Data solution)
* Any metadata in the Package description
* Verification status

### Discoverability

The discoverability component will simply hard-code various permutations of the "Indexing" criteria to provide the user browsable lists of packages.

These could include lists such as:

* Most popular packages
* Most recent packages
* Verified packages
* Etc.

### Access

Roles should be deliberately narrow in their scope, with higher level roles being able to grant lower-level roles, but not able to perform their duties implicitly.


### Publication

The publication of packages should be performed via the API by users who have the Author role.

The mechanism used is a three-step process:

* Request pre-authorized storage URI for a package version via API
* Upload the package definition archive via the storage URI
* Set the metadata for the package version

This can be done via a single user-interaction if an application (such as Hasura V3 CLI) abstracts these three steps.


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
* Etc.

