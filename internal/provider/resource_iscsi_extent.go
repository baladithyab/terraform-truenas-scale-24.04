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

var _ resource.Resource = &ISCSIExtentResource{}
var _ resource.ResourceWithImportState = &ISCSIExtentResource{}

func NewISCSIExtentResource() resource.Resource {
	return &ISCSIExtentResource{}
}

type ISCSIExtentResource struct {
	client *truenas.Client
}

type ISCSIExtentResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Disk           types.String `tfsdk:"disk"`
	Path           types.String `tfsdk:"path"`
	Filesize       types.Int64  `tfsdk:"filesize"`
	Comment        types.String `tfsdk:"comment"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	ReadOnly       types.Bool   `tfsdk:"readonly"`
	Blocksize      types.Int64  `tfsdk:"blocksize"`
	PBlocksize     types.Bool   `tfsdk:"pblocksize"`
	AvailThreshold types.Int64  `tfsdk:"avail_threshold"`
	Serial         types.String `tfsdk:"serial"`
	RPM            types.String `tfsdk:"rpm"`
	Xen            types.Bool   `tfsdk:"xen"`
	InsecureTPC    types.Bool   `tfsdk:"insecure_tpc"`
}

func (r *ISCSIExtentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_extent"
}

func (r *ISCSIExtentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an iSCSI extent (storage) on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Extent identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Extent name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Extent type (DISK, FILE)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disk": schema.StringAttribute{
				MarkdownDescription: "Disk device (for DISK type)",
				Optional:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "File path (for FILE type)",
				Optional:            true,
			},
			"filesize": schema.Int64Attribute{
				MarkdownDescription: "File size in bytes (for FILE type)",
				Optional:            true,
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable extent",
				Optional:            true,
				Computed:            true,
			},
			"readonly": schema.BoolAttribute{
				MarkdownDescription: "Read-only extent",
				Optional:            true,
				Computed:            true,
			},
			"blocksize": schema.Int64Attribute{
				MarkdownDescription: "Block size (512, 1024, 2048, 4096)",
				Optional:            true,
				Computed:            true,
			},
			"pblocksize": schema.BoolAttribute{
				MarkdownDescription: "Use physical block size",
				Optional:            true,
				Computed:            true,
			},
			"avail_threshold": schema.Int64Attribute{
				MarkdownDescription: "Available space threshold percentage",
				Optional:            true,
				Computed:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "Serial number",
				Optional:            true,
				Computed:            true,
			},
			"rpm": schema.StringAttribute{
				MarkdownDescription: "RPM (SSD, 5400, 7200, 10000, 15000)",
				Optional:            true,
				Computed:            true,
			},
			"xen": schema.BoolAttribute{
				MarkdownDescription: "Xen compatibility mode",
				Optional:            true,
				Computed:            true,
			},
			"insecure_tpc": schema.BoolAttribute{
				MarkdownDescription: "Allow insecure third-party copy",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *ISCSIExtentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSIExtentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ISCSIExtentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	}

	if !data.Disk.IsNull() {
		createReq["disk"] = data.Disk.ValueString()
	}
	if !data.Path.IsNull() {
		createReq["path"] = data.Path.ValueString()
	}
	if !data.Filesize.IsNull() {
		createReq["filesize"] = data.Filesize.ValueInt64()
	}
	if !data.Comment.IsNull() {
		createReq["comment"] = data.Comment.ValueString()
	}
	if !data.Enabled.IsNull() {
		createReq["enabled"] = data.Enabled.ValueBool()
	}
	if !data.ReadOnly.IsNull() {
		createReq["ro"] = data.ReadOnly.ValueBool()
	}
	if !data.Blocksize.IsNull() {
		createReq["blocksize"] = data.Blocksize.ValueInt64()
	}
	if !data.PBlocksize.IsNull() {
		createReq["pblocksize"] = data.PBlocksize.ValueBool()
	}
	if !data.AvailThreshold.IsNull() {
		createReq["avail_threshold"] = data.AvailThreshold.ValueInt64()
	}
	if !data.Serial.IsNull() {
		createReq["serial"] = data.Serial.ValueString()
	}
	if !data.RPM.IsNull() {
		createReq["rpm"] = data.RPM.ValueString()
	}
	if !data.Xen.IsNull() {
		createReq["xen"] = data.Xen.ValueBool()
	}
	if !data.InsecureTPC.IsNull() {
		createReq["insecure_tpc"] = data.InsecureTPC.ValueBool()
	}

	respBody, err := r.client.Post("/iscsi/extent", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create iSCSI extent, got error: %s", err))
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

	r.readISCSIExtent(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSIExtentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ISCSIExtentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readISCSIExtent(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSIExtentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ISCSIExtentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}
	if !data.Comment.IsNull() {
		updateReq["comment"] = data.Comment.ValueString()
	}
	if !data.Enabled.IsNull() {
		updateReq["enabled"] = data.Enabled.ValueBool()
	}
	if !data.ReadOnly.IsNull() {
		updateReq["ro"] = data.ReadOnly.ValueBool()
	}

	endpoint := fmt.Sprintf("/iscsi/extent/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update iSCSI extent, got error: %s", err))
		return
	}

	r.readISCSIExtent(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSIExtentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ISCSIExtentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/iscsi/extent/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete iSCSI extent, got error: %s", err))
		return
	}
}

func (r *ISCSIExtentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ISCSIExtentResource) readISCSIExtent(ctx context.Context, data *ISCSIExtentResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/iscsi/extent/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read iSCSI extent, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if extentType, ok := result["type"].(string); ok {
		data.Type = types.StringValue(extentType)
	}
	if disk, ok := result["disk"].(string); ok && disk != "" {
		data.Disk = types.StringValue(disk)
	}
	if path, ok := result["path"].(string); ok && path != "" {
		data.Path = types.StringValue(path)
	}
	if filesize, ok := result["filesize"].(float64); ok {
		data.Filesize = types.Int64Value(int64(filesize))
	}
	if comment, ok := result["comment"].(string); ok {
		data.Comment = types.StringValue(comment)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		data.Enabled = types.BoolValue(enabled)
	}
	if ro, ok := result["ro"].(bool); ok {
		data.ReadOnly = types.BoolValue(ro)
	}
	if blocksize, ok := result["blocksize"].(float64); ok {
		data.Blocksize = types.Int64Value(int64(blocksize))
	}
}

