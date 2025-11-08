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

var _ resource.Resource = &NFSShareResource{}
var _ resource.ResourceWithImportState = &NFSShareResource{}

func NewNFSShareResource() resource.Resource {
	return &NFSShareResource{}
}

type NFSShareResource struct {
	client *truenas.Client
}

type NFSShareResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Path     types.String `tfsdk:"path"`
	Comment  types.String `tfsdk:"comment"`
	Networks types.List   `tfsdk:"networks"`
	Hosts    types.List   `tfsdk:"hosts"`
	ReadOnly types.Bool   `tfsdk:"readonly"`
	Maproot  types.String `tfsdk:"maproot_user"`
	Mapall   types.String `tfsdk:"mapall_user"`
	Security types.List   `tfsdk:"security"`
	Enabled  types.Bool   `tfsdk:"enabled"`
}

func (r *NFSShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nfs_share"
}

func (r *NFSShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an NFS share on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "NFS share identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"networks": schema.ListAttribute{
				MarkdownDescription: "List of authorized networks (e.g., 192.168.1.0/24)",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"hosts": schema.ListAttribute{
				MarkdownDescription: "List of authorized hosts",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"readonly": schema.BoolAttribute{
				MarkdownDescription: "Export as read-only",
				Optional:            true,
				Computed:            true,
			},
			"maproot_user": schema.StringAttribute{
				MarkdownDescription: "Map root user to this user",
				Optional:            true,
				Computed:            true,
			},
			"mapall_user": schema.StringAttribute{
				MarkdownDescription: "Map all users to this user",
				Optional:            true,
				Computed:            true,
			},
			"security": schema.ListAttribute{
				MarkdownDescription: "Security mechanisms (SYS, KRB5, KRB5I, KRB5P)",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Enable this NFS share",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *NFSShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NFSShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NFSShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"path": data.Path.ValueString(),
	}

	if !data.Comment.IsNull() {
		createReq["comment"] = data.Comment.ValueString()
	}
	if !data.Networks.IsNull() {
		var networks []string
		resp.Diagnostics.Append(data.Networks.ElementsAs(ctx, &networks, false)...)
		createReq["networks"] = networks
	}

	// hosts and security are required by TrueNAS API - default to empty array if not specified
	if !data.Hosts.IsNull() && !data.Hosts.IsUnknown() {
		var hosts []string
		resp.Diagnostics.Append(data.Hosts.ElementsAs(ctx, &hosts, false)...)
		createReq["hosts"] = hosts
	} else {
		createReq["hosts"] = []string{} // Default: allow all hosts
	}

	if !data.ReadOnly.IsNull() {
		createReq["ro"] = data.ReadOnly.ValueBool()
	}
	if !data.Maproot.IsNull() {
		createReq["maproot_user"] = data.Maproot.ValueString()
	}
	if !data.Mapall.IsNull() {
		createReq["mapall_user"] = data.Mapall.ValueString()
	}

	// security is required by TrueNAS API - default to empty array if not specified
	if !data.Security.IsNull() && !data.Security.IsUnknown() {
		var security []string
		resp.Diagnostics.Append(data.Security.ElementsAs(ctx, &security, false)...)
		createReq["security"] = security
	} else {
		createReq["security"] = []string{} // Default: no security restrictions
	}

	if !data.Enabled.IsNull() {
		createReq["enabled"] = data.Enabled.ValueBool()
	}

	respBody, err := r.client.Post("/sharing/nfs", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create NFS share, got error: %s", err))
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

	r.readNFSShare(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NFSShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NFSShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readNFSShare(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NFSShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NFSShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Path.IsNull() {
		updateReq["path"] = data.Path.ValueString()
	}
	if !data.Comment.IsNull() {
		updateReq["comment"] = data.Comment.ValueString()
	}
	if !data.Networks.IsNull() {
		var networks []string
		resp.Diagnostics.Append(data.Networks.ElementsAs(ctx, &networks, false)...)
		updateReq["networks"] = networks
	}
	if !data.Hosts.IsNull() {
		var hosts []string
		resp.Diagnostics.Append(data.Hosts.ElementsAs(ctx, &hosts, false)...)
		updateReq["hosts"] = hosts
	}
	if !data.ReadOnly.IsNull() {
		updateReq["ro"] = data.ReadOnly.ValueBool()
	}
	if !data.Maproot.IsNull() {
		updateReq["maproot_user"] = data.Maproot.ValueString()
	}
	if !data.Mapall.IsNull() {
		updateReq["mapall_user"] = data.Mapall.ValueString()
	}
	if !data.Security.IsNull() {
		var security []string
		resp.Diagnostics.Append(data.Security.ElementsAs(ctx, &security, false)...)
		updateReq["security"] = security
	}
	if !data.Enabled.IsNull() {
		updateReq["enabled"] = data.Enabled.ValueBool()
	}

	endpoint := fmt.Sprintf("/sharing/nfs/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update NFS share, got error: %s", err))
		return
	}

	r.readNFSShare(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NFSShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NFSShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/sharing/nfs/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete NFS share, got error: %s", err))
		return
	}
}

func (r *NFSShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NFSShareResource) readNFSShare(ctx context.Context, data *NFSShareResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/sharing/nfs/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read NFS share, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Read ID
	if id, ok := result["id"].(float64); ok {
		data.ID = types.StringValue(strconv.Itoa(int(id)))
	}

	// Read path
	if path, ok := result["path"].(string); ok {
		data.Path = types.StringValue(path)
	}

	// Read comment
	if comment, ok := result["comment"].(string); ok {
		data.Comment = types.StringValue(comment)
	} else {
		data.Comment = types.StringNull()
	}

	// Read enabled
	if enabled, ok := result["enabled"].(bool); ok {
		data.Enabled = types.BoolValue(enabled)
	}

	// Read readonly
	if ro, ok := result["ro"].(bool); ok {
		data.ReadOnly = types.BoolValue(ro)
	}

	// Read networks list
	if networks, ok := result["networks"].([]interface{}); ok && len(networks) > 0 {
		networkList := make([]string, len(networks))
		for i, network := range networks {
			if networkStr, ok := network.(string); ok {
				networkList[i] = networkStr
			}
		}
		networkValues := make([]types.String, len(networkList))
		for i, network := range networkList {
			networkValues[i] = types.StringValue(network)
		}
		listValue, _ := types.ListValueFrom(ctx, types.StringType, networkValues)
		data.Networks = listValue
	} else {
		data.Networks = types.ListNull(types.StringType)
	}

	// Read hosts list
	if hosts, ok := result["hosts"].([]interface{}); ok && len(hosts) > 0 {
		hostList := make([]string, len(hosts))
		for i, host := range hosts {
			if hostStr, ok := host.(string); ok {
				hostList[i] = hostStr
			}
		}
		hostValues := make([]types.String, len(hostList))
		for i, host := range hostList {
			hostValues[i] = types.StringValue(host)
		}
		listValue, _ := types.ListValueFrom(ctx, types.StringType, hostValues)
		data.Hosts = listValue
	} else {
		// Default to empty list instead of null to match Create behavior
		emptyList, _ := types.ListValueFrom(ctx, types.StringType, []types.String{})
		data.Hosts = emptyList
	}

	// Read security list
	if security, ok := result["security"].([]interface{}); ok && len(security) > 0 {
		securityList := make([]string, len(security))
		for i, sec := range security {
			if secStr, ok := sec.(string); ok {
				securityList[i] = secStr
			}
		}
		securityValues := make([]types.String, len(securityList))
		for i, sec := range securityList {
			securityValues[i] = types.StringValue(sec)
		}
		listValue, _ := types.ListValueFrom(ctx, types.StringType, securityValues)
		data.Security = listValue
	} else {
		// Default to empty list instead of null to match Create behavior
		emptyList, _ := types.ListValueFrom(ctx, types.StringType, []types.String{})
		data.Security = emptyList
	}

	// Read maproot_user
	if maproot, ok := result["maproot_user"].(string); ok && maproot != "" {
		data.Maproot = types.StringValue(maproot)
	} else {
		data.Maproot = types.StringNull()
	}

	// Read mapall_user
	if mapall, ok := result["mapall_user"].(string); ok && mapall != "" {
		data.Mapall = types.StringValue(mapall)
	} else {
		data.Mapall = types.StringNull()
	}
}
