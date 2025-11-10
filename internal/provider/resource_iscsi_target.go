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
	"github.com/baladithyab/terraform-provider-truenas/internal/truenas"
)

var _ resource.Resource = &ISCSITargetResource{}
var _ resource.ResourceWithImportState = &ISCSITargetResource{}

func NewISCSITargetResource() resource.Resource {
	return &ISCSITargetResource{}
}

type ISCSITargetResource struct {
	client *truenas.Client
}

type ISCSITargetResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Alias        types.String `tfsdk:"alias"`
	Mode         types.String `tfsdk:"mode"`
	Groups       types.List   `tfsdk:"groups"`
	AuthNetworks types.List   `tfsdk:"auth_networks"`
}

func (r *ISCSITargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iscsi_target"
}

func (r *ISCSITargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an iSCSI target on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Target identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Target name (IQN will be generated)",
				Required:            true,
			},
			"alias": schema.StringAttribute{
				MarkdownDescription: "Target alias",
				Optional:            true,
				Computed:            true,
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "Target mode (ISCSI, FC, BOTH)",
				Optional:            true,
				Computed:            true,
			},
			"groups": schema.ListAttribute{
				MarkdownDescription: "List of portal group IDs",
				ElementType:         types.Int64Type,
				Optional:            true,
				Computed:            true,
			},
			"auth_networks": schema.ListAttribute{
				MarkdownDescription: "List of authorized networks (CIDR notation)",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *ISCSITargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ISCSITargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ISCSITargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	if !data.Alias.IsNull() {
		createReq["alias"] = data.Alias.ValueString()
	}
	if !data.Mode.IsNull() {
		createReq["mode"] = data.Mode.ValueString()
	}
	if !data.Groups.IsNull() {
		var groups []int64
		data.Groups.ElementsAs(ctx, &groups, false)
		createReq["groups"] = groups
	}
	if !data.AuthNetworks.IsNull() {
		var networks []string
		data.AuthNetworks.ElementsAs(ctx, &networks, false)
		createReq["auth_networks"] = networks
	}

	respBody, err := r.client.Post("/iscsi/target", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create iSCSI target, got error: %s", err))
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

	r.readISCSITarget(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSITargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ISCSITargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readISCSITarget(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSITargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ISCSITargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Name.IsNull() {
		updateReq["name"] = data.Name.ValueString()
	}
	if !data.Alias.IsNull() {
		updateReq["alias"] = data.Alias.ValueString()
	}
	if !data.Mode.IsNull() {
		updateReq["mode"] = data.Mode.ValueString()
	}
	if !data.Groups.IsNull() {
		var groups []int64
		data.Groups.ElementsAs(ctx, &groups, false)
		updateReq["groups"] = groups
	}
	if !data.AuthNetworks.IsNull() {
		var networks []string
		data.AuthNetworks.ElementsAs(ctx, &networks, false)
		updateReq["auth_networks"] = networks
	}

	endpoint := fmt.Sprintf("/iscsi/target/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update iSCSI target, got error: %s", err))
		return
	}

	r.readISCSITarget(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ISCSITargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ISCSITargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/iscsi/target/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete iSCSI target, got error: %s", err))
		return
	}
}

func (r *ISCSITargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ISCSITargetResource) readISCSITarget(ctx context.Context, data *ISCSITargetResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/iscsi/target/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read iSCSI target, got error: %s", err))
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
	if alias, ok := result["alias"].(string); ok {
		data.Alias = types.StringValue(alias)
	}
	if mode, ok := result["mode"].(string); ok {
		data.Mode = types.StringValue(mode)
	}
	if groups, ok := result["groups"].([]interface{}); ok {
		groupList := make([]int64, 0, len(groups))
		for _, g := range groups {
			if groupMap, ok := g.(map[string]interface{}); ok {
				if portal, ok := groupMap["portal"].(float64); ok {
					groupList = append(groupList, int64(portal))
				}
			}
		}
		list, _ := types.ListValueFrom(ctx, types.Int64Type, groupList)
		data.Groups = list
	}
	if networks, ok := result["auth_networks"].([]interface{}); ok {
		networkList := make([]string, 0, len(networks))
		for _, n := range networks {
			if str, ok := n.(string); ok {
				networkList = append(networkList, str)
			}
		}
		list, _ := types.ListValueFrom(ctx, types.StringType, networkList)
		data.AuthNetworks = list
	}
}
