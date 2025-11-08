package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &StaticRouteResource{}
var _ resource.ResourceWithImportState = &StaticRouteResource{}

func NewStaticRouteResource() resource.Resource {
	return &StaticRouteResource{}
}

type StaticRouteResource struct {
	client *truenas.Client
}

type StaticRouteResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
	Description types.String `tfsdk:"description"`
}

func (r *StaticRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_route"
}

func (r *StaticRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a static route on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Route identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"destination": schema.StringAttribute{
				MarkdownDescription: "Destination network in CIDR notation (e.g., 192.168.2.0/24 or default for 0.0.0.0/0)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "Gateway IP address",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Route description",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *StaticRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*truenas.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *truenas.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *StaticRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StaticRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"destination": data.Destination.ValueString(),
		"gateway":     data.Gateway.ValueString(),
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}

	respBody, err := r.client.Post("/staticroute", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create static route, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(strconv.Itoa(int(id)))
	}

	r.readStaticRoute(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StaticRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readStaticRoute(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StaticRouteResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Gateway.IsNull() {
		updateReq["gateway"] = data.Gateway.ValueString()
	}
	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}

	endpoint := fmt.Sprintf("/staticroute/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update static route, got error: %s", err))
		return
	}

	r.readStaticRoute(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StaticRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StaticRouteResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/staticroute/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete static route, got error: %s", err))
		return
	}
}

func (r *StaticRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *StaticRouteResource) readStaticRoute(ctx context.Context, data *StaticRouteResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/staticroute/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read static route, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if destination, ok := result["destination"].(string); ok {
		data.Destination = types.StringValue(destination)
	}
	if gateway, ok := result["gateway"].(string); ok {
		data.Gateway = types.StringValue(gateway)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
}
