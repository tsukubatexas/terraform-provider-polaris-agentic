package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tsukubatexas/terraform-provider-polaris/internal/generated"
)

var testPathParamPattern = regexp.MustCompile(`\{([^{}]+)\}`)

func TestExpandPath(t *testing.T) {
	got, err := expandPath("/catalogs/{catalogName}/namespaces/{namespace}", map[string]string{
		"catalogName": "risk",
		"namespace":   "credit scores",
	})
	if err != nil {
		t.Fatalf("expandPath returned error: %v", err)
	}
	want := "/catalogs/risk/namespaces/credit%20scores"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExpandPathMissingParam(t *testing.T) {
	_, err := expandPath("/catalogs/{catalogName}", map[string]string{})
	if err == nil {
		t.Fatal("expected missing parameter error")
	}
}

func TestMethodAndPathRejectsUnsupportedMethod(t *testing.T) {
	resource := &schema.Resource{Schema: map[string]*schema.Schema{
		"operation_id": {Type: schema.TypeString, Optional: true},
		"method":       {Type: schema.TypeString, Optional: true},
		"path":         {Type: schema.TypeString, Optional: true},
	}}
	data := resource.Data(nil)
	if err := data.Set("method", "TRACE"); err != nil {
		t.Fatalf("set method: %v", err)
	}
	if err := data.Set("path", "/catalogs"); err != nil {
		t.Fatalf("set path: %v", err)
	}
	_, _, err := methodAndPath(data, "operation_id", "method", "path")
	if err == nil || !strings.Contains(err.Error(), "unsupported HTTP method") {
		t.Fatalf("expected unsupported method error, got %v", err)
	}
}

func TestAllGeneratedOpenAPIOperationsAreCallable(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if strings.ContainsAny(r.URL.EscapedPath(), "{}") {
			t.Fatalf("path still contains an unexpanded OpenAPI placeholder: %s", r.URL.EscapedPath())
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer server.Close()

	client, err := newClient(clientConfig{Endpoint: server.URL, Realm: "POLARIS", Token: "test-token"})
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}

	for id := range generated.Operations {
		op, err := operationByID(id)
		if err != nil {
			t.Fatalf("operationByID(%s): %v", id, err)
		}
		body := ""
		switch op.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			body = "{}"
		}
		resp, err := client.do(
			context.Background(),
			op.Method,
			op.Path,
			dummyPathParams(op.Path),
			nil,
			nil,
			body,
		)
		if err != nil {
			t.Fatalf("operation %s should be callable through generic client: %v", id, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("operation %s status got %d", id, resp.StatusCode)
		}
	}

	if requests != len(generated.Operations) {
		t.Fatalf("requests got %d want %d", requests, len(generated.Operations))
	}
}

func TestSafeHTTPBodyTruncates(t *testing.T) {
	got := safeHTTPBody([]byte(strings.Repeat("a", 5000)))
	if !strings.Contains(got, "truncated") {
		t.Fatalf("expected truncation marker, got %q", got)
	}
	if len(got) > 4200 {
		t.Fatalf("truncated body too long: %d", len(got))
	}
}

func dummyPathParams(pathTemplate string) map[string]string {
	params := map[string]string{}
	for _, match := range testPathParamPattern.FindAllStringSubmatch(pathTemplate, -1) {
		params[match[1]] = "test-value"
	}
	return params
}
