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

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *truenas.Client
}

type UserResourceModel struct {
	ID        types.String `tfsdk:"id"`
	UID       types.Int64  `tfsdk:"uid"`
	Username  types.String `tfsdk:"username"`
	FullName  types.String `tfsdk:"full_name"`
	Email     types.String `tfsdk:"email"`
	Password  types.String `tfsdk:"password"`
	Group     types.Int64  `tfsdk:"group"`
	Home      types.String `tfsdk:"home"`
	Shell     types.String `tfsdk:"shell"`
	SshPubkey types.String `tfsdk:"sshpubkey"`
	Locked    types.Bool   `tfsdk:"locked"`
	Sudo      types.Bool   `tfsdk:"sudo"`
	SmbAuth   types.Bool   `tfsdk:"smb"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user account on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"uid": schema.Int64Attribute{
				MarkdownDescription: "User ID (UID). If not specified, next available UID will be used",
				Optional:            true,
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username",
				Required:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "Full name of the user",
				Optional:            true,
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address",
				Optional:            true,
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "User password",
				Optional:            true,
				Sensitive:           true,
			},
			"group": schema.Int64Attribute{
				MarkdownDescription: "Primary group ID",
				Optional:            true,
				Computed:            true,
			},
			"home": schema.StringAttribute{
				MarkdownDescription: "Home directory path",
				Optional:            true,
				Computed:            true,
			},
			"shell": schema.StringAttribute{
				MarkdownDescription: "Login shell (e.g., /bin/bash, /usr/bin/zsh)",
				Optional:            true,
				Computed:            true,
			},
			"sshpubkey": schema.StringAttribute{
				MarkdownDescription: "SSH public key",
				Optional:            true,
				Computed:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Lock the account",
				Optional:            true,
				Computed:            true,
			},
			"sudo": schema.BoolAttribute{
				MarkdownDescription: "Allow sudo access",
				Optional:            true,
				Computed:            true,
			},
			"smb": schema.BoolAttribute{
				MarkdownDescription: "Enable SMB authentication",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"username": data.Username.ValueString(),
	}

	if !data.UID.IsNull() {
		createReq["uid"] = data.UID.ValueInt64()
	}
	if !data.FullName.IsNull() {
		createReq["full_name"] = data.FullName.ValueString()
	}
	if !data.Email.IsNull() {
		createReq["email"] = data.Email.ValueString()
	}
	if !data.Password.IsNull() {
		createReq["password"] = data.Password.ValueString()
	}
	if !data.Group.IsNull() {
		createReq["group"] = data.Group.ValueInt64()
	}
	if !data.Home.IsNull() {
		createReq["home"] = data.Home.ValueString()
	}
	if !data.Shell.IsNull() {
		createReq["shell"] = data.Shell.ValueString()
	}
	if !data.SshPubkey.IsNull() {
		createReq["sshpubkey"] = data.SshPubkey.ValueString()
	}
	if !data.Locked.IsNull() {
		createReq["locked"] = data.Locked.ValueBool()
	}
	if !data.Sudo.IsNull() {
		createReq["sudo"] = data.Sudo.ValueBool()
	}
	if !data.SmbAuth.IsNull() {
		createReq["smb"] = data.SmbAuth.ValueBool()
	}

	respBody, err := r.client.Post("/user", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
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

	r.readUser(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readUser(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Username.IsNull() {
		updateReq["username"] = data.Username.ValueString()
	}
	if !data.FullName.IsNull() {
		updateReq["full_name"] = data.FullName.ValueString()
	}
	if !data.Email.IsNull() {
		updateReq["email"] = data.Email.ValueString()
	}
	if !data.Password.IsNull() {
		updateReq["password"] = data.Password.ValueString()
	}
	if !data.Group.IsNull() {
		updateReq["group"] = data.Group.ValueInt64()
	}
	if !data.Home.IsNull() {
		updateReq["home"] = data.Home.ValueString()
	}
	if !data.Shell.IsNull() {
		updateReq["shell"] = data.Shell.ValueString()
	}
	if !data.SshPubkey.IsNull() {
		updateReq["sshpubkey"] = data.SshPubkey.ValueString()
	}
	if !data.Locked.IsNull() {
		updateReq["locked"] = data.Locked.ValueBool()
	}
	if !data.Sudo.IsNull() {
		updateReq["sudo"] = data.Sudo.ValueBool()
	}
	if !data.SmbAuth.IsNull() {
		updateReq["smb"] = data.SmbAuth.ValueBool()
	}

	endpoint := fmt.Sprintf("/user/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	r.readUser(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/user/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UserResource) readUser(ctx context.Context, data *UserResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/user/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if uid, ok := result["uid"].(float64); ok {
		data.UID = types.Int64Value(int64(uid))
	}
	if username, ok := result["username"].(string); ok {
		data.Username = types.StringValue(username)
	}
	if fullName, ok := result["full_name"].(string); ok {
		data.FullName = types.StringValue(fullName)
	}
	if email, ok := result["email"].(string); ok {
		data.Email = types.StringValue(email)
	}
	if group, ok := result["group"].(map[string]interface{}); ok {
		if gid, ok := group["id"].(float64); ok {
			data.Group = types.Int64Value(int64(gid))
		}
	}
	if home, ok := result["home"].(string); ok {
		data.Home = types.StringValue(home)
	}
	if shell, ok := result["shell"].(string); ok {
		data.Shell = types.StringValue(shell)
	}
}
