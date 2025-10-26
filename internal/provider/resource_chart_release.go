package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &ChartReleaseResource{}
var _ resource.ResourceWithImportState = &ChartReleaseResource{}

func NewChartReleaseResource() resource.Resource {
	return &ChartReleaseResource{}
}

type ChartReleaseResource struct {
	client *truenas.Client
}

type ChartReleaseResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ReleaseName types.String `tfsdk:"release_name"`
	Catalog     types.String `tfsdk:"catalog"`
	Train       types.String `tfsdk:"train"`
	Item        types.String `tfsdk:"item"`
	Version     types.String `tfsdk:"version"`
	Values      types.String `tfsdk:"values"`
	Status      types.String `tfsdk:"status"`
}

func (r *ChartReleaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chart_release"
}

func (r *ChartReleaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kubernetes application (chart release) on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Chart release identifier (same as release_name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"release_name": schema.StringAttribute{
				MarkdownDescription: "Name of the chart release",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"catalog": schema.StringAttribute{
				MarkdownDescription: "Catalog name (e.g., TRUENAS)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"train": schema.StringAttribute{
				MarkdownDescription: "Catalog train (e.g., charts, community, enterprise)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"item": schema.StringAttribute{
				MarkdownDescription: "Chart item name (e.g., plex, nextcloud, minecraft)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Chart version to deploy",
				Required:            true,
			},
			"values": schema.StringAttribute{
				MarkdownDescription: "Chart values in JSON format",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Current status of the chart release",
				Computed:            true,
			},
		},
	}
}

func (r *ChartReleaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ChartReleaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ChartReleaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"release_name": data.ReleaseName.ValueString(),
		"catalog":      data.Catalog.ValueString(),
		"train":        data.Train.ValueString(),
		"item":         data.Item.ValueString(),
		"version":      data.Version.ValueString(),
	}

	// Parse values if provided
	if !data.Values.IsNull() && data.Values.ValueString() != "" {
		var values map[string]interface{}
		if err := json.Unmarshal([]byte(data.Values.ValueString()), &values); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse values JSON: %s", err))
			return
		}
		createReq["values"] = values
	}

	respBody, err := r.client.Post("/chart/release", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create chart release, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if id, ok := result["id"].(string); ok {
		data.ID = types.StringValue(id)
	} else if name, ok := result["name"].(string); ok {
		data.ID = types.StringValue(name)
	}

	r.readChartRelease(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChartReleaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ChartReleaseResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readChartRelease(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChartReleaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ChartReleaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Version.IsNull() {
		updateReq["version"] = data.Version.ValueString()
	}

	// Parse values if provided
	if !data.Values.IsNull() && data.Values.ValueString() != "" {
		var values map[string]interface{}
		if err := json.Unmarshal([]byte(data.Values.ValueString()), &values); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse values JSON: %s", err))
			return
		}
		updateReq["values"] = values
	}

	endpoint := fmt.Sprintf("/chart/release/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update chart release, got error: %s", err))
		return
	}

	r.readChartRelease(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChartReleaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ChartReleaseResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/chart/release/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete chart release, got error: %s", err))
		return
	}
}

func (r *ChartReleaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ChartReleaseResource) readChartRelease(ctx context.Context, data *ChartReleaseResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/chart/release/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read chart release, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if name, ok := result["name"].(string); ok {
		data.ReleaseName = types.StringValue(name)
		data.ID = types.StringValue(name)
	}
	if catalog, ok := result["catalog"].(string); ok {
		data.Catalog = types.StringValue(catalog)
	}
	if train, ok := result["catalog_train"].(string); ok {
		data.Train = types.StringValue(train)
	}
	if chart, ok := result["chart_metadata"].(map[string]interface{}); ok {
		if name, ok := chart["name"].(string); ok {
			data.Item = types.StringValue(name)
		}
		if version, ok := chart["version"].(string); ok {
			data.Version = types.StringValue(version)
		}
	}
	if status, ok := result["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if config, ok := result["config"].(map[string]interface{}); ok {
		configJSON, _ := json.Marshal(config)
		data.Values = types.StringValue(string(configJSON))
	}
}

