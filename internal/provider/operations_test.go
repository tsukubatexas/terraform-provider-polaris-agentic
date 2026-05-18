package provider

import "testing"

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
