package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	httpClient      = &http.Client{Timeout: 30 * time.Second}
	httpRetryDelays = []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
)

var defaultSpecs = []specSource{
	{Path: "spec/polaris-management-service.yml", Required: true},
	{Path: "spec/polaris-catalog-service.yaml"},
	{Path: "spec/iceberg-rest-catalog-open-api.yaml"},
	{Path: "spec/polaris-catalog-apis/generic-tables-api.yaml"},
	{Path: "spec/polaris-catalog-apis/notifications-api.yaml"},
	{Path: "spec/polaris-catalog-apis/oauth-tokens-api.yaml"},
	{Path: "spec/polaris-catalog-apis/policy-apis.yaml"},
}

type specSource struct {
	Path     string
	Required bool
}

type releaseResponse struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

type openAPISpec struct {
	Info struct {
		Title   string `yaml:"title"`
		Version string `yaml:"version"`
	} `yaml:"info"`
	Paths map[string]map[string]yaml.Node `yaml:"paths"`
}

type operation struct {
	OperationID string   `yaml:"operationId"`
	Summary     string   `yaml:"summary"`
	Tags        []string `yaml:"tags"`
}

type generatedOperation struct {
	ID      string
	Spec    string
	Method  string
	Path    string
	Summary string
	Tags    []string
}

func main() {
	release := flag.String("release", "latest", "GitHub release tag, or latest")
	out := flag.String("out", "internal/generated/operations_gen.go", "Generated Go file path")
	docsOut := flag.String("docs-out", "docs/generated-operations.md", "Generated Markdown operation inventory")
	terraformDocsDir := flag.String("terraform-docs-dir", "docs", "Generated Terraform Registry-style documentation directory")
	examplesDir := flag.String("examples-dir", "examples", "Generated Terraform example directory")
	specCacheDir := flag.String("spec-cache-dir", "specs", "Directory where fetched specs are cached")
	flag.Parse()

	tag := *release
	if tag == "latest" {
		latest, err := latestRelease()
		must(err)
		tag = latest.TagName
	}
	if tag == "" {
		die("empty release tag")
	}

	ops := map[string]generatedOperation{}
	skipped := []string{}
	for _, source := range defaultSpecs {
		body, ok, err := fetchSpec(tag, source)
		must(err)
		if !ok {
			skipped = append(skipped, source.Path)
			continue
		}
		cacheSpec(*specCacheDir, tag, source.Path, body)

		specOps, err := parseSpec(source.Path, body)
		must(err)
		for _, op := range specOps {
			if existing, ok := ops[op.ID]; ok {
				op.ID = stableOperationID(op.Spec, op.Method, op.Path)
				if _, ok := ops[op.ID]; ok {
					die("duplicate operation id %q from %s and %s", op.ID, existing.Spec, op.Spec)
				}
			}
			ops[op.ID] = op
		}
	}

	must(writeOperations(*out, tag, ops))
	must(writeDocs(*docsOut, tag, ops))
	must(writeTerraformProviderDocs(*terraformDocsDir, tag, ops))
	must(writeTerraformExamples(*examplesDir, tag))
	fmt.Printf("Generated %d Polaris operations from %s\n", len(ops), tag)
	if len(skipped) > 0 {
		fmt.Printf("Skipped optional specs missing in %s: %s\n", tag, strings.Join(skipped, ", "))
	}
}

func latestRelease() (*releaseResponse, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/apache/polaris/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "terraform-provider-polaris-agentic-generator")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := doHTTPRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub release lookup failed with HTTP %d: %s", resp.StatusCode, string(body))
	}
	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return &release, nil
}

func fetchSpec(tag string, source specSource) ([]byte, bool, error) {
	specURL := fmt.Sprintf("https://raw.githubusercontent.com/apache/polaris/%s/%s", tag, source.Path)
	req, err := http.NewRequest(http.MethodGet, specURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", "terraform-provider-polaris-agentic-generator")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := doHTTPRequest(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound && !source.Required {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("fetch %s failed with HTTP %d: %s", specURL, resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	return body, true, err
}

func doHTTPRequest(req *http.Request) (*http.Response, error) {
	attempts := len(httpRetryDelays) + 1
	var lastErr error
	var lastStatus int
	var lastBody []byte

	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := httpClient.Do(req.Clone(req.Context()))
		if err == nil && !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		if resp != nil {
			lastStatus = resp.StatusCode
			lastBody, _ = io.ReadAll(io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
		}
		lastErr = err

		if attempt < len(httpRetryDelays) {
			time.Sleep(httpRetryDelays[attempt])
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("request %s failed after %d attempts: %w", req.URL.String(), attempts, lastErr)
	}
	return nil, fmt.Errorf("request %s failed after %d attempts with HTTP %d: %s", req.URL.String(), attempts, lastStatus, strings.TrimSpace(string(lastBody)))
}

func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || (status >= http.StatusInternalServerError && status <= 599)
}

func parseSpec(specPath string, body []byte) ([]generatedOperation, error) {
	var spec openAPISpec
	if err := yaml.Unmarshal(body, &spec); err != nil {
		return nil, err
	}
	var ops []generatedOperation
	for p, methods := range spec.Paths {
		for method, opNode := range methods {
			method = strings.ToUpper(method)
			if !isHTTPMethod(method) {
				continue
			}
			var op operation
			if err := opNode.Decode(&op); err != nil {
				return nil, fmt.Errorf("decode operation %s %s from %s: %w", method, p, specPath, err)
			}
			id := op.OperationID
			if id == "" {
				id = stableOperationID(specPath, method, p)
			}
			ops = append(ops, generatedOperation{
				ID:      id,
				Spec:    specPath,
				Method:  method,
				Path:    p,
				Summary: strings.TrimSpace(op.Summary),
				Tags:    op.Tags,
			})
		}
	}
	return ops, nil
}

func isHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func stableOperationID(parts ...string) string {
	joined := strings.Join(parts, "_")
	re := regexp.MustCompile(`[^A-Za-z0-9]+`)
	return strings.Trim(re.ReplaceAllString(joined, "_"), "_")
}

func writeOperations(filename string, tag string, ops map[string]generatedOperation) error {
	if err := os.MkdirAll(path.Dir(filename), 0o755); err != nil {
		return err
	}
	keys := sortedKeys(ops)
	var buf bytes.Buffer
	buf.WriteString("// Code generated by cmd/polaris-provider-gen. DO NOT EDIT.\n\n")
	buf.WriteString("package generated\n\n")
	buf.WriteString("type Operation struct {\n")
	buf.WriteString("\tID string\n\tSpec string\n\tMethod string\n\tPath string\n\tSummary string\n\tTags []string\n")
	buf.WriteString("}\n\n")
	fmt.Fprintf(&buf, "var ReleaseTag = %q\n\n", tag)
	buf.WriteString("var GeneratedAt = \"reproducible\"\n\n")
	buf.WriteString("var Operations = map[string]Operation{\n")
	for _, key := range keys {
		op := ops[key]
		fmt.Fprintf(&buf, "\t%q: {ID: %q, Spec: %q, Method: %q, Path: %q, Summary: %q, Tags: %#v},\n", op.ID, op.ID, op.Spec, op.Method, op.Path, op.Summary, op.Tags)
	}
	buf.WriteString("}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(filename, formatted, 0o644)
}

func writeDocs(filename, tag string, ops map[string]generatedOperation) error {
	if err := os.MkdirAll(path.Dir(filename), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Generated Polaris Operations\n\nRelease: `%s`\n\n", tag)
	buf.WriteString("This inventory is generated by `cmd/polaris-provider-gen` and is refreshed by `make generate`.\n\n")
	buf.WriteString("| Operation ID | Method | Path | Tags | Summary | Spec |\n")
	buf.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, key := range sortedKeys(ops) {
		op := ops[key]
		fmt.Fprintf(&buf, "| `%s` | `%s` | `%s` | %s | %s | `%s` |\n", op.ID, op.Method, op.Path, markdownCell(strings.Join(op.Tags, ", ")), markdownCell(op.Summary), op.Spec)
	}
	return os.WriteFile(filename, buf.Bytes(), 0o644)
}

func cacheSpec(root, tag, specPath string, body []byte) {
	filename := path.Join(root, tag, specPath)
	if err := os.MkdirAll(path.Dir(filename), 0o755); err != nil {
		die("cache mkdir failed: %v", err)
	}
	if err := os.WriteFile(filename, body, 0o644); err != nil {
		die("cache write failed: %v", err)
	}
}

func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func must(err error) {
	if err != nil {
		die("%v", err)
	}
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
