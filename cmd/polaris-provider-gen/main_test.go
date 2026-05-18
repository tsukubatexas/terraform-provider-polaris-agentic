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

func TestWriteOperationsIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "operations_gen.go")

	ops := map[string]generatedOperation{
		"b": {ID: "b", Spec: "spec/b.yaml", Method: "POST", Path: "/b", Summary: "b", Tags: []string{"B"}},
		"a": {ID: "a", Spec: "spec/a.yaml", Method: "GET", Path: "/a", Summary: "a", Tags: []string{"A"}},
	}

	if err := writeOperations(out, "apache-polaris-0.0.0", ops); err != nil {
		t.Fatalf("writeOperations: %v", err)
	}

	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	text := string(body)

	if !strings.Contains(text, "var GeneratedAt = \"reproducible\"") {
		t.Fatalf("expected reproducible GeneratedAt, got:\n%s", text)
	}

	idxA := strings.Index(text, "\"a\": {")
	idxB := strings.Index(text, "\"b\": {")
	if idxA == -1 || idxB == -1 {
		t.Fatalf("expected both operations, got:\n%s", text)
	}
	if idxA > idxB {
		t.Fatalf("expected sorted output (a before b), got:\n%s", text)
	}
}

func TestWriteDocsIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "generated-operations.md")

	ops := map[string]generatedOperation{
		"b": {ID: "b", Spec: "spec/b.yaml", Method: "POST", Path: "/b"},
		"a": {ID: "a", Spec: "spec/a.yaml", Method: "GET", Path: "/a"},
	}

	if err := writeDocs(out, "apache-polaris-0.0.0", ops); err != nil {
		t.Fatalf("writeDocs: %v", err)
	}

	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	text := string(body)

	idxA := strings.Index(text, "| `a` |")
	idxB := strings.Index(text, "| `b` |")
	if idxA == -1 || idxB == -1 {
		t.Fatalf("expected both operations, got:\n%s", text)
	}
	if idxA > idxB {
		t.Fatalf("expected sorted output (a before b), got:\n%s", text)
	}
}
