# Connector Package Distribution

## Purpose

Connector API, definition and packaging are specified respectively by:

* [NDC specification](http://hasura.github.io/ndc-spec/)
* [Deployment Specification](https://github.com/hasura/ndc-hub/blob/main/rfcs/0000-deployment.md)
* [Packaging Specification (WIP)](https://github.com/hasura/ndc-hub/pull/89/files)

This new distribution specification details how connector packages are intended to be owned, stored, indexed, searched, fetched and automatically published.

The intuition for this system is inspired by other package management systems such as NPM, Cabal, etc.

There was a previous implementation of these concepts as described (TODO: Get docs links from Shraddha)


## Out of Scope for this RFC

The format of the metadata used for indexing, etc.

Authentication mechanisms.


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

### Database

The database backing the API provides all of the APIs state management capabilities outside of package archive storage (as described in "Storage").

The initial implementation of the Database will be Postgres.

### API

The API provides the user-interaction layer that mediates the database and storage components. No direct user interaction should occur with either the database, or storage, except for the case when the API delegates a storage interaction - such as providing an author a pre-authorized URL for publication, or providing a Hasura V3 project user a public storage URL for a package definition.

The various functions of the API are described as follows:

#### Administration

* TODO

#### Authoring

* TODO

#### Discoverability

* TODO

#### Acquisition

* TODO

### Indexing
### Discoverability
### Access
### Publication
### Validation (TODO: What did I mean by this?)


