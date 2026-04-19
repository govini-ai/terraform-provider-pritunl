package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/govini-ai/terraform-provider-pritunl/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RouteResource{}
var _ resource.ResourceWithImportState = &RouteResource{}

func NewRouteResource() resource.Resource {
	return &RouteResource{}
}

// RouteResource defines the resource implementation.
type RouteResource struct {
	client *client.Client
}

// RouteResourceModel describes the resource data model.
type RouteResourceModel struct {
	ID       types.String `tfsdk:"id"`
	ServerID types.String `tfsdk:"server_id"`
	Network  types.String `tfsdk:"network"`
	Comment  types.String `tfsdk:"comment"`
	Nat      types.Bool   `tfsdk:"nat"`
}

func (r *RouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_route"
}

func (r *RouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a route on a Pritunl VPN server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Route ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "Server ID this route belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network": schema.StringAttribute{
				Description: "Network CIDR for the route.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Description: "Comment/description for the route.",
				Optional:    true,
			},
			"nat": schema.BoolAttribute{
				Description: "Enable NAT for this route.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *RouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ServerID.ValueString()
	route := &client.Route{
		Network: data.Network.ValueString(),
		Comment: data.Comment.ValueString(),
		Nat:     data.Nat.ValueBool(),
	}

	var created *client.Route
	err := r.client.WithServerStopped(serverID, func() error {
		tflog.Debug(ctx, "Creating route", map[string]interface{}{
			"server_id": serverID,
			"network":   route.Network,
		})

		var createErr error
		created, createErr = r.client.CreateRoute(serverID, route)
		return createErr
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create route: %s", err))
		return
	}

	data.ID = types.StringValue(created.ID)
	data.Network = types.StringValue(created.Network)
	data.Comment = types.StringValue(created.Comment)
	data.Nat = types.BoolValue(created.Nat)

	tflog.Trace(ctx, "Created route", map[string]interface{}{"id": created.ID})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	route, err := r.client.GetRoute(data.ServerID.ValueString(), data.ID.ValueString())
	if err != nil {
		// Check if route was not found (drift detected)
		if strings.Contains(err.Error(), "not found") {
			tflog.Warn(ctx, "Route not found, removing from state", map[string]interface{}{
				"server_id": data.ServerID.ValueString(),
				"route_id":  data.ID.ValueString(),
			})
			// Remove resource from state by not setting it
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read route: %s", err))
		return
	}

	data.ID = types.StringValue(route.ID)
	data.Network = types.StringValue(route.Network)
	data.Comment = types.StringValue(route.Comment)
	data.Nat = types.BoolValue(route.Nat)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ServerID.ValueString()
	route := &client.Route{
		Network: data.Network.ValueString(),
		Comment: data.Comment.ValueString(),
		Nat:     data.Nat.ValueBool(),
	}

	var updated *client.Route
	err := r.client.WithServerStopped(serverID, func() error {
		tflog.Debug(ctx, "Updating route", map[string]interface{}{
			"server_id": serverID,
			"route_id":  data.ID.ValueString(),
		})

		var updateErr error
		updated, updateErr = r.client.UpdateRoute(serverID, data.ID.ValueString(), route)
		return updateErr
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update route: %s", err))
		return
	}

	data.Network = types.StringValue(updated.Network)
	data.Comment = types.StringValue(updated.Comment)
	data.Nat = types.BoolValue(updated.Nat)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverID := data.ServerID.ValueString()

	err := r.client.WithServerStopped(serverID, func() error {
		tflog.Debug(ctx, "Deleting route", map[string]interface{}{
			"server_id": serverID,
			"route_id":  data.ID.ValueString(),
		})

		return r.client.DeleteRoute(serverID, data.ID.ValueString())
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete route: %s", err))
		return
	}
}

func (r *RouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: server_id/route_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID format: server_id/route_id",
		)
		return
	}

	serverID := parts[0]
	routeID := parts[1]

	route, err := r.client.GetRoute(serverID, routeID)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import route: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), route.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_id"), serverID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network"), route.Network)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("comment"), route.Comment)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("nat"), route.Nat)...)
}
