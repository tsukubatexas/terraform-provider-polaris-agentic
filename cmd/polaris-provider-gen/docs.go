package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
)

type operationCounts struct {
	Management int
	Catalog    int
	Policy     int
	Generic    int
	Other      int
}

func writeTerraformProviderDocs(root, tag string, ops map[string]generatedOperation) error {
	if err := writeFile(path.Join(root, "index.md"), terraformIndexDoc(tag, ops)); err != nil {
		return err
	}
	if err := writeFile(path.Join(root, "resources", "rest_resource.md"), terraformRestResourceDoc(tag)); err != nil {
		return err
	}
	if err := writeFile(path.Join(root, "data-sources", "rest_call.md"), terraformRestCallDoc(tag)); err != nil {
		return err
	}
	if err := writeFile(path.Join(root, "guides", "complete-polaris-configuration.md"), completePolarisGuide(tag)); err != nil {
		return err
	}
	return nil
}

func writeTerraformExamples(root, tag string) error {
	files := map[string]string{
		path.Join(root, "provider", "provider.tf"):                             providerExample(tag),
		path.Join(root, "resources", "polaris_rest_resource", "resource.tf"):   resourceExample(tag),
		path.Join(root, "data-sources", "polaris_rest_call", "data-source.tf"): dataSourceExample(tag),
		path.Join(root, "complete-polaris", "main.tf"):                         completePolarisExample(tag),
		path.Join(root, "complete-polaris", "README.md"):                       completePolarisExampleReadme(tag),
	}
	for filename, body := range files {
		if err := writeFile(filename, body); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(filename string, body string) error {
	if err := os.MkdirAll(path.Dir(filename), 0o755); err != nil {
		return err
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return os.WriteFile(filename, []byte(body), 0o644)
}

func terraformIndexDoc(tag string, ops map[string]generatedOperation) string {
	counts := countOperations(ops)
	var buf bytes.Buffer
	buf.WriteString(terraformGeneratedHeader(tag, "Polaris Provider", "Manage Apache Polaris through generated OpenAPI operation IDs."))
	buf.WriteString("# Polaris Provider\n\n")
	buf.WriteString("The Polaris provider exposes Apache Polaris management and catalog APIs through generated OpenAPI operation IDs. It is intentionally small: the generated operation registry supplies method/path metadata, while Terraform configuration supplies lifecycle semantics, path parameters, JSON bodies, and IDs.\n\n")
	fmt.Fprintf(&buf, "Generated from Apache Polaris release `%s` with `%d` operations: `%d` management, `%d` catalog, `%d` policy, `%d` generic-table, `%d` other.\n\n", tag, len(ops), counts.Management, counts.Catalog, counts.Policy, counts.Generic, counts.Other)
	buf.WriteString("## Example Usage\n\n")
	buf.WriteString("Use one provider alias for management endpoints and one alias for Iceberg REST catalog endpoints. Polaris realms are configured through the provider `realm` argument, which sends the `Polaris-Realm` header; this provider does not create realms as Terraform resources.\n\n")
	buf.WriteString("```terraform\n")
	buf.WriteString(strings.TrimSpace(providerExampleBody()))
	buf.WriteString("\n```\n\n")
	buf.WriteString("## Documentation\n\n")
	buf.WriteString("- [polaris_rest_resource](resources/rest_resource.md) manages create/read/update/delete lifecycles.\n")
	buf.WriteString("- [polaris_rest_call](data-sources/rest_call.md) reads from Polaris without owning lifecycle.\n")
	buf.WriteString("- [Generated operation inventory](generated-operations.md) lists every generated `operation_id`.\n")
	buf.WriteString("- [Complete Polaris configuration guide](guides/complete-polaris-configuration.md) shows realm header, catalog, principal, roles, namespace, table, and grants together.\n\n")
	buf.WriteString("## Schema\n\n")
	buf.WriteString("### Optional\n\n")
	buf.WriteString("- `endpoint` (String) Base Polaris API endpoint. Can also be set with `POLARIS_ENDPOINT`.\n")
	buf.WriteString("- `realm` (String) Polaris realm sent as the `Polaris-Realm` header. Can also be set with `POLARIS_REALM`.\n")
	buf.WriteString("- `token` (String, Sensitive) Static bearer token. Can also be set with `POLARIS_TOKEN`.\n")
	buf.WriteString("- `client_id` (String) OAuth client ID. Can also be set with `POLARIS_CLIENT_ID`.\n")
	buf.WriteString("- `client_secret` (String, Sensitive) OAuth client secret. Can also be set with `POLARIS_CLIENT_SECRET`.\n")
	buf.WriteString("- `oauth_token_url` (String) OAuth token endpoint. Can also be set with `POLARIS_OAUTH_TOKEN_URL`.\n")
	buf.WriteString("- `oauth_scope` (String) OAuth scope for client credentials. Can also be set with `POLARIS_OAUTH_SCOPE`.\n")
	buf.WriteString("- `insecure_skip_tls_verify` (Boolean) Defaults to `false`. Use only for isolated tests.\n\n")
	buf.WriteString("## Authentication\n\n")
	buf.WriteString("For local smoke tests, a short-lived bearer token is enough. For automation, prefer OAuth client credentials from a CI secret store or workload identity integration. Do not commit token values, generated Terraform state, or client secrets.\n")
	return buf.String()
}

func terraformRestResourceDoc(tag string) string {
	var buf bytes.Buffer
	buf.WriteString(terraformGeneratedHeader(tag, "polaris_rest_resource Resource", "Manage a Polaris REST object with generated OpenAPI operation IDs."))
	buf.WriteString("# polaris_rest_resource\n\n")
	buf.WriteString("`polaris_rest_resource` maps Terraform lifecycle phases to Polaris REST operations. Use generated `*_operation_id` values when possible. Use explicit `*_method` and `*_path` only for an endpoint that is not present in the generated operation inventory.\n\n")
	buf.WriteString("## Example Usage\n\n")
	buf.WriteString("```terraform\n")
	buf.WriteString(strings.TrimSpace(resourceExampleBody()))
	buf.WriteString("\n```\n\n")
	buf.WriteString("## Schema\n\n")
	buf.WriteString("### Optional\n\n")
	buf.WriteString("- `create_operation_id` (String) OpenAPI operation ID used for create.\n")
	buf.WriteString("- `read_operation_id` (String) OpenAPI operation ID used for read.\n")
	buf.WriteString("- `update_operation_id` (String) OpenAPI operation ID used for update.\n")
	buf.WriteString("- `delete_operation_id` (String) OpenAPI operation ID used for delete.\n")
	buf.WriteString("- `create_method`, `read_method`, `update_method`, `delete_method` (String) HTTP method fallback when an operation ID is not set.\n")
	buf.WriteString("- `create_path`, `read_path`, `update_path`, `delete_path` (String) HTTP path fallback when an operation ID is not set.\n")
	buf.WriteString("- `path_params` (Map of String) Values for OpenAPI path placeholders such as `{catalogName}` or `{namespace}`.\n")
	buf.WriteString("- `query_params` (Map of String) Query string parameters.\n")
	buf.WriteString("- `headers` (Map of String) Additional HTTP headers for this resource.\n")
	buf.WriteString("- `body` (String, Sensitive) JSON request body used for create and update.\n")
	buf.WriteString("- `id_attribute` (String) Dot-path in the create response used as Terraform ID, for example `name` or `principal.name`.\n")
	buf.WriteString("- `expected_status_codes` (List of Number) Accepted status codes for create/update/delete. Defaults to `200`, `201`, and `204`.\n\n")
	buf.WriteString("### Read-Only\n\n")
	buf.WriteString("- `status_code` (Number) Last HTTP status code.\n")
	buf.WriteString("- `response_body` (String, Sensitive) Last JSON response body.\n\n")
	buf.WriteString("## Lifecycle Notes\n\n")
	buf.WriteString("The generic resource cannot infer Terraform semantics from OpenAPI alone. Always choose explicit read and delete operations where Polaris supports them. Some Polaris operations, such as grant revocation, require a request body on a non-create phase; those should be handled by a typed resource once the provider grows beyond the generic bootstrap resource.\n")
	return buf.String()
}

func terraformRestCallDoc(tag string) string {
	var buf bytes.Buffer
	buf.WriteString(terraformGeneratedHeader(tag, "polaris_rest_call Data Source", "Read from Apache Polaris with generated OpenAPI operation IDs."))
	buf.WriteString("# polaris_rest_call\n\n")
	buf.WriteString("`polaris_rest_call` runs a Polaris REST call during Terraform refresh. It is useful for discovery, validation, and output values when Terraform should not own the lifecycle of the remote object.\n\n")
	buf.WriteString("## Example Usage\n\n")
	buf.WriteString("```terraform\n")
	buf.WriteString(strings.TrimSpace(dataSourceExampleBody()))
	buf.WriteString("\n```\n\n")
	buf.WriteString("## Schema\n\n")
	buf.WriteString("### Optional\n\n")
	buf.WriteString("- `operation_id` (String) OpenAPI operation ID from the generated Polaris operation registry.\n")
	buf.WriteString("- `method` (String) HTTP method fallback when `operation_id` is not set.\n")
	buf.WriteString("- `path` (String) HTTP path fallback when `operation_id` is not set.\n")
	buf.WriteString("- `path_params` (Map of String) Values for OpenAPI path placeholders.\n")
	buf.WriteString("- `query_params` (Map of String) Query string parameters.\n")
	buf.WriteString("- `headers` (Map of String) Additional HTTP headers.\n")
	buf.WriteString("- `body` (String, Sensitive) Optional JSON request body for advanced read-style operations.\n")
	buf.WriteString("- `expected_status_codes` (List of Number) Accepted status codes. Defaults to `200`.\n\n")
	buf.WriteString("### Read-Only\n\n")
	buf.WriteString("- `status_code` (Number) HTTP status code.\n")
	buf.WriteString("- `response_body` (String, Sensitive) JSON response body.\n")
	return buf.String()
}

func completePolarisGuide(tag string) string {
	var buf bytes.Buffer
	buf.WriteString(terraformGeneratedHeader(tag, "Complete Polaris Configuration Guide", "Configure a Polaris realm header, catalog, namespace, table, principals, roles, and grants."))
	buf.WriteString("# Complete Polaris Configuration\n\n")
	buf.WriteString("This guide shows how the generic Terraform provider pieces fit together for one end-to-end Polaris setup. The realm is provider configuration, not a resource: `realm` sends the `Polaris-Realm` header on every request.\n\n")
	buf.WriteString("The example uses two provider aliases because Polaris management APIs and Iceberg REST catalog APIs usually have different base paths:\n\n")
	buf.WriteString("- `polaris.management` targets `/api/management/v1` for catalogs, principals, roles, and grants.\n")
	buf.WriteString("- `polaris.catalog` targets `/api/catalog` for Iceberg REST operations such as namespace and table creation.\n\n")
	buf.WriteString("Terraform can create control-plane objects and initial metadata objects. Data writes and ACID commits should still go through an Iceberg engine such as Trino, Spark, or PyIceberg.\n\n")
	buf.WriteString("Grant resources in this generic example are additive because Polaris grant revocation needs a request body on revoke. Deleting the catalog role removes its grants, but strict per-grant destroy semantics should be implemented as a typed provider resource.\n\n")
	buf.WriteString("## Full Example\n\n")
	buf.WriteString("```terraform\n")
	buf.WriteString(strings.TrimSpace(completePolarisExampleBody()))
	buf.WriteString("\n```\n")
	return buf.String()
}

func providerExample(tag string) string {
	return generatedTerraformComment(tag) + providerExampleBody()
}

func resourceExample(tag string) string {
	return generatedTerraformComment(tag) + resourceExampleBody()
}

func dataSourceExample(tag string) string {
	return generatedTerraformComment(tag) + dataSourceExampleBody()
}

func completePolarisExample(tag string) string {
	return generatedTerraformComment(tag) + completePolarisExampleBody()
}

func completePolarisExampleReadme(tag string) string {
	return fmt.Sprintf(`# Complete Polaris Example

This example is generated by `+"`cmd/polaris-provider-gen`"+` from Apache Polaris release `+"`%s`"+`.

It demonstrates one complete Terraform shape:

- Realm configuration through the provider `+"`realm`"+` argument.
- Management API alias for catalog, principal, roles, and grants.
- Catalog API alias for Iceberg namespace and table metadata.
- A catalog-level grant and a table-level grant.

The example is a template. Set real endpoint, OAuth, and storage values before running it.
`, tag)
}

func providerExampleBody() string {
	return `terraform {
  required_providers {
    polaris = {
      source = "tsukubatexas/polaris"
    }
  }
}

variable "polaris_base_url" {
  type = string
}

variable "polaris_realm" {
  type    = string
  default = "POLARIS"
}

variable "oauth_token_url" {
  type = string
}

variable "oauth_scope" {
  type = string
}

variable "client_id" {
  type = string
}

variable "client_secret" {
  type      = string
  sensitive = true
}

provider "polaris" {
  alias    = "management"
  endpoint = "${var.polaris_base_url}/api/management/v1"
  realm    = var.polaris_realm

  oauth_token_url = var.oauth_token_url
  oauth_scope     = var.oauth_scope
  client_id       = var.client_id
  client_secret   = var.client_secret
}

provider "polaris" {
  alias    = "catalog"
  endpoint = "${var.polaris_base_url}/api/catalog"
  realm    = var.polaris_realm

  oauth_token_url = var.oauth_token_url
  oauth_scope     = var.oauth_scope
  client_id       = var.client_id
  client_secret   = var.client_secret
}`
}

func resourceExampleBody() string {
	return `resource "polaris_rest_resource" "catalog" {
  provider = polaris.management

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
        "default-base-location" = "s3://risk-warehouse/"
      }
      storageConfigInfo = {
        storageType      = "S3"
        allowedLocations = ["s3://risk-warehouse/"]
      }
    }
  })

  id_attribute = "name"
}`
}

func dataSourceExampleBody() string {
	return `data "polaris_rest_call" "catalogs" {
  provider = polaris.management

  operation_id = "listCatalogs"
}

output "catalogs_json" {
  value     = data.polaris_rest_call.catalogs.response_body
  sensitive = true
}`
}

func completePolarisExampleBody() string {
	return `terraform {
  required_providers {
    polaris = {
      source = "tsukubatexas/polaris"
    }
  }
}

variable "polaris_base_url" {
  description = "Polaris base URL without a trailing slash, for example https://polaris.example.com"
  type        = string
}

variable "polaris_realm" {
  description = "Polaris realm. This becomes the Polaris-Realm request header."
  type        = string
  default     = "POLARIS"
}

variable "oauth_token_url" {
  type = string
}

variable "oauth_scope" {
  type = string
}

variable "client_id" {
  type = string
}

variable "client_secret" {
  type      = string
  sensitive = true
}

variable "catalog_name" {
  type    = string
  default = "risk"
}

variable "warehouse_location" {
  description = "Base location allowed for the Polaris catalog."
  type        = string
  default     = "s3://risk-warehouse"
}

variable "namespace" {
  description = "Iceberg namespace parts."
  type        = list(string)
  default     = ["risk", "curated"]
}

variable "table_name" {
  type    = string
  default = "credit_events"
}

locals {
  namespace_path     = join("\u001F", var.namespace)
  namespace_location = "${var.warehouse_location}/${join("/", var.namespace)}"
  table_location     = "${local.namespace_location}/${var.table_name}"
}

provider "polaris" {
  alias    = "management"
  endpoint = "${var.polaris_base_url}/api/management/v1"
  realm    = var.polaris_realm

  oauth_token_url = var.oauth_token_url
  oauth_scope     = var.oauth_scope
  client_id       = var.client_id
  client_secret   = var.client_secret
}

provider "polaris" {
  alias    = "catalog"
  endpoint = "${var.polaris_base_url}/api/catalog"
  realm    = var.polaris_realm

  oauth_token_url = var.oauth_token_url
  oauth_scope     = var.oauth_scope
  client_id       = var.client_id
  client_secret   = var.client_secret
}

resource "polaris_rest_resource" "catalog" {
  provider = polaris.management

  create_operation_id = "createCatalog"
  read_operation_id   = "getCatalog"
  delete_operation_id = "deleteCatalog"

  path_params = {
    catalogName = var.catalog_name
  }

  body = jsonencode({
    catalog = {
      type = "INTERNAL"
      name = var.catalog_name
      properties = {
        "default-base-location" = var.warehouse_location
      }
      storageConfigInfo = {
        storageType      = "S3"
        allowedLocations = [var.warehouse_location]
      }
    }
  })

  id_attribute = "name"
}

resource "polaris_rest_resource" "principal" {
  provider = polaris.management

  create_operation_id = "createPrincipal"
  read_operation_id   = "getPrincipal"
  delete_operation_id = "deletePrincipal"

  path_params = {
    principalName = "risk-analyst"
  }

  body = jsonencode({
    principal = {
      name = "risk-analyst"
      properties = {
        owner = "risk-team"
      }
    }
  })

  id_attribute = "principal.name"
}

resource "polaris_rest_resource" "principal_role" {
  provider = polaris.management

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
        purpose = "risk-read-access"
      }
    }
  })

  id_attribute = "name"
}

resource "polaris_rest_resource" "assign_principal_role" {
  provider = polaris.management

  create_operation_id = "assignPrincipalRole"
  read_operation_id   = "listPrincipalRolesAssigned"
  delete_operation_id = "revokePrincipalRole"

  path_params = {
    principalName     = polaris_rest_resource.principal.id
    principalRoleName = polaris_rest_resource.principal_role.id
  }

  body = jsonencode({
    principalRole = {
      name = polaris_rest_resource.principal_role.id
    }
  })
}

resource "polaris_rest_resource" "catalog_role" {
  provider = polaris.management

  create_operation_id = "createCatalogRole"
  read_operation_id   = "getCatalogRole"
  delete_operation_id = "deleteCatalogRole"

  path_params = {
    catalogName     = polaris_rest_resource.catalog.id
    catalogRoleName = "risk-table-reader"
  }

  body = jsonencode({
    catalogRole = {
      name = "risk-table-reader"
      properties = {
        purpose = "read-risk-table"
      }
    }
  })

  id_attribute = "name"
}

resource "polaris_rest_resource" "assign_catalog_role" {
  provider = polaris.management

  create_operation_id = "assignCatalogRoleToPrincipalRole"
  read_operation_id   = "listCatalogRolesForPrincipalRole"
  delete_operation_id = "revokeCatalogRoleFromPrincipalRole"

  path_params = {
    principalRoleName = polaris_rest_resource.principal_role.id
    catalogName       = polaris_rest_resource.catalog.id
    catalogRoleName   = polaris_rest_resource.catalog_role.id
  }

  body = jsonencode({
    catalogRole = {
      name = polaris_rest_resource.catalog_role.id
    }
  })
}

resource "polaris_rest_resource" "namespace" {
  provider = polaris.catalog

  create_operation_id = "createNamespace"
  read_operation_id   = "loadNamespaceMetadata"
  delete_operation_id = "dropNamespace"

  path_params = {
    prefix    = polaris_rest_resource.catalog.id
    namespace = local.namespace_path
  }

  body = jsonencode({
    namespace = var.namespace
    properties = {
      owner    = "risk-team"
      location = local.namespace_location
    }
  })

  depends_on = [
    polaris_rest_resource.catalog,
  ]
}

resource "polaris_rest_resource" "table" {
  provider = polaris.catalog

  create_operation_id = "createTable"
  read_operation_id   = "loadTable"
  delete_operation_id = "dropTable"

  path_params = {
    prefix    = polaris_rest_resource.catalog.id
    namespace = local.namespace_path
    table     = var.table_name
  }

  body = jsonencode({
    name     = var.table_name
    location = local.table_location
    schema = {
      type = "struct"
      fields = [
        {
          id       = 1
          name     = "event_id"
          type     = "string"
          required = true
        },
        {
          id       = 2
          name     = "risk_score"
          type     = "double"
          required = false
        },
        {
          id       = 3
          name     = "event_time"
          type     = "timestamptz"
          required = false
        }
      ]
      "identifier-field-ids" = [1]
    }
    properties = {
      owner                  = "risk-team"
      "write.format.default" = "parquet"
    }
  })

  expected_status_codes = [200, 201, 204]

  depends_on = [
    polaris_rest_resource.namespace,
  ]
}

resource "polaris_rest_resource" "catalog_read_grant" {
  provider = polaris.management

  create_operation_id = "addGrantToCatalogRole"
  read_operation_id   = "listGrantsForCatalogRole"

  path_params = {
    catalogName     = polaris_rest_resource.catalog.id
    catalogRoleName = polaris_rest_resource.catalog_role.id
  }

  body = jsonencode({
    grant = {
      type      = "catalog"
      privilege = "CATALOG_MANAGE_METADATA"
    }
  })

  depends_on = [
    polaris_rest_resource.assign_catalog_role,
  ]
}

resource "polaris_rest_resource" "table_read_grant" {
  provider = polaris.management

  create_operation_id = "addGrantToCatalogRole"
  read_operation_id   = "listGrantsForCatalogRole"

  path_params = {
    catalogName     = polaris_rest_resource.catalog.id
    catalogRoleName = polaris_rest_resource.catalog_role.id
  }

  body = jsonencode({
    grant = {
      type      = "table"
      namespace = var.namespace
      tableName = var.table_name
      privilege = "TABLE_READ_DATA"
    }
  })

  depends_on = [
    polaris_rest_resource.table,
    polaris_rest_resource.assign_catalog_role,
  ]
}

data "polaris_rest_call" "tables" {
  provider = polaris.catalog

  operation_id = "listTables"

  path_params = {
    prefix    = polaris_rest_resource.catalog.id
    namespace = local.namespace_path
  }

  depends_on = [
    polaris_rest_resource.table,
  ]
}

output "tables_json" {
  value     = data.polaris_rest_call.tables.response_body
  sensitive = true
}`
}

func terraformGeneratedHeader(tag, pageTitle, description string) string {
	return fmt.Sprintf(`---
page_title: "%s"
description: |-
  %s
---

<!-- Code generated by cmd/polaris-provider-gen from %s. DO NOT EDIT. -->

`, pageTitle, description, tag)
}

func generatedTerraformComment(tag string) string {
	return fmt.Sprintf("# Code generated by cmd/polaris-provider-gen from %s. DO NOT EDIT.\n\n", tag)
}

func markdownCell(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\n", " ")
	return value
}

func countOperations(ops map[string]generatedOperation) operationCounts {
	var counts operationCounts
	for _, op := range ops {
		switch {
		case strings.Contains(op.Spec, "polaris-management-service"):
			counts.Management++
		case strings.Contains(op.Spec, "iceberg-rest-catalog"):
			counts.Catalog++
		case strings.Contains(op.Spec, "policy-apis"):
			counts.Policy++
		case strings.Contains(op.Spec, "generic-tables-api"):
			counts.Generic++
		default:
			counts.Other++
		}
	}
	return counts
}
