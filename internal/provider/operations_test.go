package provider

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

func TestSafeHTTPBodyTruncates(t *testing.T) {
	got := safeHTTPBody([]byte(strings.Repeat("a", 5000)))
	if !strings.Contains(got, "truncated") {
		t.Fatalf("expected truncation marker, got %q", got)
	}
	if len(got) > 4200 {
		t.Fatalf("truncated body too long: %d", len(got))
	}
}
