package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func restCallDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRestCallRead,
		Schema: map[string]*schema.Schema{
			"operation_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "OpenAPI operationId from the generated Polaris operation registry.",
			},
			"method": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTP method. Required when operation_id is not set.",
			},
			"path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTTP path template. Required when operation_id is not set.",
			},
			"path_params": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Values for OpenAPI path placeholders.",
			},
			"query_params": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Query string parameters.",
			},
			"headers": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Additional HTTP headers.",
			},
			"body": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Optional JSON request body for advanced read-style operations.",
			},
			"expected_status_codes": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Accepted HTTP status codes. Defaults to 200.",
			},
			"status_code": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"response_body": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceRestCallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	method, path, err := methodAndPath(d, "operation_id", "method", "path")
	if err != nil {
		return diag.FromErr(err)
	}
	resp, err := meta.(*Client).do(ctx, method, path, stringMap(d, "path_params"), stringMap(d, "query_params"), stringMap(d, "headers"), stringValue(d, "body"))
	if err != nil {
		return diag.FromErr(err)
	}
	if err := checkStatus(resp.StatusCode, intSet(d, "expected_status_codes", []int{200}), resp.Body); err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("status_code", resp.StatusCode)
	_ = d.Set("response_body", resp.Body)
	d.SetId(stableID(method, path, strings.Join(mapValues(stringMap(d, "path_params")), ","), strings.Join(mapValues(stringMap(d, "query_params")), ",")))
	return nil
}

func methodAndPath(d *schema.ResourceData, operationAttr, methodAttr, pathAttr string) (string, string, error) {
	if opID := stringValue(d, operationAttr); opID != "" {
		op, err := operationByID(opID)
		if err != nil {
			return "", "", err
		}
		return op.Method, op.Path, nil
	}
	method := stringValue(d, methodAttr)
	path := stringValue(d, pathAttr)
	if method == "" || path == "" {
		return "", "", fmt.Errorf("%s or %s/%s is required", operationAttr, methodAttr, pathAttr)
	}
	return strings.ToUpper(method), path, nil
}

func mapValues(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for k, v := range values {
		result = append(result, k+"="+v)
	}
	return result
}
