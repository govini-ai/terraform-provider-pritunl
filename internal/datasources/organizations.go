package datasources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &OrganizationsDataSource{}

func NewOrganizationsDataSource() datasource.DataSource {
	return &OrganizationsDataSource{}
}

// OrganizationsDataSource defines the data source implementation.
type OrganizationsDataSource struct {
	client *client.Client
}

// OrganizationsDataSourceModel describes the data source data model.
type OrganizationsDataSourceModel struct {
	Organizations []OrganizationModel `tfsdk:"organizations"`
}

// OrganizationModel describes a single organization.
type OrganizationModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *OrganizationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

func (d *OrganizationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Pritunl organizations.",
		Attributes: map[string]schema.Attribute{
			"organizations": schema.ListNestedAttribute{
				Description: "List of organizations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Organization ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Organization name.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *OrganizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := d.client.ListOrganizations()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organizations: %s", err))
		return
	}

	data.Organizations = make([]OrganizationModel, len(orgs))
	for i, org := range orgs {
		data.Organizations[i] = OrganizationModel{
			ID:   types.StringValue(org.ID),
			Name: types.StringValue(org.Name),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
