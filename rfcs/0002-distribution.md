# Connector Package Distribution

## Purpose

Connector API, definition and packaging are specified respectively by:

* [NDC specification](http://hasura.github.io/ndc-spec/)
* [Deployment Specification](https://github.com/hasura/ndc-hub/blob/main/rfcs/0000-deployment.md)
* [Packaging Specification (WIP)](https://github.com/hasura/ndc-hub/pull/89/files)

This new distribution specification details how connector packages are intended to be owned, stored, indexed, searched, fetched and automatically published.

The intuition for this system is inspired by other package management systems such as NPM, Cabal, etc.

## Out of Scope for this RFC

The format of the metadata used for indexing, etc.


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
### Storage
### API
### Indexing
### Discoverability
### Access
### Publication
### Validation



