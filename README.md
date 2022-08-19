# KCP API Catalog

A library for kcp API catalog code.

The API Catalog provides publicly available interfaces (APIs, code, CLIs) that allows cluster admin/API providers to curate a collection of available APIs (`APIExport`) in the KCP environment.

## What is API Catalog?

Each available API is encapsulated in a new API called `CatalogEntry` that is created by the cluster admin/API provider in a specific workspace.

The `CatalogEntry` contains a list of `ExportReference` and each `ExportReference` links to a corresponding `APIExport`. This `1:N` pattern between `CatalogEntry` and `APIExport` facilitates variety of relationships between multiple `APIExport` such as dependencies or grouping.

In addition of `APIExport` identifier information, the `CatalogEntry` have additional fields such as `Icons` and `Description` to allow cluster admin/API provider to add more information to represent the API specification and information for UI/Console usage.

All of `CatalogEntry` in the same workspace is considered to be in the same group (catalog). The availability (available for binding) of a CatalogEntry depends on the location of the workspace and permissions to access the workspace. For example, if the catalog workspace is at `root` level, then the `CatalogEntry` in that catalog is usable for all tenants' workspaces. If the workspace belongs to a specific organization such as `root:redhat`, then it will only be accessible and usable to tenants under that `redhat` organization.

When API consumer wants to bind an available API from a `CatalogEntry`, an `APIBinding` is created with each `ExportReference` in `Exports`.

## Current Goals

- Initial Catalog API spec to support optional information such as `Description` in the spec
- Link `CatalogEntry` API to singular or multiple `APIExport` using a list of `ExportReference` (name of the APIExport and workspace path)
- CLI command `bind` to create `APIBinding` from `ExportReference` in `CatalogEntry` assuming all necessary RBAC are granted

## Future Goals

- Replace `ExportReference` reference in `Exports` with a `IdentityHash` to avoid exposing `APIExport` name and location
- Create a RBAC generation/control mechanism to facilitate the full `APIBinding` usage such as permission request and maximum permission policy
- Catalog controller to perform validation and additional reconciliation
- More CLI commands

For contributions, issues, or general discussion, please see the [kcp repository](https://github.com/kcp-dev/kcp).
