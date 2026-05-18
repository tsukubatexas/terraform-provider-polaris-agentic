package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestClientOAuthAndRequestHeaders(t *testing.T) {
	var tokenRequests int32
	var apiRequests int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			atomic.AddInt32(&tokenRequests, 1)
			if r.Method != http.MethodPost {
				t.Fatalf("token method got %s", r.Method)
			}
			if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/x-www-form-urlencoded") {
				t.Fatalf("token content-type got %q", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			if got := r.Form.Get("client_id"); got != "client-id" {
				t.Fatalf("client_id got %q", got)
			}
			if got := r.Form.Get("client_secret"); got != "client-secret" {
				t.Fatalf("client_secret got %q", got)
			}
			if got := r.Form.Get("scope"); got != "PRINCIPAL_ROLE:ALL" {
				t.Fatalf("scope got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "test-token"})
		case "/api/catalogs/risk":
			atomic.AddInt32(&apiRequests, 1)
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Fatalf("authorization got %q", got)
			}
			if got := r.Header.Get("Polaris-Realm"); got != "POLARIS" {
				t.Fatalf("realm got %q", got)
			}
			if got := r.Header.Get("User-Agent"); got != "terraform-provider-polaris-agentic/test" {
				t.Fatalf("user-agent got %q", got)
			}
			if got := r.URL.Query().Get("include"); got != "metadata" {
				t.Fatalf("query include got %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"name": "risk"})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := newClient(clientConfig{
		Endpoint:      server.URL + "/api",
		Realm:         "POLARIS",
		ClientID:      "client-id",
		ClientSecret:  "client-secret",
		OAuthTokenURL: server.URL + "/oauth/token",
		OAuthScope:    "PRINCIPAL_ROLE:ALL",
	})
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}
	client.UserAgent = "terraform-provider-polaris-agentic/test"

	for i := 0; i < 2; i++ {
		resp, err := client.do(context.Background(), "GET", "/catalogs/{catalogName}", map[string]string{"catalogName": "risk"}, map[string]string{"include": "metadata"}, nil, "")
		if err != nil {
			t.Fatalf("client.do: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status got %d", resp.StatusCode)
		}
	}

	if got := atomic.LoadInt32(&tokenRequests); got != 1 {
		t.Fatalf("token requests got %d", got)
	}
	if got := atomic.LoadInt32(&apiRequests); got != 2 {
		t.Fatalf("api requests got %d", got)
	}
}

func TestMapValuesIsDeterministic(t *testing.T) {
	got := strings.Join(mapValues(map[string]string{
		"z": "last",
		"a": "first",
		"m": "middle",
	}), ",")
	want := "a=first,m=middle,z=last"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
