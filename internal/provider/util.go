package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func stringMap(d *schema.ResourceData, key string) map[string]string {
	raw, ok := d.GetOk(key)
	if !ok {
		return map[string]string{}
	}
	result := map[string]string{}
	for k, v := range raw.(map[string]interface{}) {
		result[k] = fmt.Sprint(v)
	}
	return result
}

func intSet(d *schema.ResourceData, key string, fallback []int) map[int]struct{} {
	raw, ok := d.GetOk(key)
	if !ok {
		result := map[int]struct{}{}
		for _, code := range fallback {
			result[code] = struct{}{}
		}
		return result
	}
	result := map[int]struct{}{}
	for _, v := range raw.([]interface{}) {
		result[v.(int)] = struct{}{}
	}
	return result
}

func checkStatus(status int, accepted map[int]struct{}, body string) error {
	if _, ok := accepted[status]; ok {
		return nil
	}
	return fmt.Errorf("unexpected HTTP status %d: %s", status, safeHTTPBody([]byte(body)))
}

func safeHTTPBody(body []byte) string {
	const maxBody = 4096
	if len(body) <= maxBody {
		return string(body)
	}
	return string(body[:maxBody]) + "... [truncated]"
}

func stableID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])[:24]
}

func extractJSONPath(body, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	var value interface{}
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		return "", err
	}
	current := value
	for _, part := range strings.Split(path, ".") {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("id_attribute %q could not traverse %q", path, part)
		}
		current, ok = obj[part]
		if !ok {
			return "", fmt.Errorf("id_attribute %q missing %q", path, part)
		}
	}
	return fmt.Sprint(current), nil
}
