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
var _ datasource.DataSource = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	client *client.Client
}

// UsersDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	Users          []UserModel  `tfsdk:"users"`
}

// UserModel describes a single user.
type UserModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Email    types.String `tfsdk:"email"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Groups   types.List   `tfsdk:"groups"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all Pritunl users in an organization.",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{
				Description: "Organization ID to list users from.",
				Required:    true,
			},
			"users": schema.ListNestedAttribute{
				Description: "List of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "User ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Username.",
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
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := d.client.ListUsers(data.OrganizationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read users: %s", err))
		return
	}

	data.Users = make([]UserModel, len(users))
	for i, user := range users {
		var groupsList types.List
		if len(user.Groups) > 0 {
			var diags = resp.Diagnostics
			groupsList, diags = types.ListValueFrom(ctx, types.StringType, user.Groups)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			groupsList = types.ListNull(types.StringType)
		}

		data.Users[i] = UserModel{
			ID:       types.StringValue(user.ID),
			Name:     types.StringValue(user.Name),
			Email:    types.StringValue(user.Email),
			Disabled: types.BoolValue(user.Disabled),
			Groups:   groupsList,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
