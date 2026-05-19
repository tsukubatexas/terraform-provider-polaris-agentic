# AI-Generated Terraform Provider Guide

This guide is the practical entry point for using `tsukubatexas/polaris` to manage Apache Polaris with Terraform.

The provider is generated from the Apache Polaris OpenAPI operation registry. It intentionally starts with generic building blocks:

- `polaris_rest_resource` for Terraform-managed create/read/update/delete lifecycles.
- `polaris_rest_call` for read-style calls and discovery.
- [generated-operations.md](generated-operations.md) as the list of supported `operationId` values.

## Mental Model

Every Terraform resource maps to one or more Polaris REST operations.

```text
OpenAPI operationId -> method + path template
path_params         -> values for {catalogName}, {principalName}, ...
body                -> JSON request body for create/update
id_attribute        -> dot-path in the JSON response used as Terraform ID
```

Use the operation IDs from [generated-operations.md](generated-operations.md). For endpoints that are not yet in the generated registry, use explicit `*_method` and `*_path` fields.

## Provider Configuration

Use a static bearer token for local smoke tests:

```hcl
provider "polaris" {
  endpoint = "http://localhost:8181/api/management/v1"
  realm    = "POLARIS"
  token    = var.polaris_token
}
```

Use OAuth client credentials for normal automation:

```hcl
provider "polaris" {
  endpoint        = "https://polaris.example.com/api/management/v1"
  realm           = "POLARIS"
  oauth_token_url = "https://login.microsoftonline.com/<tenant-id>/oauth2/v2.0/token"
  oauth_scope     = "api://<polaris-app-id>/.default"
  client_id       = var.client_id
  client_secret   = var.client_secret
}
```

The same values can be provided by environment variables:

```text
POLARIS_ENDPOINT
POLARIS_REALM
POLARIS_TOKEN
POLARIS_CLIENT_ID
POLARIS_CLIENT_SECRET
POLARIS_OAUTH_TOKEN_URL
POLARIS_OAUTH_SCOPE
```

## Create an Internal Catalog

This is the same lifecycle proven by `scripts/test_catalog.sh`: Terraform creates a catalog in a real Polaris container and then destroys it.

```hcl
resource "polaris_rest_resource" "risk_catalog" {
  create_operation_id = "createCatalog"
  read_operation_id   = "getCatalog"
  delete_operation_id = "deleteCatalog"

  path_params = {
    catalogName = "risk"
  }

  body = jsonencode({
    catalog = {
      type = "INTERNAL"
      name = "risk"
      properties = {
        "default-base-location" = "abfss://warehouse@riskstore.dfs.core.windows.net/iceberg/risk"
      }
      storageConfigInfo = {
        storageType      = "AZURE"
        allowedLocations = ["abfss://warehouse@riskstore.dfs.core.windows.net/iceberg/risk"]
        tenantId         = "<tenant-id>"
      }
    }
  })

  id_attribute = "name"
}
```

For a local S3-style test catalog:

```hcl
storageConfigInfo = {
  storageType      = "S3"
  allowedLocations = ["s3://agentic-test/"]
}
```

## Create Principals and Roles

Create a Polaris principal:

```hcl
resource "polaris_rest_resource" "analyst_principal" {
  create_operation_id = "createPrincipal"
  read_operation_id   = "getPrincipal"
  delete_operation_id = "deletePrincipal"

  path_params = {
    principalName = "analyst"
  }

  body = jsonencode({
    principal = {
      name = "analyst"
      properties = {
        owner = "data-platform"
      }
    }
  })

  id_attribute = "principal.name"
}
```

Create a principal role:

```hcl
resource "polaris_rest_resource" "risk_reader_role" {
  create_operation_id = "createPrincipalRole"
  read_operation_id   = "getPrincipalRole"
  delete_operation_id = "deletePrincipalRole"

  path_params = {
    principalRoleName = "risk-reader"
  }

  body = jsonencode({
    principalRole = {
      name = "risk-reader"
      properties = {
        purpose = "read-risk-catalog"
      }
    }
  })

  id_attribute = "name"
}
```

Assign the principal role to the principal:

```hcl
resource "polaris_rest_resource" "analyst_has_risk_reader" {
  create_operation_id = "assignPrincipalRole"
  read_operation_id   = "listPrincipalRolesAssigned"
  delete_operation_id = "revokePrincipalRole"

  path_params = {
    principalName     = polaris_rest_resource.analyst_principal.id
    principalRoleName = polaris_rest_resource.risk_reader_role.id
  }

  body = jsonencode({
    principalRole = {
      name = polaris_rest_resource.risk_reader_role.id
    }
  })

  depends_on = [
    polaris_rest_resource.analyst_principal,
    polaris_rest_resource.risk_reader_role,
  ]
}
```

## Create Catalog Roles and Attach Them

Create a catalog role inside a catalog:

```hcl
resource "polaris_rest_resource" "risk_catalog_reader" {
  create_operation_id = "createCatalogRole"
  read_operation_id   = "getCatalogRole"
  delete_operation_id = "deleteCatalogRole"

  path_params = {
    catalogName     = polaris_rest_resource.risk_catalog.id
    catalogRoleName = "catalog-reader"
  }

  body = jsonencode({
    catalogRole = {
      name = "catalog-reader"
      properties = {
        purpose = "read-risk-catalog"
      }
    }
  })

  id_attribute = "name"
}
```

Attach the catalog role to the principal role:

```hcl
resource "polaris_rest_resource" "risk_reader_gets_catalog_reader" {
  create_operation_id = "assignCatalogRoleToPrincipalRole"
  read_operation_id   = "listCatalogRolesForPrincipalRole"
  delete_operation_id = "revokeCatalogRoleFromPrincipalRole"

  path_params = {
    principalRoleName = polaris_rest_resource.risk_reader_role.id
    catalogName       = polaris_rest_resource.risk_catalog.id
    catalogRoleName   = polaris_rest_resource.risk_catalog_reader.id
  }

  body = jsonencode({
    catalogRole = {
      name = polaris_rest_resource.risk_catalog_reader.id
    }
  })

  depends_on = [
    polaris_rest_resource.risk_catalog_reader,
    polaris_rest_resource.risk_reader_role,
  ]
}
```

## Add a Catalog Grant

Polaris grants are represented by `addGrantToCatalogRole`.

```hcl
resource "polaris_rest_resource" "risk_catalog_read_properties" {
  create_operation_id = "addGrantToCatalogRole"
  read_operation_id   = "listGrantsForCatalogRole"

  path_params = {
    catalogName     = polaris_rest_resource.risk_catalog.id
    catalogRoleName = polaris_rest_resource.risk_catalog_reader.id
  }

  body = jsonencode({
    grant = {
      type      = "catalog"
      privilege = "CATALOG_READ_PROPERTIES"
    }
  })
}
```

Current limitation: Polaris grant revocation uses `revokeGrantFromCatalogRole`, which needs a request body. The generic resource currently sends a body only for create/update phases, so use grant resources for additive automation and handle precise revocation with a typed resource once it exists.

## Read Existing Polaris State

Use `polaris_rest_call` to inspect Polaris without managing lifecycle:

```hcl
data "polaris_rest_call" "catalogs" {
  operation_id = "listCatalogs"
}

output "catalogs_json" {
  value     = data.polaris_rest_call.catalogs.response_body
  sensitive = true
}
```

Read catalog roles for one catalog:

```hcl
data "polaris_rest_call" "risk_catalog_roles" {
  operation_id = "listCatalogRoles"

  path_params = {
    catalogName = "risk"
  }
}
```

## How to Add New Polaris Objects

1. Find the `operationId` in [generated-operations.md](generated-operations.md).
2. Check the Polaris OpenAPI schema under `specs/<release>/spec/...` for the request body shape.
3. Add a `polaris_rest_resource` with create/read/delete operation IDs.
4. Put all `{path}` placeholders in `path_params`.
5. Put the JSON request in `body`.
6. Set `id_attribute` to the name field returned by Polaris.
7. Run:

```bash
make generate fmt test build
POLARIS_INFRA_MODE=docker scripts/agentic_infra_loop.sh
```

If the OpenAPI registry changes in a future Polaris release, the autonomous update loop regenerates [generated-operations.md](generated-operations.md) and opens a PR.
