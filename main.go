package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/tsukubatexas/terraform-provider-polaris/internal/provider"
)

var version = "dev"

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New(version),
	})
}
