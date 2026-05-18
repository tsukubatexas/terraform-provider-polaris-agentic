package provider

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/tsukubatexas/terraform-provider-polaris-agentic/internal/generated"
)

func TestGeneratedOperationsInventoryMatchesDocs(t *testing.T) {
	root := repoRoot(t)
	docPath := filepath.Join(root, "docs", "generated-operations.md")

	body, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read docs: %v", err)
	}

	lines := strings.Split(string(body), "\n")
	release := parseDocsRelease(lines)
	if release == "" {
		t.Fatalf("missing release line in %s", docPath)
	}
	if release != generated.ReleaseTag {
		t.Fatalf("docs release %q != generated release %q", release, generated.ReleaseTag)
	}

	docIDs := parseDocsOperationIDs(lines)
	if len(docIDs) != len(generated.Operations) {
		t.Fatalf("docs operation count %d != generated operations %d", len(docIDs), len(generated.Operations))
	}

	sortedDocIDs := append([]string(nil), docIDs...)
	sort.Strings(sortedDocIDs)
	if !slicesEqual(sortedDocIDs, docIDs) {
		t.Fatalf("docs operation IDs are not sorted")
	}

	docSet := make(map[string]struct{}, len(docIDs))
	for _, id := range docIDs {
		if _, ok := docSet[id]; ok {
			t.Fatalf("duplicate operation id %q in docs", id)
		}
		docSet[id] = struct{}{}
	}

	for id, op := range generated.Operations {
		if id != op.ID {
			t.Fatalf("generated operation key %q != op.ID %q", id, op.ID)
		}
		if op.Spec == "" || op.Method == "" || op.Path == "" {
			t.Fatalf("generated operation %q has empty fields: %#v", id, op)
		}
		if _, ok := docSet[id]; !ok {
			t.Fatalf("operation %q missing from docs inventory", id)
		}
	}

	for id := range docSet {
		if _, ok := generated.Operations[id]; !ok {
			t.Fatalf("docs inventory includes unknown operation %q", id)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func parseDocsRelease(lines []string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, "Release: `") && strings.HasSuffix(line, "`") {
			return strings.TrimSuffix(strings.TrimPrefix(line, "Release: `"), "`")
		}
	}
	return ""
}

func parseDocsOperationIDs(lines []string) []string {
	var ids []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "| `") {
			continue
		}
		parts := strings.Split(line, "`")
		if len(parts) < 3 {
			continue
		}
		ids = append(ids, parts[1])
	}
	return ids
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
