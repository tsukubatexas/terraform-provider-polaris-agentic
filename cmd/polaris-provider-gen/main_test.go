package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestWriteDocsIncludesGeneratedOperationMetadata(t *testing.T) {
	tmp := t.TempDir()
	filename := filepath.Join(tmp, "generated-operations.md")
	ops := map[string]generatedOperation{
		"createCatalog": {
			ID:      "createCatalog",
			Spec:    "spec/polaris-management-service.yml",
			Method:  "POST",
			Path:    "/catalogs",
			Summary: "Create catalog",
			Tags:    []string{"Management"},
		},
	}

	if err := writeDocs(filename, "apache-polaris-test", ops); err != nil {
		t.Fatalf("writeDocs: %v", err)
	}
	body, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read docs: %v", err)
	}
	text := string(body)
	for _, want := range []string{
		"Release: `apache-polaris-test`",
		"| Operation ID | Method | Path | Tags | Summary | Spec |",
		"`createCatalog`",
		"Management",
		"Create catalog",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("generated operations doc missing %q:\n%s", want, text)
		}
	}
}

func TestWriteTerraformProviderDocsAndExamples(t *testing.T) {
	tmp := t.TempDir()
	docsDir := filepath.Join(tmp, "docs")
	examplesDir := filepath.Join(tmp, "examples")
	ops := map[string]generatedOperation{
		"createCatalog": {
			ID:     "createCatalog",
			Spec:   "spec/polaris-management-service.yml",
			Method: "POST",
			Path:   "/catalogs",
		},
		"createNamespace": {
			ID:     "createNamespace",
			Spec:   "spec/iceberg-rest-catalog-open-api.yaml",
			Method: "POST",
			Path:   "/v1/{prefix}/namespaces",
		},
	}

	if err := writeTerraformProviderDocs(docsDir, "apache-polaris-test", ops); err != nil {
		t.Fatalf("writeTerraformProviderDocs: %v", err)
	}
	if err := writeTerraformExamples(examplesDir, "apache-polaris-test"); err != nil {
		t.Fatalf("writeTerraformExamples: %v", err)
	}

	assertFileContains(t, filepath.Join(docsDir, "index.md"), []string{
		"page_title: \"Polaris Provider\"",
		"Generated from Apache Polaris release `apache-polaris-test`",
		"resources/rest_resource.md",
		"data-sources/rest_call.md",
	})
	assertFileContains(t, filepath.Join(docsDir, "resources", "rest_resource.md"), []string{
		"# polaris_rest_resource",
		"`create_operation_id`",
	})
	assertFileContains(t, filepath.Join(docsDir, "data-sources", "rest_call.md"), []string{
		"# polaris_rest_call",
		"`operation_id`",
	})
	assertFileContains(t, filepath.Join(examplesDir, "complete-polaris", "main.tf"), []string{
		"provider \"polaris\"",
		"create_operation_id = \"createCatalog\"",
		"create_operation_id = \"createNamespace\"",
		"create_operation_id = \"createTable\"",
		"join(\"\\u001F\", var.namespace)",
	})
}

func TestDoHTTPRequestsRetriesTransientStatus(t *testing.T) {
	oldClient := httpClient
	oldRetryDelays := httpRetryDelays
	t.Cleanup(func() {
		httpClient = oldClient
		httpRetryDelays = oldRetryDelays
	})

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "try again", http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)

	httpClient = server.Client()
	httpRetryDelays = []time.Duration{0}

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := doHTTPRequest(req)
	if err != nil {
		t.Fatalf("doHTTPRequest: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != "ok" {
		t.Fatalf("body got %q", string(body))
	}
	if attempts != 2 {
		t.Fatalf("attempts got %d want 2", attempts)
	}
}

func assertFileContains(t *testing.T, filename string, wants []string) {
	t.Helper()
	body, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	text := string(body)
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("%s missing %q:\n%s", filename, want, text)
		}
	}
}
