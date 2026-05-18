terraform {
  required_providers {
    polaris = {
      source  = "local/polaris/polaris"
      version = "0.0.0"
    }
  }
}

variable "endpoint" {
  type = string
}

variable "realm" {
  type = string
}

variable "token" {
  type      = string
  sensitive = true
}

provider "polaris" {
  endpoint = var.endpoint
  realm    = var.realm
  token    = var.token
}

resource "polaris_rest_resource" "test_catalog" {
  create_operation_id = "createCatalog"
  read_operation_id   = "getCatalog"
  delete_operation_id = "deleteCatalog"

  path_params = {
    catalogName = "agentic_test"
  }

  body = jsonencode({
    catalog = {
      type = "INTERNAL"
      name = "agentic_test"
      properties = {
        "default-base-location" = "file:///tmp/agentic_test"
      }
      storageConfigInfo = {
        storageType      = "FILE"
        allowedLocations = ["file:///tmp/agentic_test"]
      }
    }
  })

  id_attribute = "catalog.name"
}
