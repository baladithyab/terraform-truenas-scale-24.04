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

var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

type GroupResource struct {
	client *truenas.Client
}

type GroupResourceModel struct {
	ID      types.String `tfsdk:"id"`
	GID     types.Int64  `tfsdk:"gid"`
	Name    types.String `tfsdk:"name"`
	Sudo    types.Bool   `tfsdk:"sudo"`
	SmbAuth types.Bool   `tfsdk:"smb"`
	Users   types.List   `tfsdk:"users"`
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a group on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Group identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gid": schema.Int64Attribute{
				MarkdownDescription: "Group ID (GID). If not specified, next available GID will be used",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Group name",
				Required:            true,
			},
			"sudo": schema.BoolAttribute{
				MarkdownDescription: "Allow sudo access for group members",
				Optional:            true,
				Computed:            true,
			},
			"smb": schema.BoolAttribute{
				MarkdownDescription: "Enable SMB authentication for group",
				Optional:            true,
				Computed:            true,
			},
			"users": schema.ListAttribute{
				MarkdownDescription: "List of user IDs that are members of this group",
				Optional:            true,
				Computed:            true,
				ElementType:         types.Int64Type,
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	if !data.GID.IsNull() {
		createReq["gid"] = data.GID.ValueInt64()
	}
	if !data.Sudo.IsNull() {
		createReq["sudo"] = data.Sudo.ValueBool()
	}
	if !data.SmbAuth.IsNull() {
		createReq["smb"] = data.SmbAuth.ValueBool()
	}
	if !data.Users.IsNull() {
		var users []int64
		resp.Diagnostics.Append(data.Users.ElementsAs(ctx, &users, false)...)
		createReq["users"] = users
	}

	respBody, err := r.client.Post("/group", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err))
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

	r.readGroup(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readGroup(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}
	if !data.Sudo.IsNull() {
		updateReq["sudo"] = data.Sudo.ValueBool()
	}
	if !data.SmbAuth.IsNull() {
		updateReq["smb"] = data.SmbAuth.ValueBool()
	}
	if !data.Users.IsNull() {
		var users []int64
		resp.Diagnostics.Append(data.Users.ElementsAs(ctx, &users, false)...)
		updateReq["users"] = users
	}

	endpoint := fmt.Sprintf("/group/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update group, got error: %s", err))
		return
	}

	r.readGroup(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/group/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete group, got error: %s", err))
		return
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *GroupResource) readGroup(ctx context.Context, data *GroupResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/group/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if gid, ok := result["gid"].(float64); ok {
		data.GID = types.Int64Value(int64(gid))
	}
	if name, ok := result["group"].(string); ok {
		data.Name = types.StringValue(name)
	}
	if sudo, ok := result["sudo"].(bool); ok {
		data.Sudo = types.BoolValue(sudo)
	}
	if smb, ok := result["smb"].(bool); ok {
		data.SmbAuth = types.BoolValue(smb)
	}
}

