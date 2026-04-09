package datasources

import (
	"context"
	"fmt"

	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *client.Client
}

// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Email          types.String `tfsdk:"email"`
	Disabled       types.Bool   `tfsdk:"disabled"`
	Groups         types.List   `tfsdk:"groups"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a Pritunl user by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User ID. Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Organization ID the user belongs to.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Username. Either id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "User email address.",
				Computed:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the user is disabled.",
				Computed:    true,
			},
			"groups": schema.ListAttribute{
				Description: "List of groups the user belongs to.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var user *client.User
	var err error

	orgID := data.OrganizationID.ValueString()

	if !data.ID.IsNull() {
		user, err = d.client.GetUser(orgID, data.ID.ValueString())
	} else if !data.Name.IsNull() {
		user, err = d.client.GetUserByName(orgID, data.Name.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user: %s", err))
		return
	}

	data.ID = types.StringValue(user.ID)
	data.Name = types.StringValue(user.Name)
	data.Email = types.StringValue(user.Email)
	data.Disabled = types.BoolValue(user.Disabled)

	if len(user.Groups) > 0 {
		groupsList, diags := types.ListValueFrom(ctx, types.StringType, user.Groups)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Groups = groupsList
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
