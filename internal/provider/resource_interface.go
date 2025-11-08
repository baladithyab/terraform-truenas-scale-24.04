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

var _ resource.Resource = &InterfaceResource{}
var _ resource.ResourceWithImportState = &InterfaceResource{}

func NewInterfaceResource() resource.Resource {
	return &InterfaceResource{}
}

type InterfaceResource struct {
	client *truenas.Client
}

type InterfaceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	IPV4DHCP    types.Bool   `tfsdk:"ipv4_dhcp"`
	IPV6Auto    types.Bool   `tfsdk:"ipv6_auto"`
	Aliases     types.List   `tfsdk:"aliases"`
	MTU         types.Int64  `tfsdk:"mtu"`
	// VLAN specific
	VLANParentInterface types.String `tfsdk:"vlan_parent_interface"`
	VLANTag             types.Int64  `tfsdk:"vlan_tag"`
	VLANPCP             types.Int64  `tfsdk:"vlan_pcp"`
	// Bridge specific
	BridgeMembers types.List `tfsdk:"bridge_members"`
	// LAG specific
	LAGPorts    types.List   `tfsdk:"lag_ports"`
	LAGProtocol types.String `tfsdk:"lag_protocol"`
}

type InterfaceAlias struct {
	Address types.String `tfsdk:"address"`
	Netmask types.Int64  `tfsdk:"netmask"`
}

func (r *InterfaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_interface"
}

func (r *InterfaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a network interface on TrueNAS",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Interface identifier (same as name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Interface name (e.g., eth0, vlan10, br0, bond0)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Interface type (PHYSICAL, VLAN, BRIDGE, LINK_AGGREGATION)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Interface description",
				Optional:            true,
				Computed:            true,
			},
			"ipv4_dhcp": schema.BoolAttribute{
				MarkdownDescription: "Use DHCP for IPv4",
				Optional:            true,
				Computed:            true,
			},
			"ipv6_auto": schema.BoolAttribute{
				MarkdownDescription: "Use auto-configuration for IPv6",
				Optional:            true,
				Computed:            true,
			},
			"aliases": schema.ListNestedAttribute{
				MarkdownDescription: "Static IP addresses",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "IP address",
							Required:            true,
						},
						"netmask": schema.Int64Attribute{
							MarkdownDescription: "Netmask (CIDR notation, e.g., 24 for /24)",
							Required:            true,
						},
					},
				},
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "Maximum Transmission Unit",
				Optional:            true,
				Computed:            true,
			},
			"vlan_parent_interface": schema.StringAttribute{
				MarkdownDescription: "Parent interface for VLAN (required when type is VLAN)",
				Optional:            true,
			},
			"vlan_tag": schema.Int64Attribute{
				MarkdownDescription: "VLAN tag (required when type is VLAN)",
				Optional:            true,
			},
			"vlan_pcp": schema.Int64Attribute{
				MarkdownDescription: "VLAN Priority Code Point (0-7)",
				Optional:            true,
				Computed:            true,
			},
			"bridge_members": schema.ListAttribute{
				MarkdownDescription: "Bridge member interfaces (required when type is BRIDGE)",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"lag_ports": schema.ListAttribute{
				MarkdownDescription: "LAG member ports (required when type is LINK_AGGREGATION)",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"lag_protocol": schema.StringAttribute{
				MarkdownDescription: "LAG protocol (LACP, FAILOVER, LOADBALANCE, ROUNDROBIN, NONE)",
				Optional:            true,
			},
		},
	}
}

func (r *InterfaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InterfaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}
	if !data.IPV4DHCP.IsNull() {
		createReq["ipv4_dhcp"] = data.IPV4DHCP.ValueBool()
	}
	if !data.IPV6Auto.IsNull() {
		createReq["ipv6_auto"] = data.IPV6Auto.ValueBool()
	}
	if !data.MTU.IsNull() {
		createReq["mtu"] = data.MTU.ValueInt64()
	}

	// Handle aliases
	if !data.Aliases.IsNull() {
		var aliasList []InterfaceAlias
		data.Aliases.ElementsAs(ctx, &aliasList, false)

		aliases := make([]map[string]interface{}, 0, len(aliasList))
		for _, a := range aliasList {
			alias := map[string]interface{}{
				"address": a.Address.ValueString(),
				"netmask": a.Netmask.ValueInt64(),
			}
			aliases = append(aliases, alias)
		}
		createReq["aliases"] = aliases
	}

	// VLAN specific
	if !data.VLANParentInterface.IsNull() {
		createReq["vlan_parent_interface"] = data.VLANParentInterface.ValueString()
	}
	if !data.VLANTag.IsNull() {
		createReq["vlan_tag"] = data.VLANTag.ValueInt64()
	}
	if !data.VLANPCP.IsNull() {
		createReq["vlan_pcp"] = data.VLANPCP.ValueInt64()
	}

	// Bridge specific
	if !data.BridgeMembers.IsNull() {
		var members []string
		data.BridgeMembers.ElementsAs(ctx, &members, false)
		createReq["bridge_members"] = members
	}

	// LAG specific
	if !data.LAGPorts.IsNull() {
		var ports []string
		data.LAGPorts.ElementsAs(ctx, &ports, false)
		createReq["lag_ports"] = ports
	}
	if !data.LAGProtocol.IsNull() {
		createReq["lag_protocol"] = data.LAGProtocol.ValueString()
	}

	respBody, err := r.client.Post("/interface", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create interface, got error: %s", err))
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

	r.readInterface(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InterfaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readInterface(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InterfaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := map[string]interface{}{}

	if !data.Description.IsNull() {
		updateReq["description"] = data.Description.ValueString()
	}
	if !data.IPV4DHCP.IsNull() {
		updateReq["ipv4_dhcp"] = data.IPV4DHCP.ValueBool()
	}
	if !data.IPV6Auto.IsNull() {
		updateReq["ipv6_auto"] = data.IPV6Auto.ValueBool()
	}

	endpoint := fmt.Sprintf("/interface/id/%s", data.ID.ValueString())
	_, err := r.client.Put(endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update interface, got error: %s", err))
		return
	}

	r.readInterface(ctx, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InterfaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/interface/id/%s", data.ID.ValueString())
	_, err := r.client.Delete(endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete interface, got error: %s", err))
		return
	}
}

func (r *InterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *InterfaceResource) readInterface(ctx context.Context, data *InterfaceResourceModel, diags *diag.Diagnostics) {
	endpoint := fmt.Sprintf("/interface/id/%s", data.ID.ValueString())
	respBody, err := r.client.Get(endpoint)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read interface, got error: %s", err))
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		diags.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	if name, ok := result["name"].(string); ok {
		data.Name = types.StringValue(name)
		data.ID = types.StringValue(name)
	}
	if ifType, ok := result["type"].(string); ok {
		data.Type = types.StringValue(ifType)
	}
	if description, ok := result["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if ipv4dhcp, ok := result["ipv4_dhcp"].(bool); ok {
		data.IPV4DHCP = types.BoolValue(ipv4dhcp)
	}
	if ipv6auto, ok := result["ipv6_auto"].(bool); ok {
		data.IPV6Auto = types.BoolValue(ipv6auto)
	}
	if mtu, ok := result["mtu"].(float64); ok {
		data.MTU = types.Int64Value(int64(mtu))
	}
}
