package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource defines the resource implementation.
type UserResource struct {
	client *client.Client
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Email          types.String `tfsdk:"email"`
	Disabled       types.Bool   `tfsdk:"disabled"`
	Groups         types.List   `tfsdk:"groups"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Pritunl user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "User ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "Organization ID the user belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Username.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "User email address.",
				Optional:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the user is disabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"groups": schema.ListAttribute{
				Description: "List of groups the user belongs to.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := &client.User{
		Name:     data.Name.ValueString(),
		Email:    data.Email.ValueString(),
		Disabled: data.Disabled.ValueBool(),
	}

	// Handle groups
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		var groups []string
		diags := data.Groups.ElementsAs(ctx, &groups, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.Groups = groups
	}

	tflog.Debug(ctx, "Creating user", map[string]interface{}{
		"organization_id": data.OrganizationID.ValueString(),
		"name":            user.Name,
	})

	created, err := r.client.CreateUser(data.OrganizationID.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user: %s", err))
		return
	}

	data.ID = types.StringValue(created.ID)
	data.Name = types.StringValue(created.Name)
	data.Email = types.StringValue(created.Email)
	data.Disabled = types.BoolValue(created.Disabled)

	if len(created.Groups) > 0 {
		groupsList, diags := types.ListValueFrom(ctx, types.StringType, created.Groups)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Groups = groupsList
	} else {
		data.Groups = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "Created user", map[string]interface{}{"id": created.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(data.OrganizationID.ValueString(), data.ID.ValueString())
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

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := &client.User{
		Name:     data.Name.ValueString(),
		Email:    data.Email.ValueString(),
		Disabled: data.Disabled.ValueBool(),
	}

	// Handle groups
	if !data.Groups.IsNull() && !data.Groups.IsUnknown() {
		var groups []string
		diags := data.Groups.ElementsAs(ctx, &groups, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		user.Groups = groups
	}

	tflog.Debug(ctx, "Updating user", map[string]interface{}{
		"organization_id": data.OrganizationID.ValueString(),
		"user_id":         data.ID.ValueString(),
	})

	updated, err := r.client.UpdateUser(data.OrganizationID.ValueString(), data.ID.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user: %s", err))
		return
	}

	data.Name = types.StringValue(updated.Name)
	data.Email = types.StringValue(updated.Email)
	data.Disabled = types.BoolValue(updated.Disabled)

	if len(updated.Groups) > 0 {
		groupsList, diags := types.ListValueFrom(ctx, types.StringType, updated.Groups)
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

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting user", map[string]interface{}{
		"organization_id": data.OrganizationID.ValueString(),
		"user_id":         data.ID.ValueString(),
	})

	err := r.client.DeleteUser(data.OrganizationID.ValueString(), data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: organization_id/user_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID format: organization_id/user_id",
		)
		return
	}

	orgID := parts[0]
	userID := parts[1]

	user, err := r.client.GetUser(orgID, userID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import user: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), user.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("organization_id"), orgID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), user.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), user.Email)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("disabled"), user.Disabled)...)
}
