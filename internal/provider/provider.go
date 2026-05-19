package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"endpoint": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_ENDPOINT", nil),
					Description: "Base Polaris API endpoint, for example https://polaris.example/api/management/v1.",
				},
				"realm": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_REALM", nil),
					Description: "Optional Polaris realm sent as Polaris-Realm header.",
				},
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_TOKEN", nil),
					Description: "Bearer token. If omitted, client credentials can be used.",
				},
				"client_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_CLIENT_ID", nil),
				},
				"client_secret": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_CLIENT_SECRET", nil),
				},
				"oauth_token_url": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_OAUTH_TOKEN_URL", nil),
				},
				"oauth_scope": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("POLARIS_OAUTH_SCOPE", nil),
				},
				"insecure_skip_tls_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Disable TLS certificate verification. Use only in isolated tests.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"polaris_rest_resource": restResource(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"polaris_rest_call": restCallDataSource(),
			},
		}

		p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			endpoint, _ := d.Get("endpoint").(string)
			if endpoint == "" {
				return nil, diag.Errorf("endpoint is required. Set provider endpoint or POLARIS_ENDPOINT")
			}
			client, err := newClient(clientConfig{
				Endpoint:              endpoint,
				Realm:                 stringValue(d, "realm"),
				Token:                 stringValue(d, "token"),
				ClientID:              stringValue(d, "client_id"),
				ClientSecret:          stringValue(d, "client_secret"),
				OAuthTokenURL:         stringValue(d, "oauth_token_url"),
				OAuthScope:            stringValue(d, "oauth_scope"),
				InsecureSkipTLSVerify: d.Get("insecure_skip_tls_verify").(bool),
			})
			if err != nil {
				return nil, diag.FromErr(err)
			}
			client.UserAgent = "terraform-provider-polaris/" + version
			return client, nil
		}

		return p
	}
}

func stringValue(d *schema.ResourceData, key string) string {
	value, _ := d.Get(key).(string)
	return value
}
