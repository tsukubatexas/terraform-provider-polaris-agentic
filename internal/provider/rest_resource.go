package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func restResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRestCreate,
		ReadContext:   resourceRestRead,
		UpdateContext: resourceRestUpdate,
		DeleteContext: resourceRestDelete,
		Schema: map[string]*schema.Schema{
			"create_operation_id": operationSchema("OpenAPI operationId used for create."),
			"read_operation_id":   operationSchema("OpenAPI operationId used for read."),
			"update_operation_id": operationSchema("OpenAPI operationId used for update."),
			"delete_operation_id": operationSchema("OpenAPI operationId used for delete."),
			"create_method":       methodSchema("HTTP method for create when create_operation_id is not set."),
			"read_method":         methodSchema("HTTP method for read when read_operation_id is not set."),
			"update_method":       methodSchema("HTTP method for update when update_operation_id is not set."),
			"delete_method":       methodSchema("HTTP method for delete when delete_operation_id is not set."),
			"create_path":         pathSchema("HTTP path template for create when create_operation_id is not set."),
			"read_path":           pathSchema("HTTP path template for read when read_operation_id is not set."),
			"update_path":         pathSchema("HTTP path template for update when update_operation_id is not set."),
			"delete_path":         pathSchema("HTTP path template for delete when delete_operation_id is not set."),
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
				Description: "JSON request body used for create and update.",
			},
			"id_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dot-path in the JSON response used as Terraform ID, for example catalog.name.",
			},
			"expected_status_codes": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Description: "Accepted HTTP status codes for mutating calls. Defaults to 200, 201 and 204.",
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

func operationSchema(description string) *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Optional: true, Description: description}
}

func methodSchema(description string) *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Optional: true, Description: description}
}

func pathSchema(description string) *schema.Schema {
	return &schema.Schema{Type: schema.TypeString, Optional: true, Description: description}
}

func resourceRestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resp, method, path, err := runPhase(ctx, d, meta.(*Client), "create")
	if err != nil {
		return diag.FromErr(err)
	}
	id := ""
	if idAttr := stringValue(d, "id_attribute"); idAttr != "" && resp.Body != "" {
		id, err = extractJSONPath(resp.Body, idAttr)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if id == "" {
		id = stableID(method, path, stringValue(d, "body"))
	}
	d.SetId(id)
	_ = d.Set("status_code", resp.StatusCode)
	_ = d.Set("response_body", resp.Body)
	return resourceRestRead(ctx, d, meta)
}

func resourceRestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if stringValue(d, "read_operation_id") == "" && stringValue(d, "read_method") == "" && stringValue(d, "read_path") == "" {
		return nil
	}
	resp, _, _, err := runPhase(ctx, d, meta.(*Client), "read")
	if err != nil {
		return diag.FromErr(err)
	}
	if resp.StatusCode == 404 {
		d.SetId("")
		return nil
	}
	if err := checkStatus(resp.StatusCode, intSet(d, "expected_status_codes", []int{200}), resp.Body); err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("status_code", resp.StatusCode)
	_ = d.Set("response_body", resp.Body)
	return nil
}

func resourceRestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resp, _, _, err := runPhase(ctx, d, meta.(*Client), "update")
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("status_code", resp.StatusCode)
	_ = d.Set("response_body", resp.Body)
	return resourceRestRead(ctx, d, meta)
}

func resourceRestDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if stringValue(d, "delete_operation_id") == "" && stringValue(d, "delete_method") == "" && stringValue(d, "delete_path") == "" {
		d.SetId("")
		return nil
	}
	if _, _, _, err := runPhase(ctx, d, meta.(*Client), "delete"); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func runPhase(ctx context.Context, d *schema.ResourceData, client *Client, phase string) (*response, string, string, error) {
	method, path, err := methodAndPath(d, phase+"_operation_id", phase+"_method", phase+"_path")
	if err != nil {
		return nil, "", "", err
	}
	body := ""
	if phase == "create" || phase == "update" {
		body = stringValue(d, "body")
	}
	resp, err := client.do(ctx, method, path, stringMap(d, "path_params"), stringMap(d, "query_params"), stringMap(d, "headers"), body)
	if err != nil {
		return nil, "", "", err
	}
	if phase != "read" {
		if err := checkStatus(resp.StatusCode, intSet(d, "expected_status_codes", []int{200, 201, 204}), resp.Body); err != nil {
			return nil, "", "", err
		}
	}
	return resp, method, path, nil
}
