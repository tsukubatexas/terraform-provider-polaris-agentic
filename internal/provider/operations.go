package provider

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/tsukubatexas/terraform-provider-polaris-agentic/internal/generated"
)

var pathParamPattern = regexp.MustCompile(`\{([^{}]+)\}`)

func operationByID(id string) (generated.Operation, error) {
	op, ok := generated.Operations[id]
	if !ok {
		return generated.Operation{}, fmt.Errorf("unknown Polaris OpenAPI operation_id %q", id)
	}
	return op, nil
}

func expandPath(pathTemplate string, params map[string]string) (string, error) {
	if pathTemplate == "" {
		return "", fmt.Errorf("path is required")
	}
	missing := []string{}
	expanded := pathParamPattern.ReplaceAllStringFunc(pathTemplate, func(match string) string {
		name := strings.Trim(match, "{}")
		value, ok := params[name]
		if !ok || value == "" {
			missing = append(missing, name)
			return match
		}
		return url.PathEscape(value)
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("missing path_params: %s", strings.Join(missing, ", "))
	}
	if !strings.HasPrefix(expanded, "/") {
		expanded = "/" + expanded
	}
	return expanded, nil
}
