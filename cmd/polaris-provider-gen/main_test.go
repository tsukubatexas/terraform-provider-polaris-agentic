package main

import (
	"strings"
	"testing"
)

func TestParseSpecGeneratesStableFallbackIDs(t *testing.T) {
	body := []byte(`
openapi: 3.0.0
info:
  title: Test
  version: 1.0.0
paths:
  /catalogs/{catalogName}:
    parameters:
      - name: catalogName
        in: path
    get:
      summary: Load catalog
      tags:
        - Catalogs
      responses:
        "200":
          description: ok
    trace:
      responses:
        "200":
          description: ignored
`)

	ops, err := parseSpec("spec/test.yaml", body)
	if err != nil {
		t.Fatalf("parseSpec: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops length got %d", len(ops))
	}
	op := ops[0]
	if op.ID != "spec_test_yaml_GET_catalogs_catalogName" {
		t.Fatalf("id got %q", op.ID)
	}
	if op.Method != "GET" {
		t.Fatalf("method got %q", op.Method)
	}
	if op.Path != "/catalogs/{catalogName}" {
		t.Fatalf("path got %q", op.Path)
	}
	if op.Summary != "Load catalog" {
		t.Fatalf("summary got %q", op.Summary)
	}
	if strings.Join(op.Tags, ",") != "Catalogs" {
		t.Fatalf("tags got %#v", op.Tags)
	}
}

func TestStableOperationIDNormalizesSeparators(t *testing.T) {
	got := stableOperationID("spec/polaris-catalog-apis/policy-apis.yaml", "POST", "/v1/{prefix}/tables/{table}")
	want := "spec_polaris_catalog_apis_policy_apis_yaml_POST_v1_prefix_tables_table"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
