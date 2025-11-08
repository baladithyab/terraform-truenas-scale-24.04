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

var _ resource.Resource = &SMBShareResource{}
var _ resource.ResourceWithImportState = &SMBShareResource{}

func NewSMBShareResource() resource.Resource {
	return &SMBShareResource{}
}

type SMBShareResource struct {
	client *truenas.Client
}

type SMBShareResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Path       types.String `tfsdk:"path"`
	Comment    types.String `tfsdk:"comment"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Browsable  types.Bool   `tfsdk:"browsable"`
	Guestok    types.Bool   `tfsdk:"guestok"`
	ReadOnly   types.Bool   `tfsdk:"readonly"`
	Recyclebin types.Bool   `tfsdk:"recyclebin"`
	Shadowcopy types.Bool   `tfsdk:"shadowcopy"`
	Hostsallow types.List   `tfsdk:"hostsallow"`
	Hostsdeny  types.List   `tfsdk:"hostsdeny"`
}

func (r *SMBShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_smb_share"
}

func (r *SMBShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an SMB/CIFS share on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SMB share identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Share name",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Path to be shared",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment about the share",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable this SMB share",
				Optional:            true,
				Computed:            true,
			},
			"browsable": schema.BoolAttribute{
				MarkdownDescription: "Make share visible in network browser",
				Optional:            true,
				Computed:            true,
			},
			"guestok": schema.BoolAttribute{
				MarkdownDescription: "Allow guest access",
				Optional:            true,
				Computed:            true,
			},
			"readonly": schema.BoolAttribute{
				MarkdownDescription: "Export as read-only",
				Optional:            true,
				Computed:            true,
			},
			"recyclebin": schema.BoolAttribute{
				MarkdownDescription: "Enable recycle bin",
				Optional:            true,
				Computed:            true,
			},
			"shadowcopy": schema.BoolAttribute{
				MarkdownDescription: "Enable shadow copies (previous versions)",
				Optional:            true,
				Computed:            true,
			},
			"hostsallow": schema.ListAttribute{
				MarkdownDescription: "List of allowed hosts/networks",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"hostsdeny": schema.ListAttribute{
				MarkdownDescription: "List of denied hosts/networks",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (r *SMBShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SMBShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SMBShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
		"path": data.Path.ValueString(),
	}

	if !data.Comment.IsNull() {
		createReq["comment"] = data.Comment.ValueString()
	}
	if !data.Enabled.IsNull() {
		createReq["enabled"] = data.Enabled.ValueBool()
	}
	if !data.Browsable.IsNull() {
		createReq["browsable"] = data.Browsable.ValueBool()
	}
	if !data.Guestok.IsNull() {
		createReq["guestok"] = data.Guestok.ValueBool()
	}
	if !data.ReadOnly.IsNull() {
		createReq["ro"] = data.ReadOnly.ValueBool()
	}
	if !data.Recyclebin.IsNull() {
		createReq["recyclebin"] = data.Recyclebin.ValueBool()
	}
	if !data.Shadowcopy.IsNull() {
		createReq["shadowcopy"] = data.Shadowcopy.ValueBool()
	}
	if !data.Hostsallow.IsNull() {
		var hostsallow []string
		resp.Diagnostics.Append(data.Hostsallow.ElementsAs(ctx, &hostsallow, false)...)
		createReq["hostsallow"] = hostsallow
	}
	if !data.Hostsdeny.IsNull() {
		var hostsdeny []string
		resp.Diagnostics.Append(data.Hostsdeny.ElementsAs(ctx, &hostsdeny, false)...)
		createReq["hostsdeny"] = hostsdeny
	}

	respBody, err := r.client.Post("/sharing/smb", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create SMB share, got error: %s", err))
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

	r.readSMBShare(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMBShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SMBShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readSMBShare(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMBShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SMBShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}
	if !data.Path.IsNull() {
		updateReq["path"] = data.Path.ValueString()
	}
	if !data.Comment.IsNull() {
		updateReq["comment"] = data.Comment.ValueString()
	}
	if !data.Enabled.IsNull() {
		updateReq["enabled"] = data.Enabled.ValueBool()
	}
	if !data.Browsable.IsNull() {
		updateReq["browsable"] = data.Browsable.ValueBool()
	}
	if !data.Guestok.IsNull() {
		updateReq["guestok"] = data.Guestok.ValueBool()
	}
	if !data.ReadOnly.IsNull() {
		updateReq["ro"] = data.ReadOnly.ValueBool()
	}
	if !data.Recyclebin.IsNull() {
		updateReq["recyclebin"] = data.Recyclebin.ValueBool()
	}
	if !data.Shadowcopy.IsNull() {
		updateReq["shadowcopy"] = data.Shadowcopy.ValueBool()
	}

	endpoint := fmt.Sprintf("/sharing/smb/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update SMB share, got error: %s", err))
		return
	}

	r.readSMBShare(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMBShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SMBShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/sharing/smb/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete SMB share, got error: %s", err))
		return
	}
}

func (r *SMBShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SMBShareResource) readSMBShare(ctx context.Context, data *SMBShareResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/sharing/smb/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read SMB share, got error: %s", err))
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
	if path, ok := result["path"].(string); ok {
		data.Path = types.StringValue(path)
	}
	if comment, ok := result["comment"].(string); ok {
		data.Comment = types.StringValue(comment)
	}
	if enabled, ok := result["enabled"].(bool); ok {
		data.Enabled = types.BoolValue(enabled)
	}
}
