package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type grippdProvider struct{ version string }

// New returns the provider factory providerserver.Serve wants.
func New(version string) func() provider.Provider {
	return func() provider.Provider { return &grippdProvider{version: version} }
}

func (p *grippdProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "grippd"
	resp.Version = p.version
}

// The provider takes no configuration: it only records declarations, and grippd resolves the
// values itself when an app deploys.
func (p *grippdProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Declares the values grippd keeps for a project (generated, asked for or minted). " +
			"Only the declaration is in OpenTofu's state; values never are.",
	}
}

func (p *grippdProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}

func (p *grippdProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{newEnvResource}
}

func (p *grippdProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}
